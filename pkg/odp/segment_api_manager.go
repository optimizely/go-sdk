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
	"errors"
	"fmt"
	"strings"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
)

// ODP GraphQL API
// - https://api.zaius.com/v3/graphql
// - test ODP public API key = "W4WzcEs-ABgXorzY7h1LCQ"
/*

 [GraphQL Request]

 // fetch info with fs_user_id for ["has_email", "has_email_opted_in", "push_on_sale"] segments
 curl -i -H 'Content-Type: application/json' -H 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' -X POST -d '{"query":"query {customer(fs_user_id: \"tester-101\") {audiences(subset:[\"has_email\",\"has_email_opted_in\",\"push_on_sale\"]) {edges {node {name state}}}}}"}' https://api.zaius.com/v3/graphql
 // fetch info with vuid for ["has_email", "has_email_opted_in", "push_on_sale"] segments
 curl -i -H 'Content-Type: application/json' -H 'x-api-key: W4WzcEs-ABgXorzY7h1LCQ' -X POST -d '{"query":"query {customer(vuid: \"d66a9d81923d4d2f99d8f64338976322\") {audiences(subset:[\"has_email\",\"has_email_opted_in\",\"push_on_sale\"]) {edges {node {name state}}}}}"}' https://api.zaius.com/v3/graphql
 query MyQuery {
   customer(vuid: "d66a9d81923d4d2f99d8f64338976322") {
     audiences(subset:["has_email","has_email_opted_in","push_on_sale"]) {
       edges {
         node {
           name
           state
         }
       }
     }
   }
 }
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
         "classification": "InvalidIdentifierException"
       }
     }
   ],
   "data": {
     "customer": null
   }
 }
*/

// Audience represents an ODP Audience
type Audience struct {
	Name        string `json:"name"`
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}

func (s Audience) isQualified() bool {
	return s.State == "qualified"
}

// SegmentAPIManager is used to fetch qualified ODP segments
type SegmentAPIManager struct {
	requester utils.Requester
}

// NewSegmentAPIManager creates and returns a new instance SegmentAPIManager.
func NewSegmentAPIManager(sdkKey string, requester utils.Requester, logger logging.OptimizelyLogProducer) *SegmentAPIManager {
	if requester == nil {
		if logger == nil {
			logger = logging.GetLogger(sdkKey, "SegmentAPIManager")
		}
		requester = utils.NewHTTPRequester(logger)
	}
	return &SegmentAPIManager{requester: requester}
}

// FetchSegments returns qualified ODP segments
func (s *SegmentAPIManager) FetchSegments(apiKey, apiHost, userKey, userValue string, segmentsToCheck []string, handler func([]string, error)) {

	// Creating query for odp request
	query := s.createRequestQuery(userKey, userValue, segmentsToCheck)

	// Creating request
	apiEndpoint := apiHost + "/v3/graphql"
	headers := []utils.Header{{Name: "Content-Type", Value: "application/json"}, {Name: "x-api-key", Value: apiKey}}

	// handling edge cases
	response, _, statusCode, err := s.requester.Post(apiEndpoint, map[string]string{"query": query}, headers...)
	if response == nil {
		if err != nil {
			handler(nil, fmt.Errorf(fetchSegmentsFailedError, "invalid response"))
			return
		}
	}
	if err != nil {
		handler(nil, fmt.Errorf(fetchSegmentsFailedError, err.Error()))
		return
	}
	if statusCode >= 400 {
		handler(nil, fmt.Errorf(fetchSegmentsFailedError, fmt.Sprintf("%d", statusCode)))
		return
	}

	// Checking if response is decodable
	responseMap := map[string]interface{}{}
	if err = json.Unmarshal(response, &responseMap); err != nil {
		handler(nil, fmt.Errorf(fetchSegmentsFailedError, "decode error"))
		return
	}

	// most meaningful ODP errors are returned in 200 success JSON under {"errors": ...}
	if odpErrors, ok := s.extractComponent("errors", responseMap).([]interface{}); ok {
		if odpError, ok := odpErrors[0].(map[string]interface{}); ok {
			if errorClass, ok := s.extractComponent("extensions.classification", odpError).(string); ok {
				if errorClass == "InvalidIdentifierException" {
					handler(nil, errors.New(invalidSegmentIdentifier))
					return
				}
				handler(nil, fmt.Errorf(fetchSegmentsFailedError, errorClass))
				return
			}
		}
		handler(nil, fmt.Errorf(fetchSegmentsFailedError, "decode error"))
		return
	}

	// Retrieving audience edges from response
	audienceDictionaries, ok := s.extractComponent("data.customer.audiences.edges", responseMap).([]interface{})
	if !ok {
		handler(nil, fmt.Errorf(fetchSegmentsFailedError, "decode error"))
		return
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
	handler(returnSegments, nil)
}

// Creates graphql query
func (s SegmentAPIManager) createRequestQuery(userKey, userValue string, segmentsToCheck []string) string {
	return fmt.Sprintf(`query %s`, s.getCustomerFilter(userKey, userValue, segmentsToCheck))
}

// Creates filter for customer
func (s SegmentAPIManager) getCustomerFilter(userKey, userValue string, segmentsToCheck []string) string {
	return fmt.Sprintf(`{customer(%s: %q) %s}`, userKey, userValue, s.getAudiencesFilter(segmentsToCheck))
}

// Creates filter for audiences
func (s SegmentAPIManager) getAudiencesFilter(segmentsToCheck []string) string {
	subsetFilter := s.makeSubsetFilter(segmentsToCheck)
	return fmt.Sprintf(`{audiences%s %s}`, subsetFilter, s.getEdgesFilter())
}

// Creates filter for edges
func (s SegmentAPIManager) getEdgesFilter() string {
	return fmt.Sprintf(`{edges %s}`, s.getNodesFilter())
}

// Creates filter for nodes
func (s SegmentAPIManager) getNodesFilter() string {
	return `{node {name state}}`
}

// Creates filter for subset
func (s SegmentAPIManager) makeSubsetFilter(segmentsToCheck []string) string {
	// segments = []: (fetch none)
	//   --> subsetFilter = "(subset:[])"
	// segments = ["a"]: (fetch one segment)
	//   --> subsetFilter = "(subset:[\"a\"])"

	escapedSegments := []string{}
	for _, v := range segmentsToCheck {
		escapedSegments = append(escapedSegments, fmt.Sprintf(`%q`, v))
	}
	return fmt.Sprintf("(subset:[%s])", strings.Join(escapedSegments, ","))
}

// Extract deep-json contents with keypath "a.b.c"
// { "a": { "b": { "c": "contents" } } }
func (s SegmentAPIManager) extractComponent(keyPath string, dict map[string]interface{}) interface{} {
	var current interface{} = dict
	paths := strings.Split(keyPath, ".")
	for _, path := range paths {
		if v, ok := current.(map[string]interface{})[path]; ok {
			current = v
			continue
		}
		current = nil
		break
	}
	return current
}
