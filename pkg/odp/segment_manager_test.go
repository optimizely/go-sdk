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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SegmentManagerTestSuite struct {
	suite.Suite
	segmentManager     *DefaultSegmentManager
	config             Config
	segmentAPIManager  *MockSegmentAPIManager
	userValue, userKey string
}

func (s *SegmentManagerTestSuite) SetupTest() {
	s.config = NewConfig("", "", nil)
	s.segmentAPIManager = &MockSegmentAPIManager{}
	s.segmentManager = NewSegmentManager("", 10, 10, WithODPConfig(s.config), WithAPIManager(s.segmentAPIManager))
	s.userValue = "test-user"
	s.userKey = "vuid"
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerNilParameters() {
	segmentManager := NewSegmentManager("", 0, 0)
	s.NotNil(segmentManager.segmentAPIManager)
	s.NotNil(segmentManager.segmentsCache)
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerCustomCache() {
	customCache := &TestCache{}
	segmentManager := NewSegmentManager("", 0, 0, WithSegmentsCache(customCache))
	s.Equal(customCache, segmentManager.segmentsCache)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNilConfig() {
	segmentManager := NewSegmentManager("", 0, 0)
	segments, err := segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, nil)
	s.Nil(segments)
	s.Error(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNoSegmentsToCheckInConfig() {
	segmentManager := NewSegmentManager("", 0, 0, WithODPConfig(NewConfig("a", "b", nil)))
	segments, err := segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, nil)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheMiss() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userKey, "123", []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheHit() {
	expectedSegments := []string{"a"}
	s.config.Update("valid", "host", []string{"new-customer"})
	s.setCache(s.userKey, s.userValue, expectedSegments)
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSegmentsError() {
	s.config.Update("invalid-key", "host", []string{"new-customer"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, nil)
	s.Error(err)
	s.Nil(segments)
}

func (s *SegmentManagerTestSuite) TestOptionsIgnoreCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userKey, s.userValue, []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, []OptimizelySegmentOption{IgnoreCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestOptionsResetCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userKey, s.userValue, []string{"a"})
	s.setCache(s.userKey, "123", []string{"a"})
	s.setCache(s.userKey, "456", []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userKey, s.userValue, []OptimizelySegmentOption{ResetCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestMakeCacheKey() {
	s.Equal("vuid-$-test-user", MakeCacheKey(s.userKey, s.userValue))
}

// Helper methods
func (s *SegmentManagerTestSuite) setCache(userKey, userValue string, value []string) {
	cacheKey := MakeCacheKey(userKey, userValue)
	s.segmentManager.segmentsCache.Save(cacheKey, value)
}

type MockSegmentAPIManager struct {
	mock.Mock
}

func (s *MockSegmentAPIManager) FetchQualifiedSegments(config Config, userKey, userValue string) ([]string, error) {
	if config.GetAPIKey() == "invalid-key" {
		return nil, fmt.Errorf(fetchSegmentsFailedError, "403 Forbidden")
	}
	return config.GetSegmentsToCheck(), nil
}

func TestSegmentManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SegmentManagerTestSuite))
}
