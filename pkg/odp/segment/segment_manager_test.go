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

// Package segment //
package segment

import (
	"fmt"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	"github.com/stretchr/testify/suite"
)

type SegmentManagerTestSuite struct {
	suite.Suite
	segmentManager    *DefaultSegmentManager
	segmentAPIManager *MockSegmentAPIManager
	userID            string
}

func (s *SegmentManagerTestSuite) SetupTest() {
	s.segmentAPIManager = &MockSegmentAPIManager{}
	s.segmentManager = NewSegmentManager("", WithSegmentsCacheSize(10), WithSegmentsCacheTimeout(10*time.Second), WithAPIManager(s.segmentAPIManager))
	s.userID = "test-user"
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerNilParameters() {
	segmentManager := NewSegmentManager("")
	s.NotNil(segmentManager.apiManager)
	s.NotNil(segmentManager.segmentsCache)
	s.Equal(utils.DefaultSegmentsCacheSize, segmentManager.segmentsCacheSize)
	s.Equal(utils.DefaultSegmentsCacheTimeout, segmentManager.segmentsCacheTimeout)
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerCustomOptions() {
	customCache := &TestCache{}
	segmentManager := NewSegmentManager("", WithSegmentsCache(customCache), WithSegmentsCacheSize(10), WithSegmentsCacheTimeout(10*time.Second), WithAPIManager(s.segmentAPIManager))
	s.Equal(customCache, segmentManager.segmentsCache)
	s.Equal(10, segmentManager.segmentsCacheSize)
	s.Equal(10*time.Second, segmentManager.segmentsCacheTimeout)
	s.Equal(s.segmentAPIManager, segmentManager.apiManager)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNilConfig() {
	segmentManager := NewSegmentManager("")
	segments, err := segmentManager.FetchQualifiedSegments("", "", s.userID, nil, nil)
	s.Nil(segments)
	s.Error(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNoSegmentsToCheckInConfig() {
	segmentManager := NewSegmentManager("")
	segments, err := segmentManager.FetchQualifiedSegments("a", "b", s.userID, nil, nil)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheMiss() {
	expectedSegments := []string{"new-customer"}
	s.Equal(10, s.segmentManager.segmentsCacheSize)
	s.Equal(10*time.Second, s.segmentManager.segmentsCacheTimeout)
	s.setCache("123", []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments("valid", "host", s.userID, expectedSegments, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheHit() {
	expectedSegments := []string{"a"}
	s.setCache(s.userID, expectedSegments)

	segments, err := s.segmentManager.FetchQualifiedSegments("valid", "host", s.userID, []string{"new-customer"}, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSegmentsError() {
	segments, err := s.segmentManager.FetchQualifiedSegments("invalid-key", "host", s.userID, []string{"new-customer"}, nil)
	s.Error(err)
	s.Nil(segments)
}

func (s *SegmentManagerTestSuite) TestOptionsIgnoreCache() {
	expectedSegments := []string{"new-customer"}
	s.setCache(s.userID, []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments("valid", "host", s.userID, expectedSegments, []OptimizelySegmentOption{IgnoreCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestOptionsResetCache() {
	expectedSegments := []string{"new-customer"}
	s.setCache(s.userID, []string{"a"})
	s.setCache("123", []string{"a"})
	s.setCache("456", []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments("valid", "host", s.userID, expectedSegments, []OptimizelySegmentOption{ResetCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestMakeCacheKey() {
	s.Equal(fmt.Sprintf("%s-$-test-user", utils.OdpFSUserIDKey), MakeCacheKey(s.userID))
}

// Helper methods
func (s *SegmentManagerTestSuite) setCache(userID string, value []string) {
	cacheKey := MakeCacheKey(userID)
	s.segmentManager.segmentsCache.Save(cacheKey, value)
}

type MockSegmentAPIManager struct {
}

func (s *MockSegmentAPIManager) FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string) ([]string, error) {
	if apiKey == "invalid-key" {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "403 Forbidden")
	}
	return segmentsToCheck, nil
}

func TestSegmentManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SegmentManagerTestSuite))
}

type TestCache struct {
}

func (l *TestCache) Save(key string, value interface{}) {
}
func (l *TestCache) Lookup(key string) interface{} {
	return nil
}
func (l *TestCache) Reset() {
}
func (l *TestCache) Remove(key string) {
}
