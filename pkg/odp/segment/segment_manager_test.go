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

// Package segment //
package segment

import (
	"fmt"
	"testing"

	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	"github.com/stretchr/testify/mock"
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
	s.segmentManager = NewSegmentManager("", 10, 10, WithOdpConfig(s.config), WithAPIManager(s.segmentAPIManager))
	s.userID = "test-user"
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerNilParameters() {
	segmentManager := NewSegmentManager("", 0, 0)
	s.NotNil(segmentManager.apiManager)
	s.NotNil(segmentManager.segmentsCache)
}

func (s *SegmentManagerTestSuite) TestNewSegmentManagerCustomCache() {
	customCache := &TestCache{}
	segmentManager := NewSegmentManager("", 0, 0, WithSegmentsCache(customCache))
	s.Equal(customCache, segmentManager.segmentsCache)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNilConfig() {
	segmentManager := NewSegmentManager("", 0, 0)
	segments, err := segmentManager.FetchQualifiedSegments(s.userID, nil)
	s.Nil(segments)
	s.Error(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsNoSegmentsToCheckInConfig() {
	segmentManager := NewSegmentManager("", 0, 0, WithOdpConfig(config.NewConfig("a", "b", nil)))
	segments, err := segmentManager.FetchQualifiedSegments(s.userID, nil)
	s.Empty(segments)
	s.Nil(err)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheMiss() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache("123", []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userID, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSuccessCacheHit() {
	expectedSegments := []string{"a"}
	s.config.Update("valid", "host", []string{"new-customer"})
	s.setCache(s.userID, expectedSegments)
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userID, nil)
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestFetchSegmentsSegmentsError() {
	s.config.Update("invalid-key", "host", []string{"new-customer"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userID, nil)
	s.Error(err)
	s.Nil(segments)
}

func (s *SegmentManagerTestSuite) TestOptionsIgnoreCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userID, []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userID, []OptimizelySegmentOption{IgnoreCache})
	s.Nil(err)
	s.Equal(expectedSegments, segments)
}

func (s *SegmentManagerTestSuite) TestOptionsResetCache() {
	expectedSegments := []string{"new-customer"}
	s.config.Update("valid", "host", expectedSegments)
	s.setCache(s.userID, []string{"a"})
	s.setCache("123", []string{"a"})
	s.setCache("456", []string{"a"})
	s.segmentAPIManager.On("FetchSegments", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	segments, err := s.segmentManager.FetchQualifiedSegments(s.userID, []OptimizelySegmentOption{ResetCache})
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
	mock.Mock
}

func (s *MockSegmentAPIManager) FetchQualifiedSegments(odpConfig config.Config, userID string) ([]string, error) {
	if odpConfig.GetAPIKey() == "invalid-key" {
		return nil, fmt.Errorf(utils.FetchSegmentsFailedError, "403 Forbidden")
	}
	return odpConfig.GetSegmentsToCheck(), nil
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
