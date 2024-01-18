/****************************************************************************
 * Copyright 2022-2023, Optimizely, Inc. and contributors                   *
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

// Package segment //
package segment

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	pkgUtils "github.com/optimizely/go-sdk/v2/pkg/utils"
)

const graphqlAPIEndpointPath = "/v3/graphql"

// APIManager represents the segment API manager.
type APIManager interface {
	// not passing ODPConfig here to avoid multiple mutex lock calls inside async requests
	FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string) ([]string, error)
}

// ODP GraphQL API
// - https://api.zaius.com/v3/graphql
// - test ODP public API key = "W4WzcEs-ABgXorzY7h1LCQ"
/*

 [GraphQL Request]

 // fetch info with fs_user_id for ["has_email", "has_email_opted_in", "push_on_sale"] segments
 curl --location --request POST 'https://api.zaius.com/v3/graphql' \
--header 'Content-Type: application/json' \
--header 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' \
--data-raw '{
  "query":  "query($userId: String, $audiences: [String]) {customer(fs_user_id: $userId){audiences(subset: $audiences) {edges {node {name state}}}}}",
  "variables": {
    "userId": "tester-101",
    "audiences": ["has_email", "has_email_opted_in", "push_on_sale"]
  }
}'

 // fetch info with vuid for ["has_email", "has_email_opted_in", "push_on_sale"] segments
 curl --location --request POST 'https://api.zaius.com/v3/graphql' \
--header 'Content-Type: application/json' \
--header 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' \
--data-raw '{
  "query":  "query($userId: String, $audiences: [String]) {customer(vuid: $userId){audiences(subset: $audiences) {edges {node {name state}}}}}",
  "variables": {
    "userId": "d66a9d81923d4d2f99d8f64338976322",
    "audiences": ["has_email", "has_email_opted_in", "push_on_sale"]
  }
}'

 [GraphQL Response]
 {
   "data": {
     "customer": {
       "audiences": {
         "edges": [
           {
             "node": {
               "name": "has_email",
               "state": "qualified",
             }
           },
           {
             "node": {
               "name": "has_email_opted_in",
               "state": "qualified",
             }
           },
            ...
         ]
       }
     }
   }
 }

 [GraphQL Error Response]
 {
   "errors": [
     {
       "message": "Exception while fetching data (/customer) : java.lang.RuntimeException: could not resolve _fs_user_id = asdsdaddddd",
       "locations": [
         {
           "line": 2,
           "column": 3
         }
       ],
       "path": [
         "customer"
       ],
       "extensions": {
         "code": "INVALID_IDENTIFIER_EXCEPTION",
         "classification": "DataFetchingException"
       }
     }
   ],
   "data": {
     "customer": null
   }
 }
*/

// DefaultSegmentAPIManager represents default implementation of Segment API Manager
type DefaultSegmentAPIManager struct {
	requester pkgUtils.Requester
}

// NewSegmentAPIManager creates and returns a new instance of DefaultSegmentAPIManager.
func NewSegmentAPIManager(sdkKey string, requester pkgUtils.Requester) *DefaultSegmentAPIManager {
	if requester == nil {
		requester = pkgUtils.NewHTTPRequester(logging.GetLogger(sdkKey, "SegmentAPIManager"), pkgUtils.Timeout(utils.DefaultSegmentFetchTimeout))
	}
	return &DefaultSegmentAPIManager{requester: requester}
}

// FetchQualifiedSegments returns qualified ODP segments
func (sm *DefaultSegmentAPIManager) FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string) ([]string, error) {

	// Creating query for odp request
	requestQuery := sm.createRequestQuery(userID, segmentsToCheck)

	// Creating request
	apiEndpoint, err := url.ParseRequestURI(fmt.Sprintf("%s%s", apiHost, graphqlAPIEndpointPath))
	if err != nil {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, err.Error())
	}
	headers := []pkgUtils.Header{{Name: pkgUtils.HeaderContentType, Value: pkgUtils.ContentTypeJSON}, {Name: utils.OdpAPIKeyHeader, Value: apiKey}}

	// handling edge cases
	response, _, _, err := sm.requester.Post(apiEndpoint.String(), requestQuery, headers...)
	if err != nil {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, err.Error())
	}

	// Checking if response is decodable
	responseMap := map[string]interface{}{}
	if err = json.Unmarshal(response, &responseMap); err != nil {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "decode error")
	}

	// most meaningful ODP errors are returned in 200 success JSON under {"errors": ...}
	if odpErrors, ok := sm.extractComponent("errors", responseMap).([]interface{}); ok {
		if odpError, ok := odpErrors[0].(map[string]interface{}); ok {
			if errorClass, ok := sm.extractComponent("extensions.classification", odpError).(string); ok {
				if errorCode, ok := sm.extractComponent("extensions.code", odpError).(string); ok && errorCode == "INVALID_IDENTIFIER_EXCEPTION" {
					return nil, errors.New(utils.InvalidSegmentIdentifier)
				}
				return nil, fmt.Errorf(utils.FetchSegmentsFailedError, errorClass)
			}
		}
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "decode error")
	}

	// Retrieving audience edges from response
	audienceDictionaries, ok := sm.extractComponent("data.customer.audiences.edges", responseMap).([]interface{})
	if !ok {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "decode error")
	}

	// Parsing and returning qualified segments
	returnSegments := []string{}
	for _, audDict := range audienceDictionaries {
		convertedAudDict, ok := audDict.(map[string]interface{})
		if !ok {
			continue
		}
		if jsonbody, err := json.Marshal(convertedAudDict["node"]); err == nil {
			var audience Audience
			if err := json.Unmarshal(jsonbody, &audience); err != nil {
				continue
			}
			if audience.isQualified() {
				returnSegments = append(returnSegments, audience.Name)
			}
		}
	}
	return returnSegments, nil
}

// Creates graphql query
func (sm DefaultSegmentAPIManager) createRequestQuery(userID string, segmentsToCheck []string) map[string]interface{} {
	query := fmt.Sprintf(
		`query($userId: String, $audiences: [String]) {customer(%s: $userId) {audiences(subset: $audiences) {edges {node {name state}}}}}`,
		utils.OdpFSUserIDKey)
	requestQuery := map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"userId":    userID,
			"audiences": segmentsToCheck,
		},
	}
	return requestQuery
}

// Extract deep-json contents with keypath "a.b.c"
// { "a": { "b": { "c": "contents" } } }
func (sm DefaultSegmentAPIManager) extractComponent(keyPath string, dict map[string]interface{}) interface{} {
	var current interface{} = dict
	paths := strings.Split(keyPath, ".")
	for _, path := range paths {
		v, ok := current.(map[string]interface{})
		if ok {
			current = v[path]
			continue
		}
		current = nil
		break
	}
	return current
}
