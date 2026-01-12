/****************************************************************************
 * Copyright 2019,2021-2023 Optimizely, Inc. and contributors               *
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

package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/logging"

	"github.com/stretchr/testify/assert"
)

func TestClientFunction(t *testing.T) {
	requester := &HTTPRequester{}
	fn := Client(http.Client{Timeout: 125})
	fn(requester)
	assert.Equal(t, time.Duration(125), requester.client.Timeout)
}

func TestNewHTTPRequesterWithClient(t *testing.T) {
	fn := Client(http.Client{Timeout: 125})
	requester := NewHTTPRequester(logging.GetLogger("", ""), fn)
	assert.Equal(t, time.Duration(125), requester.client.Timeout)
}

func TestNewHTTPRequesterWithoutClient(t *testing.T) {
	requester := NewHTTPRequester(logging.GetLogger("", ""))
	assert.Equal(t, defaultTTL, requester.client.Timeout)
}

func TestHeaders(t *testing.T) {
	requester := &HTTPRequester{}
	fn := Headers(Header{"one", "1"})
	fn(requester)
	assert.Equal(t, []Header{{"one", "1"}}, requester.headers)
}

func TestAddHeaders(t *testing.T) {

	req, _ := http.NewRequest("GET", "", nil)
	requester := &HTTPRequester{logger: logging.GetLogger("", "")}
	headers := []Header{{"one", "1"}}

	fn := Headers(Header{"two", "2"})
	fn(requester)
	requester.addHeaders(req, headers)
	assert.Equal(t, req.Header, http.Header{"One": []string{"1"}, "Two": []string{"2"}})
}

func TestGet(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(">request: ", r)
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, "Hello, client")
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	var httpreq Requester
	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	resp, headers, code, err := httpreq.Get(ts.URL + "/good")
	assert.NotEqual(t, headers.Get(HeaderContentType), "")
	assert.Nil(t, err)
	assert.Equal(t, "Hello, client\n", string(resp))

	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	_, headers, code, err = httpreq.Get(ts.URL + "/bad")
	assert.Equal(t, errors.New("400 Bad Request"), err)
	assert.Equal(t, code, http.StatusBadRequest)
}

func TestGetObj(t *testing.T) {

	type resp struct {
		Fld1 string
		Fld2 int
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(">request: ", r)
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, `{"fld1":"Hello, client", "fld2": 123}`)
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, `bad bad response`)
		}
	}))
	defer ts.Close()

	var httpreq Requester
	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	r := resp{}
	err := httpreq.GetObj(ts.URL+"/good", &r)
	assert.Nil(t, err)
	assert.Equal(t, resp{Fld1: "Hello, client", Fld2: 123}, r)

	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	err = httpreq.GetObj(ts.URL+"/bad", &r)
	assert.NotNil(t, err)
}

func TestPost(t *testing.T) {

	type body struct {
		Fld1 string
		Fld2 int
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(">request: ", r)
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, "Hello, client")
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()
	b := body{"one", 1}
	var httpreq Requester
	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	resp, headers, code, err := httpreq.Post(ts.URL+"/good", b)

	assert.Nil(t, err)
	assert.NotEqual(t, headers.Get(HeaderContentType), "")
	assert.Equal(t, "Hello, client\n", string(resp))
	assert.Equal(t, code, http.StatusOK)

	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	_, _, code, err = httpreq.Post(ts.URL+"/bad", nil)
	assert.Equal(t, errors.New("400 Bad Request"), err)
	assert.Equal(t, code, http.StatusBadRequest)
}

func TestPostObj(t *testing.T) {

	type body struct {
		Fld1 string
		Fld2 int
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(">request: ", r)
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, `{"fld1":"Hello, client", "fld2": 123}`)
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, `bad bad response`)
		}
	}))
	defer ts.Close()

	var httpreq Requester
	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	b := body{"one", 1}
	r := body{}
	err := httpreq.PostObj(ts.URL+"/good", b, &r)
	assert.Nil(t, err)
	assert.Equal(t, body{Fld1: "Hello, client", Fld2: 123}, r)

	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	err = httpreq.PostObj(ts.URL+"/bad", b, &r)
	assert.NotNil(t, err)
}

type mockLogger struct {
	Errors []error
}

func (m *mockLogger) Debug(message string)   {}
func (m *mockLogger) Info(message string)    {}
func (m *mockLogger) Warning(message string) {}
func (m *mockLogger) Error(message string, err interface{}) {
	if err, ok := err.(error); ok {
		m.Errors = append(m.Errors, err)
	}
}

func TestGetBad(t *testing.T) {
	// Using a mockLogger to ensure we're logging the expected error message
	mLogger := &mockLogger{}
	httpreq := NewHTTPRequester(mLogger)

	badURL := "http://ww.bad-url.fake/blah12345"
	_, _, _, err := httpreq.Get(badURL)
	returnedErr, ok := err.(*url.Error)
	assert.True(t, ok, "url error")

	// Check to make sure we have some log for bad url
	assert.NotNil(t, mLogger.Errors)
	// If we didn't get the expected error, we need to stop before we do the rest
	// of the checks that depend on that error.
	if !assert.Len(t, mLogger.Errors, 1, "logged error") {
		t.FailNow()
	}
	// Check to make sure the error that was logged is the same as what was returned
	loggedErr, ok := mLogger.Errors[0].(*url.Error)
	assert.True(t, ok, "is URL error")
	assert.Equal(t, returnedErr, loggedErr, "expected same error")
	assert.Equal(t, badURL, loggedErr.URL, "expected the URL we requested")
}

func TestGetBadWithResponse(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(">request: ", r)
		if r.URL.String() == "/good" {
			fmt.Fprintln(w, `{"fld1":"Hello, client", "fld2": 123}`)
		}
		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, `bad bad response`)
		}
	}))
	defer ts.Close()

	httpreq := NewHTTPRequester(logging.GetLogger("", ""), Retries(1))
	data, _, _, err := httpreq.Get(ts.URL + "/bad")
	assert.Equal(t, "400 Bad Request", err.Error())
	assert.Equal(t, "bad bad response\n", string(data))
}

func TestGetRetry(t *testing.T) {
	called := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Print("request: ", r)

		if r.URL.String() == "/test" {
			called++
			if called >= 5 {
				fmt.Fprintln(w, "Hello, client")
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}

		if r.URL.String() == "/bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
		if r.URL.String() == "/good" {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	httpreq := NewHTTPRequester(logging.GetLogger("", ""), Retries(10))

	st := time.Now()
	resp, _, _, err := httpreq.Get(ts.URL + "/test")
	assert.Nil(t, err)
	assert.Equal(t, "Hello, client\n", string(resp))
	assert.Equal(t, 5, called, "called 5 retries")
	elapsed := time.Since(st)

	// With exponential backoff: 200ms + 400ms + 800ms + 1000ms = 2400ms for 4 delays (5 attempts)
	// Allow some tolerance for test execution overhead
	assert.True(t, elapsed >= 2*time.Second && elapsed <= 10*time.Second, "took %s", elapsed)

	httpreq = NewHTTPRequester(logging.GetLogger("", ""), Retries(3))
	called = 0
	_, _, _, err = httpreq.Get(ts.URL + "/test")
	assert.Equal(t, errors.New("400 Bad Request"), err)
	assert.Equal(t, 3, called, "called 3 retries")

	httpreq = NewHTTPRequester(logging.GetLogger("", ""), Retries(1))
	called = 0
	_, _, _, err = httpreq.Get(ts.URL + "/test")
	assert.Equal(t, errors.New("400 Bad Request"), err)
	assert.Equal(t, 1, called, "called 1 retries")

	httpreq = NewHTTPRequester(logging.GetLogger("", ""))
	called = 0
	_, _, _, err = httpreq.Get(ts.URL + "/test")
	assert.Equal(t, errors.New("400 Bad Request"), err)
	assert.Equal(t, 1, called, "called 1 retries")
}

func TestGetRetryExponentialBackoff(t *testing.T) {
	callTimes := []time.Time{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callTimes = append(callTimes, time.Now())
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	// Use 4 retries so we get 3 delay intervals to measure
	httpreq := NewHTTPRequester(logging.GetLogger("", ""), Retries(4))

	_, _, _, _ = httpreq.Get(ts.URL + "/test")
	assert.Equal(t, 4, len(callTimes), "should have made 4 attempts")

	// Verify exponential backoff pattern in intervals
	// Expected delays: ~200ms, ~400ms, ~800ms (with tolerance)
	if len(callTimes) >= 4 {
		interval1 := callTimes[1].Sub(callTimes[0])
		interval2 := callTimes[2].Sub(callTimes[1])
		interval3 := callTimes[3].Sub(callTimes[2])

		// First interval should be ~200ms (allow 150-350ms for tolerance)
		assert.True(t, interval1 >= 150*time.Millisecond && interval1 <= 350*time.Millisecond,
			"first interval should be ~200ms, got %v", interval1)

		// Second interval should be ~400ms (allow 300-600ms for tolerance)
		assert.True(t, interval2 >= 300*time.Millisecond && interval2 <= 600*time.Millisecond,
			"second interval should be ~400ms, got %v", interval2)

		// Third interval should be ~800ms (allow 600-1100ms for tolerance)
		assert.True(t, interval3 >= 600*time.Millisecond && interval3 <= 1100*time.Millisecond,
			"third interval should be ~800ms, got %v", interval3)

		// Verify exponential growth: each interval should be roughly double the previous
		assert.True(t, interval2 > interval1, "second interval should be greater than first")
		assert.True(t, interval3 > interval2, "third interval should be greater than second")
	}
}

func TestString(t *testing.T) {
	assert.Equal(t, "{timeout: 5s, retries: 1}", NewHTTPRequester(logging.GetLogger("", "")).String())
	assert.Equal(t, "{timeout: 19s, retries: 10}",
		NewHTTPRequester(logging.GetLogger("", ""), Retries(10), Timeout(time.Duration(19)*time.Second)).String())

}
