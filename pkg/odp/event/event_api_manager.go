/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package event //
package event

import (
	"fmt"
	"net/url"

	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	pkgUtils "github.com/optimizely/go-sdk/v2/pkg/utils"
)

// APIManager represents the event API manager.
type APIManager interface {
	// not passing ODPConfig here to avoid multiple mutex lock calls inside the batch events loop
	SendOdpEvents(apiKey, apiHost string, events []Event) (canRetry bool, err error)
}

// ODP REST Events API
// - https://api.zaius.com/v3/events
// - test ODP public API key = "W4WzcEs-ABgXorzY7h1LCQ"
/*
 [Event Request]
 curl -i -H 'Content-Type: application/json' -H 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' -X POST -d '{"type":"fullstack","action":"identified","identifiers":{"fs_user_id": "abc"},"data":{"idempotence_id":"xyz","source":"go-sdk","data_source_type":"sdk","data_source_version":"2.0.0-beta"}}' https://api.zaius.com/v3/events
 [Event Response]
 {"title":"Accepted","status":202,"timestamp":"2022-06-30T20:59:52.046Z"}
*/

// DefaultEventAPIManager represents default implementation of Event API Manager
type DefaultEventAPIManager struct {
	requester pkgUtils.Requester
}

// NewEventAPIManager creates and returns a new instance of DefaultEventAPIManager.
func NewEventAPIManager(sdkKey string, requester pkgUtils.Requester) *DefaultEventAPIManager {
	if requester == nil {
		requester = pkgUtils.NewHTTPRequester(logging.GetLogger(sdkKey, "EventAPIManager"), pkgUtils.Timeout(utils.DefaultOdpEventTimeout))
	}
	return &DefaultEventAPIManager{requester: requester}
}

// SendOdpEvents sends events to ODP's RESTful API
func (s *DefaultEventAPIManager) SendOdpEvents(apiKey, apiHost string, events []Event) (canRetry bool, err error) {

	// Creating request
	apiEndpoint, err := url.ParseRequestURI(fmt.Sprintf("%s%s", apiHost, utils.ODPEventsAPIEndpointPath))
	if err != nil {
		return false, fmt.Errorf(utils.OdpEventFailed, err.Error())
	}
	headers := []pkgUtils.Header{{Name: pkgUtils.HeaderContentType, Value: pkgUtils.ContentTypeJSON}, {Name: utils.OdpAPIKeyHeader, Value: apiKey}}

	_, _, status, err := s.requester.Post(apiEndpoint.String(), events, headers...)
	// handling edge cases
	if err == nil {
		return false, nil
	}
	if status >= 400 && status < 500 { // no retry (client error)
		return false, fmt.Errorf(utils.OdpEventFailed, err.Error())
	}
	return true, fmt.Errorf(utils.OdpEventFailed, err.Error())
}
