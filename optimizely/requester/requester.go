package requester

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

const DefaultTTL = 5 * time.Second

var requesterLogger = logging.GetLogger("Requester")

// Header element to be sent
type Header struct {
	Name, Value string
}

// Timeout sets http client timeout
func Timeout(timeout time.Duration) func(r *Requester) {
	return func(r *Requester) {
		r.client = http.Client{Timeout: timeout}
	}
}

// Retries sets max number of retries for failed calls
func Retries(retries int) func(r *Requester) {
	return func(r *Requester) {
		r.retries = retries
	}
}

// Headers sets request headers
func Headers(headers ...Header) func(r *Requester) {
	return func(r *Requester) {
		r.headers = []Header{}
		r.headers = append(r.headers, headers...)
	}
}

// Requester contains main info
type Requester struct {
	url     string
	client  http.Client
	retries int
	headers []Header
	ttl     time.Duration // time-to-live
}

// New makes Requester with api and parameters. Sets defaults
// url has base part of request's url, like https://cdn.optimizely.com/datafiles/
func New(url string, params ...func(*Requester)) *Requester {

	res := Requester{
		url:     url,
		retries: 1,
		headers: []Header{{"Content-Type", "application/json"}, {"Accept", "application/json"}},
		client:  http.Client{Timeout: DefaultTTL},
	}

	for _, param := range params {
		param(&res)
	}
	return &res
}

// Get executes HTTP GET with uri and optional extra headers, returns body in []bytes
// url created as url+sdkKey.json
func (r Requester) Get(uri string, headers ...Header) (response []byte, code int, err error) {
	return r.Do(uri, "GET", nil, headers)
}

// GetObj executes HTTP GET with uri and optional extra headers, returns filled object
func (r Requester) GetObj(uri string, result interface{}, headers ...Header) error {
	b, _, err := r.Do(uri, "GET", nil, headers)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, result)
}

// Do executes request and returns response body for requested uri (sdkKey.json).
func (r Requester) Do(uri string, method string, body io.Reader, headers []Header) (response []byte, code int, err error) {

	single := func(request *http.Request) (response []byte, code int, e error) {
		resp, doErr := r.client.Do(request)
		if doErr != nil {
			requesterLogger.Error(fmt.Sprintf("failed to send request %v", request), e)
			return nil, 0, doErr
		}
		defer func() {
			if e := resp.Body.Close(); e != nil {
				requesterLogger.Warning(fmt.Sprintf("can't close body for %s request, %s", request.URL, e))
			}
		}()

		if response, err = ioutil.ReadAll(resp.Body); err != nil {
			requesterLogger.Error("failed to read body", err)
			return nil, resp.StatusCode, err
		}

		if resp.StatusCode >= 400 {
			requesterLogger.Warning(fmt.Sprintf("error status code=%d", resp.StatusCode))
			return response, resp.StatusCode, errors.New(resp.Status)
		}

		return response, resp.StatusCode, nil
	}

	reqURL := fmt.Sprintf("%s%s", r.url, uri)
	requesterLogger.Debug(fmt.Sprintf("request %s", reqURL))
	req, err := http.NewRequest(method, reqURL, body)
	log.Print(req)
	if err != nil {
		requesterLogger.Error(fmt.Sprintf("failed to make request %s", reqURL), err)
		return nil, 0, err
	}

	r.addHeaders(req, headers)

	for i := 0; i < r.retries; i++ {

		if response, code, err = single(req); err == nil {
			triedMsg := ""
			if i > 0 {
				triedMsg = fmt.Sprintf(", tried %d time(s)", i+1)
			}
			requesterLogger.Debug(fmt.Sprintf("completed %s%s", reqURL, triedMsg))
			return response, code, err
		}
		requesterLogger.Debug(fmt.Sprintf("failed %s with %v", reqURL, err))

		if i != r.retries {
			delay := time.Duration(500) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return response, code, err
}

func (r Requester) addHeaders(req *http.Request, headers []Header) *http.Request {
	for _, h := range r.headers {
		req.Header.Add(h.Name, h.Value)
	}
	for _, h := range headers {
		req.Header.Add(h.Name, h.Value)
	}
	return req
}

func (r Requester) String() string {
	return fmt.Sprintf("{url: %s, timeout: %v, retries: %d}",
		r.url, r.client.Timeout, r.retries)
}
