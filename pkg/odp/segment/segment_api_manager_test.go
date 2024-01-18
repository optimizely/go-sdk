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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	pkgUtils "github.com/optimizely/go-sdk/v2/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type SegmentAPIManagerTestSuite struct {
	suite.Suite
	segmentAPIManager                                                                                    *DefaultSegmentAPIManager
	apiKey, userID                                                                                       string
	goodResponseData, goodEmptyResponseData                                                              string
	invalidIdentifierResponseData, invalidErrorResponseData, otherExceptionResponseData, badResponseData string
	invalidEdgeResponseData, invalidNodeResponseData                                                     string
	liveOdpAPIKey, liveOdpAPIHost, liveOdpValidUserID                                                    string
}

func (s *SegmentAPIManagerTestSuite) SetupTest() {
	s.apiKey = "test-api-key"
	s.segmentAPIManager = NewSegmentAPIManager("", nil)
	s.userID = "test-user-value"
	s.liveOdpAPIKey = "W4WzcEs-ABgXorzY7h1LCQ"
	s.liveOdpAPIHost = "https://api.zaius.com"
	s.liveOdpValidUserID = "tester-101"
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
				"code": "INVALID_IDENTIFIER_EXCEPTION",
				"classification": "DataFetchingException"
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
	requester := pkgUtils.NewHTTPRequester(logging.GetLogger("", ""))
	segmentManager := NewSegmentAPIManager("", requester)
	s.Equal(requester, segmentManager.requester)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsSuccess() {
	s.NotNil(s.segmentAPIManager.requester)
	ts := s.getTestServer(0, 0, s.goodResponseData)
	defer ts.Close()
	segmentsToCheck := []string{"a", "b", "c"}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.NoError(err)
	s.Len(segments, 1)
	s.Equal("a", segments[0])
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsSuccessWithEmptySegments() {
	ts := s.getTestServer(0, 0, s.goodEmptyResponseData)
	defer ts.Close()
	segmentsToCheck := []string{"a", "b", "c"}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.NoError(err)
	s.NotNil(segments)
	s.Len(segments, 0)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidIdentifier() {
	ts := s.getTestServer(0, 0, s.invalidIdentifierResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(errors.New(utils.InvalidSegmentIdentifier), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidError() {
	ts := s.getTestServer(0, 0, s.invalidErrorResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsOtherException() {
	ts := s.getTestServer(0, 0, s.otherExceptionResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "TestExceptionClass"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsArrayResponse() {
	ts := s.getTestServer(0, 0, `[]`)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidEdgeResponse() {
	ts := s.getTestServer(0, 0, s.invalidEdgeResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidNodeResponse() {
	ts := s.getTestServer(0, 0, s.invalidNodeResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsBadResponse() {
	ts := s.getTestServer(0, 0, s.badResponseData)
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "decode error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsNetworkTimeout() {
	ts := s.getTestServer(0, 100, s.goodResponseData)
	defer ts.Close()
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = 10 * time.Millisecond
	segmentsToCheck := []string{"a", "b", "c"}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Error(err)
	s.Nil(segments)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegments400() {
	ts := s.getTestServer(403, 0, "")
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "403 Forbidden"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegments500() {
	ts := s.getTestServer(500, 0, "")
	defer ts.Close()
	segmentsToCheck := []string{}
	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.apiKey, ts.URL, s.userID, segmentsToCheck)
	s.Nil(segments)
	s.Equal(fmt.Errorf(utils.FetchSegmentsFailedError, "500 Internal Server Error"), err)
}

func (s *SegmentAPIManagerTestSuite) TestFetchQualifiedSegmentsInvalidURL() {
	segments, err := s.segmentAPIManager.FetchQualifiedSegments("123", "456", s.userID, nil)
	s.Nil(segments)
	s.Error(err)
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
	template := "query($userId: String, $audiences: [String]) {customer(fs_user_id: $userId) {audiences(subset: $audiences) {edges {node {name state}}}}}"
	expectedBody := []map[string]interface{}{
		{"query": template, "variables": map[string]interface{}{"audiences": []string{}, "userId": "value-1"}},
		{"query": template, "variables": map[string]interface{}{"audiences": []string{"a", "b"}, "userId": "value-1"}},
	}

	for i := range segmentsToCheck {
		query := s.segmentAPIManager.createRequestQuery("value-1", segmentsToCheck[i])
		expected := expectedBody[i]
		s.True(reflect.DeepEqual(expected, query))
	}
}

// Tests with live ODP server
// func (s *SegmentAPIManagerTestSuite) TestLiveOdpGraphQL() {
// 	segmentsToCheck := []string{"segment-1"}
// 	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.liveOdpAPIKey, s.liveOdpAPIHost, s.liveOdpValidUserID, segmentsToCheck)
// 	s.NoError(err)
// 	s.Empty(segments, "none of the test segments in the live ODP server")
// }

// func (s *SegmentAPIManagerTestSuite) TestLiveOdpGraphQLDefaultParametersUserNotRegistered() {
// 	segmentsToCheck := []string{"segment-1"}
// 	segments, err := s.segmentAPIManager.FetchQualifiedSegments(s.liveOdpAPIKey, s.liveOdpAPIHost, "not-registered-user-1", segmentsToCheck)
// 	s.Error(err)
// 	s.Nil(segments)
// }

func (s *SegmentAPIManagerTestSuite) getTestServer(statusCode, timeout int, response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == graphqlAPIEndpointPath {
			s.Equal("POST", r.Method)
			s.Equal(pkgUtils.ContentTypeJSON, r.Header.Get(pkgUtils.HeaderContentType))
			s.Equal(s.apiKey, r.Header.Get(utils.OdpAPIKeyHeader))
			if timeout > 0 {
				time.Sleep(time.Duration(timeout) * time.Millisecond)
			}
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
