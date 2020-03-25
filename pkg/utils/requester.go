/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package utils //
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/optimizely/go-sdk/pkg/logging"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const defaultTTL = 5 * time.Second

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Requester is used to make outbound requests with
type Requester interface {
	Get(url string, headers ...Header) (response []byte, responseHeaders http.Header, code int, err error)
	GetObj(url string, result interface{}, headers ...Header) error

	Post(url string, body interface{}, headers ...Header) (response []byte, responseHeaders http.Header, code int, err error)
	PostObj(url string, body interface{}, result interface{}, headers ...Header) error

	String() string
}

// Header element to be sent
type Header struct {
	Name, Value string
}

// Timeout sets http client timeout
func Timeout(timeout time.Duration) func(r *HTTPRequester) {
	return func(r *HTTPRequester) {
		r.client = http.Client{Timeout: timeout}
	}
}

// Retries sets max number of retries for failed calls
func Retries(retries int) func(r *HTTPRequester) {
	return func(r *HTTPRequester) {
		r.retries = retries
	}
}

// Headers sets request headers
func Headers(headers ...Header) func(r *HTTPRequester) {
	return func(r *HTTPRequester) {
		r.headers = []Header{}
		r.headers = append(r.headers, headers...)
	}
}

// HTTPRequester contains main info
type HTTPRequester struct {
	client  http.Client
	retries int
	headers []Header
	logger  logging.OptimizelyLogProducer
}

// NewHTTPRequester makes Requester with api and parameters. Sets defaults
// api has the base part of request's url, like http://localhost/api/v1
func NewHTTPRequester(logger logging.OptimizelyLogProducer, params ...func(*HTTPRequester)) *HTTPRequester {

	res := HTTPRequester{
		retries: 1,
		headers: []Header{{"Content-Type", "application/json"}, {"Accept", "application/json"}},
		client:  http.Client{Timeout: defaultTTL},
		logger:  logger,
	}

	for _, param := range params {
		param(&res)
	}
	return &res
}

// Get executes HTTP GET with url and optional extra headers, returns body in []bytes
func (r HTTPRequester) Get(url string, headers ...Header) (response []byte, responseHeaders http.Header, code int, err error) {
	return r.Do(url, "GET", nil, headers)
}

// GetObj executes HTTP GET with url and optional extra headers, returns filled object
func (r HTTPRequester) GetObj(url string, result interface{}, headers ...Header) error {
	b, _, _, err := r.Do(url, "GET", nil, headers)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, result)
}

// Post executes HTTP POST with url, body and optional extra headers
func (r HTTPRequester) Post(url string, body interface{}, headers ...Header) (response []byte, responseHeaders http.Header, code int, err error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, nil, http.StatusBadRequest, err
	}
	return r.Do(url, "POST", bytes.NewBuffer(b), headers)
}

// PostObj executes HTTP POST with url, body and optional extra headers. Returns filled object
func (r HTTPRequester) PostObj(url string, body, result interface{}, headers ...Header) error {
	b, _, _, err := r.Post(url, body, headers...)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, result)
}

// Do executes request and returns response body for requested url
func (r HTTPRequester) Do(url, method string, body io.Reader, headers []Header) (response []byte, responseHeaders http.Header, code int, err error) {

	single := func(request *http.Request) (response []byte, responseHeaders http.Header, code int, e error) {
		resp, doErr := r.client.Do(request)
		if doErr != nil {
			r.logger.Error(fmt.Sprintf("failed to send request %v", request), e)
			return nil, http.Header{}, 0, doErr
		}
		defer func() {
			if e := resp.Body.Close(); e != nil {
				r.logger.Warning(fmt.Sprintf("can't close body for %s request, %s", request.URL, e))
			}
		}()

		if response, err = ioutil.ReadAll(resp.Body); err != nil {
			r.logger.Error("failed to read body", err)
			return nil, resp.Header, resp.StatusCode, err
		}

		if resp.StatusCode >= http.StatusBadRequest {
			r.logger.Warning(fmt.Sprintf("error status code=%d", resp.StatusCode))
			return response, resp.Header, resp.StatusCode, errors.New(resp.Status)
		}

		return response, resp.Header, resp.StatusCode, nil
	}

	r.logger.Debug(fmt.Sprintf("request %s", url))
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to make request %s", url), err)
		return nil, nil, 0, err
	}

	r.addHeaders(req, headers)

	for i := 0; i < r.retries; i++ {

		if response, responseHeaders, code, err = single(req); err == nil {
			triedMsg := ""
			if i > 0 {
				triedMsg = fmt.Sprintf(", tried %d time(s)", i+1)
			}
			r.logger.Debug(fmt.Sprintf("completed %s%s", url, triedMsg))
			return response, responseHeaders, code, err
		}
		r.logger.Debug(fmt.Sprintf("failed %s with %v", url, err))

		if i != r.retries {
			delay := time.Duration(500) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return response, responseHeaders, code, err
}

func (r HTTPRequester) addHeaders(req *http.Request, headers []Header) *http.Request {
	for _, h := range r.headers {
		req.Header.Add(h.Name, h.Value)
	}
	for _, h := range headers {
		req.Header.Add(h.Name, h.Value)
	}
	return req
}

func (r HTTPRequester) String() string {
	return fmt.Sprintf("{timeout: %v, retries: %d}", r.client.Timeout, r.retries)
}
