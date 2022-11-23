/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package odp //
package odp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type EventAPIManagerTestSuite struct {
	suite.Suite
	eventAPIManager *DefaultEventAPIManager
	apiHost, apiKey string
	events          []Event
}

func (e *EventAPIManagerTestSuite) SetupTest() {
	e.eventAPIManager = NewEventAPIManager("", nil)
	e.apiHost = "test-host"
	e.apiKey = "test-api-key"
	e.events = []Event{
		{
			Type:        "t1",
			Action:      "a1",
			Identifiers: map[string]string{"id-key-1": "id-value-1"},
			Data: map[string]interface{}{
				"key11": "value-1",
				"key12": true,
				"key13": 3.5,
				"key14": nil,
			},
		},
		{
			Type:   "t1",
			Action: "a1",
		},
	}
}

func (e *EventAPIManagerTestSuite) TestShouldSendEventsSuccessfullyAndNotSuggestRetry() {
	ts := e.getTestServer(202, 0)
	defer ts.Close()
	config := NewConfig(e.apiKey, ts.URL, nil)
	canRetry, err := e.eventAPIManager.SendODPEvents(config, e.events)
	e.NoError(err)
	e.False(canRetry)
}

func (e *EventAPIManagerTestSuite) TestShouldNotSuggestRetryFor400HttpResponse() {
	ts := e.getTestServer(400, 0)
	defer ts.Close()
	config := NewConfig(e.apiKey, ts.URL, nil)
	canRetry, err := e.eventAPIManager.SendODPEvents(config, e.events)
	e.Equal(fmt.Errorf(odpEventFailed, "400 Bad Request"), err)
	e.False(canRetry)
}

func (e *EventAPIManagerTestSuite) TestShouldNotSuggestRetryForInvalidURL() {
	config := NewConfig("123", "456", nil)
	canRetry, err := e.eventAPIManager.SendODPEvents(config, e.events)
	e.Error(err)
	e.False(canRetry)
}

func (e *EventAPIManagerTestSuite) TestShouldSuggestRetryFor500HttpResponse() {
	ts := e.getTestServer(500, 0)
	defer ts.Close()
	config := NewConfig(e.apiKey, ts.URL, nil)
	canRetry, err := e.eventAPIManager.SendODPEvents(config, e.events)
	e.Equal(fmt.Errorf(odpEventFailed, "500 Internal Server Error"), err)
	e.True(canRetry)
}

func (e *EventAPIManagerTestSuite) TestSuggestRetryForNetworkTimeout() {
	ts := e.getTestServer(202, 100)
	defer ts.Close()
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = 10 * time.Millisecond
	config := NewConfig(e.apiKey, ts.URL, nil)
	canRetry, err := e.eventAPIManager.SendODPEvents(config, e.events)
	e.Error(err)
	e.True(canRetry)
}

// func (e *EventAPIManagerTestSuite) TestLiveEvent() {
// 	config := NewConfig("W4WzcEs-ABgXorzY7h1LCQ", "https://api.zaius.com", nil)
// 	identifiers := map[string]string{ODPFSUserIDKey: "abc"}
// 	events := []Event{{
// 		Type:        ODPEventType,
// 		Action:      ODPActionIdentified,
// 		Identifiers: identifiers,
// 		Data: map[string]interface{}{
// 			"idempotence_id":      "xyz",
// 			"source":              "go-sdk",
// 			"data_source_type":    "sdk",
// 			"data_source_version": "1.8.3",
// 		},
// 	}}
// 	canRetry, err := e.eventAPIManager.SendODPEvents(config, events)
// 	e.NoError(err)
// 	e.False(canRetry)
// }

func (e *EventAPIManagerTestSuite) getTestServer(statusCode, timeout int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == eventsAPIEndpointPath {
			e.Equal("POST", r.Method)
			e.Equal(utils.ContentTypeJSON, r.Header.Get(utils.HeaderContentType))
			e.Equal(e.apiKey, r.Header.Get(ODPAPIKeyHeader))
			var requestData []Event
			e.NoError(json.NewDecoder(r.Body).Decode(&requestData))
			reflect.DeepEqual(e.events, requestData)
			if timeout > 0 {
				time.Sleep(time.Duration(timeout) * time.Millisecond)
			}
			w.WriteHeader(statusCode)
			return
		}
		e.Fail("invalid url string")
	}))
}

func TestEventAPIManagerTestSuite(t *testing.T) {
	suite.Run(t, new(EventAPIManagerTestSuite))
}
