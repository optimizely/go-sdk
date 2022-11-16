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
	"fmt"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
)

const eventsAPIEndpointPath = "/v3/events"

// EventAPIManager represents the event API manager.
type EventAPIManager interface {
	SendODPEvents(config Config, events []Event) (canRetry bool, err error)
}

// ODP REST Events API
// - https://api.zaius.com/v3/events
// - test ODP public API key = "W4WzcEs-ABgXorzY7h1LCQ"
/*
 [Event Request]
 curl -i -H 'Content-Type: application/json' -H 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' -X POST -d '{"type":"fullstack","action":"identified","identifiers":{"vuid": "123","fs_user_id": "abc"},"data":{"idempotence_id":"xyz","source":"swift-sdk"}}' https://api.zaius.com/v3/events
 [Event Response]
 {"title":"Accepted","status":202,"timestamp":"2022-06-30T20:59:52.046Z"}
*/

// DefaultEventAPIManager represents default implementation of Event API Manager
type DefaultEventAPIManager struct {
	requester utils.Requester
}

// NewEventAPIManager creates and returns a new instance of DefaultEventAPIManager.
func NewEventAPIManager(sdkKey string, requester utils.Requester) *DefaultEventAPIManager {
	if requester == nil {
		requester = utils.NewHTTPRequester(logging.GetLogger(sdkKey, "EventAPIManager"))
	}
	return &DefaultEventAPIManager{requester: requester}
}

// SendODPEvents sends events to ODP's RESTful API
func (s *DefaultEventAPIManager) SendODPEvents(config Config, events []Event) (canRetry bool, err error) {

	// Creating request
	apiEndpoint := config.GetAPIHost() + eventsAPIEndpointPath
	headers := []utils.Header{{Name: "Content-Type", Value: "application/json"}, {Name: ODPAPIKeyHeader, Value: config.GetAPIKey()}}

	_, _, status, err := s.requester.Post(apiEndpoint, events, headers...)
	// handling edge cases
	if err == nil {
		return false, nil
	}
	if status >= 400 && status < 500 { // no retry (client error)
		return false, fmt.Errorf(odpEventFailed, err.Error())
	}
	return true, fmt.Errorf(odpEventFailed, err.Error())
}
