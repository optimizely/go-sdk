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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type SegmentAPIManagerTestSuite struct {
	suite.Suite
	segmentAPIManager                                                                                    *SegmentAPIManager
	apiHost, apiKey, userValue, userKey                                                                  string
	goodResponseData, goodEmptyResponseData                                                              string
	invalidIdentifierResponseData, invalidErrorResponseData, otherExceptionResponseData, badResponseData string
	invalidEdgeResponseData, invalidNodeResponseData                                                     string
}

func (s *SegmentAPIManagerTestSuite) SetupTest() {
	s.segmentAPIManager = NewSegmentAPIManager(nil)
	s.apiHost = "test-host"
	s.apiKey = "test-api-key"
	s.userValue = "test-user-value"
	s.userKey = "vuid"
	s.goodResponseData = `{
		"data": {
			"customer": {
				"audiences": {
					"edges": [
						{
							"node": {
								"name": "a",
								"state": "qualified",
								"description": "qualifed sample"
							}
						},
						{
							"node": {
								"name": "b",
								"state": "not_qualified",
								"description": "not-qualified sample"
							}
						}
					]
				}
			}
		}
	}`
	s.goodEmptyResponseData = `
	{
        "data": {
            "customer": {
                "audiences": {
                    "edges": []
                }
            }
        }
    }`
	s.invalidIdentifierResponseData = `
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
	  }`
	s.invalidErrorResponseData = `{
		"errors": [
			"Exception while fetching data (/customer) : java.lang.RuntimeException: could not resolve _fs_user_id = asdsdaddddd"
		]
	}`
	s.otherExceptionResponseData = `
	  {
		"errors": [
		  {
			"message": "Exception while fetching data (/customer) : java.lang.RuntimeException: could not resolve _fs_user_id = asdsdaddddd",
			"extensions": {
			  "classification": "TestExceptionClass"
			}
		  }
		],
		"data": {
		  "customer": null
		}
	  }`
	s.badResponseData = `{"data": {}}`
	s.invalidEdgeResponseData = `{
		"data": {
			"customer": {
				"audiences": {
					"edges": [
						[
							"node"
						]
					]
				}
			}
		}
	}`
	s.invalidNodeResponseData = `{
		"data": {
			"customer": {
				"audiences": {
					"edges": [
						{
							"node": [

							]
						}
					]
				}
			}
		}
	}`
}

func (s *SegmentAPIManagerTestSuite) TestSegmentManagerWithRequester() {
	requester := utils.NewHTTPRequester(logging.GetLogger("", ""))
	segmentManager := NewSegmentAPIManager(requester)
	s.Equal(requester, segmentManager.requester)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsSuccess() {
	s.NotNil(s.segmentAPIManager.requester)
	ts := s.getTestServer(0, s.goodResponseData)
	defer ts.Close()
	segmentsToCheck := []string{"a", "b", "c"}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.NoError(err)
	s.Len(segments, 1)
	s.Equal("a", segments[0])
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsSuccessWithEmptySegments() {
	ts := s.getTestServer(0, s.goodEmptyResponseData)
	defer ts.Close()
	segmentsToCheck := []string{"a", "b", "c"}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.NoError(err)
	s.Len(segments, 0)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidIdentifier() {
	ts := s.getTestServer(0, s.invalidIdentifierResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(errors.New(invalidSegmentIdentifier), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidError() {
	ts := s.getTestServer(0, s.invalidErrorResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsOtherException() {
	ts := s.getTestServer(0, s.otherExceptionResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "TestExceptionClass"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsArrayResponse() {
	ts := s.getTestServer(0, `[]`)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidEdgeResponse() {
	ts := s.getTestServer(0, s.invalidEdgeResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidNodeResponse() {
	ts := s.getTestServer(0, s.invalidNodeResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsBadResponse() {
	ts := s.getTestServer(0, s.badResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegments400() {
	ts := s.getTestServer(403, "")
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "403 Forbidden"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegments500() {
	ts := s.getTestServer(500, "")
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchSegments(s.apiKey, ts.URL, s.userKey, s.userValue, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(fetchSegmentsFailedError, "500 Internal Server Error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestExtractComponent() {
	testMap := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{"c": "v"},
		},
	}
	s.True(reflect.DeepEqual(testMap["a"], s.segmentAPIManager.extractComponent("a", testMap)))
	bVal := testMap["a"].(map[string]interface{})["b"].(map[string]interface{})
	s.True(reflect.DeepEqual(bVal, s.segmentAPIManager.extractComponent("a.b", testMap)))
	s.Equal("v", s.segmentAPIManager.extractComponent("a.b.c", testMap))
	s.Nil(s.segmentAPIManager.extractComponent("a.b.c.d", testMap))
	s.Nil(s.segmentAPIManager.extractComponent("d", testMap))
}

func (s *SegmentAPIManagerTestSuite) TestCreateRequestQuery() {
	segmentsToCheck := [][]string{
		{}, {"a", "b"},
	}
	template := "query($userId: String, $audiences: [String]) {customer(key-1: $userId) {audiences(subset: $audiences) {edges {node {name state}}}}}"
	expectedBody := []map[string]interface{}{
		{"query": template, "variables": map[string]interface{}{"audiences": []string{}, "userId": "value-1"}},
		{"query": template, "variables": map[string]interface{}{"audiences": []string{"a", "b"}, "userId": "value-1"}},
	}

	for i := range segmentsToCheck {
		query := s.segmentAPIManager.createRequestQuery("key-1", "value-1", segmentsToCheck[i])
		expected := expectedBody[i]
		s.True(reflect.DeepEqual(expected, query))
	}
}

func (s *SegmentAPIManagerTestSuite) getTestServer(statusCode int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/v3/graphql" {
			s.Equal("POST", r.Method)
			s.Equal("application/json", r.Header.Get("Content-Type"))
			s.Equal(s.apiKey, r.Header.Get("x-api-key"))
			if response != "" {
				jsonData := []byte(response)
				if code, err := w.Write(jsonData); err != nil {
					w.WriteHeader(code)
				}
				return
			}
			if statusCode > 0 {
				w.WriteHeader(statusCode)
				return
			}
		}
		s.Fail("invalid url string")
	}))
}

func TestSegmentAPIManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SegmentAPIManagerTestSuite))
}
