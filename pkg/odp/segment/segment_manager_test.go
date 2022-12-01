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

	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	"github.com/stretchr/testify/suite"
)

type SegmentManagerTestSuite struct {
	suite.Suite
	segmentManager    *DefaultSegmentManager
	config            config.Config
	segmentAPIManager *MockSegmentAPIManager
	userID            string
}

func (s *SegmentManagerTestSuite) SetupTest() {
	s.config = config.NewConfig("", "", nil)
	s.segmentAPIManager = &MockSegmentAPIManager{}
	s.segmentManager = NewSegmentManager("", WithSegmentsCacheSize(10), WithSegmentsCacheTimeoutInSecs(10), WithAPIManager(s.segmentAPIManager))
	s.userID = "test-user"
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerNilParameters() {
	segmentManager := NewSegmentManager("")
	s.NotNil(segmentManager.apiManager)
	s.NotNil(segmentManager.segmentsCache)
	s.Equal(utils.DefaultSegmentsCacheSize, segmentManager.segmentsCacheSize)
	s.Equal(utils.DefaultSegmentsCacheTimeout, segmentManager.segmentsCacheTimeoutInSecs)
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerCustomOptions() {
	customCache := &TestCache{}
	segmentManager := NewSegmentManager("", WithSegmentsCache(customCache), WithSegmentsCacheSize(10), WithSegmentsCacheTimeoutInSecs(10), WithAPIManager(s.segmentAPIManager))
	s.Equal(customCache, segmentManager.segmentsCache)
	s.Equal(10, segmentManager.segmentsCacheSize)
	s.Equal(int64(10), segmentManager.segmentsCacheTimeoutInSecs)
	s.Equal(s.segmentAPIManager, segmentManager.apiManager)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNilConfig() {
	segmentManager := NewSegmentManager("")
	segments, err := segmentManager.FetchQualifiedSegments(s.config, s.userID, nil)
	s.Nil(segments)
	s.Error(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNoSegmentsToCheckInConfig() {
	odpConfig := config.NewConfig("a", "b", nil)
	segmentManager := NewSegmentManager("")
	segments, err := segmentManager.FetchQualifiedSegments(odpConfig, s.userID, nil)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheMiss() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.Equal(10, s.segmentManager.segmentsCacheSize)
	s.Equal(int64(10), s.segmentManager.segmentsCacheTimeoutInSecs)
	s.setCache("123", []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments(s.config, s.userID, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheHit() {
	expectedSegments := []string{"a"}
	s.config.Update("valid", "host", []string{"new-customer"})
	s.setCache(s.userID, expectedSegments)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.config, s.userID, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSegmentsError() {
	s.config.Update("invalid-key", "host", []string{"new-customer"})

	segments, err := s.segmentManager.FetchQualifiedSegments(s.config, s.userID, nil)
	s.Error(err)
	s.Nil(segments)
}

func (s *SegmentManagerTestSuite) TestOptionsIgnoreCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userID, []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments(s.config, s.userID, []OptimizelySegmentOption{IgnoreCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestOptionsResetCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userID, []string{"a"})
	s.setCache("123", []string{"a"})
	s.setCache("456", []string{"a"})

	segments, err := s.segmentManager.FetchQualifiedSegments(s.config, s.userID, []OptimizelySegmentOption{ResetCache})
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
