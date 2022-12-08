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

// Package odp //
package odp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/odp/cache"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockEventManager struct {
	mock.Mock
	event.Manager
}

func (m *MockEventManager) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockEventManager) IdentifyUser(userID string) {
	m.Called(userID)
}

func (m *MockEventManager) ProcessEvent(odpEvent event.Event) bool {
	return m.Called(odpEvent).Get(0).(bool)
}

func (m *MockEventManager) FlushEvents() {
	m.Called()
}

type MockSegmentManager struct {
	mock.Mock
	segment.Manager
}

func (m *MockSegmentManager) FetchQualifiedSegments(odpConfig config.Config, userID string, options []segment.OptimizelySegmentOption) (segments []string, err error) {
	args := m.Called(odpConfig, userID, options)
	if segArray, ok := args.Get(0).([]string); ok {
		segments = segArray
	}
	return segments, args.Error(1)
}

func (m *MockSegmentManager) Reset() {
	m.Called()
}

type MockConfig struct {
	mock.Mock
	config.Config
}

func (m *MockConfig) Update(apiKey, apiHost string, segmentsToCheck []string) bool {
	return m.Called(apiKey, apiHost, segmentsToCheck).Get(0).(bool)
}

func (m *MockConfig) GetAPIKey() string {
	return m.Called().Get(0).(string)
}

func (m *MockConfig) GetAPIHost() string {
	return m.Called().Get(0).(string)
}

func (m *MockConfig) GetSegmentsToCheck() []string {
	return m.Called().Get(0).([]string)
}

func (m *MockConfig) IsOdpServiceIntegrated() bool {
	return m.Called().Get(0).(bool)
}

type ODPManagerTestSuite struct {
	suite.Suite
	config         *MockConfig
	odpManager     *DefaultOdpManager
	eventManager   *MockEventManager
	segmentManager *MockSegmentManager
	userID         string
}

func (o *ODPManagerTestSuite) SetupTest() {
	o.userID = "test-user"
	o.config = &MockConfig{}
	o.eventManager = &MockEventManager{}
	o.segmentManager = &MockSegmentManager{}
	o.odpManager = NewOdpManager("", false, WithSegmentsCacheSize(0), WithSegmentsCacheTimeout(0), WithEventManager(o.eventManager), WithSegmentManager(o.segmentManager))
	o.odpManager.OdpConfig = o.config
}

func (o *ODPManagerTestSuite) TestNewODPManagerNilParametersWithDisableFalse() {
	odpManager := NewOdpManager("", false)
	o.NotNil(odpManager.OdpConfig)
	o.NotNil(odpManager.logger)
	o.NotNil(odpManager.SegmentManager)
	o.NotNil(odpManager.EventManager)
}

func (o *ODPManagerTestSuite) TestNegativeSegmentCacheSizeAndTimeout() {
	cacheTimeout := -1 * time.Second
	odpManager := NewOdpManager("", false, WithSegmentsCacheSize(-1), WithSegmentsCacheTimeout(cacheTimeout))
	o.Equal(-1, odpManager.segmentsCacheSize)
	o.Equal(cacheTimeout, odpManager.segmentsCacheTimeout)
}

func (o *ODPManagerTestSuite) TestDefaultCacheSizeAndTimeout() {
	odpManager := NewOdpManager("", false)
	o.Equal(utils.DefaultSegmentsCacheSize, odpManager.segmentsCacheSize)
	o.Equal(utils.DefaultSegmentsCacheTimeout, odpManager.segmentsCacheTimeout)
}

func (o *ODPManagerTestSuite) TestNewODPManagerWithOptionsWithDisableFalse() {
	expectedCacheTimeout := 1 * time.Second
	segmentsCache := cache.NewLRUCache(1, 2*time.Second)
	segmentManager := segment.NewSegmentManager("", segment.WithSegmentsCache(segmentsCache))
	eventManager := event.NewBatchEventManager()
	odpManager := NewOdpManager("", false, WithSegmentsCacheSize(1), WithSegmentsCacheTimeout(expectedCacheTimeout), WithSegmentsCache(segmentsCache), WithSegmentManager(segmentManager), WithEventManager(eventManager))
	o.Equal(segmentsCache, odpManager.segmentsCache)
	o.Equal(segmentManager, odpManager.SegmentManager)
	o.Equal(eventManager, odpManager.EventManager)
	o.Equal(1, odpManager.segmentsCacheSize)
	o.Equal(expectedCacheTimeout, odpManager.segmentsCacheTimeout)
	o.NotNil(odpManager.logger)
}

func (o *ODPManagerTestSuite) TestNewODPManagerWithDisableTrue() {
	odpManager := NewOdpManager("", true)
	o.NotNil(odpManager.logger)
	o.False(odpManager.enabled)
	o.Nil(odpManager.segmentsCache)
	o.Nil(odpManager.OdpConfig)
	o.Nil(odpManager.SegmentManager)
	o.Nil(odpManager.EventManager)
}

func (o *ODPManagerTestSuite) TestODPManagerAPIsWithDisableTrue() {
	o.odpManager.enabled = false
	_, _ = o.odpManager.FetchQualifiedSegments("1", []segment.OptimizelySegmentOption{})
	o.odpManager.IdentifyUser(o.userID)
	o.odpManager.SendOdpEvent("1", "abc", nil, nil)
	o.odpManager.Update("123", "456", []string{"abc"})
	o.config.AssertNumberOfCalls(o.T(), "Update", 0)
	o.segmentManager.AssertNumberOfCalls(o.T(), "FetchQualifiedSegments", 0)
	o.segmentManager.AssertNumberOfCalls(o.T(), "Reset", 0)
	o.eventManager.AssertNumberOfCalls(o.T(), "IdentifyUser", 0)
	o.eventManager.AssertNumberOfCalls(o.T(), "ProcessEvent", 0)
	o.eventManager.AssertNumberOfCalls(o.T(), "FlushEvents", 0)
}

func (o *ODPManagerTestSuite) TestODPConfigReferencePassed() {
	odpManager := NewOdpManager("", false)
	o.NotNil(odpManager.OdpConfig)
	o.Equal(odpManager.OdpConfig, odpManager.EventManager.(*event.BatchEventManager).OdpConfig)

	// Check changes reflected after update
	odpManager.OdpConfig.Update("a", "b", []string{"1234"})
	o.Equal(odpManager.OdpConfig.GetAPIKey(), "a")
	o.Equal(odpManager.OdpConfig.GetAPIHost(), "b")
	o.Equal(odpManager.OdpConfig.GetSegmentsToCheck(), []string{"1234"})

	o.Equal(odpManager.OdpConfig.GetAPIKey(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetAPIKey())
	o.Equal(odpManager.OdpConfig.GetAPIHost(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetAPIHost())
	o.Equal(odpManager.OdpConfig.GetSegmentsToCheck(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetSegmentsToCheck())
}

func (o *ODPManagerTestSuite) TestFetchQualifiedSegments() {
	expectedSegments := []string{"1"}
	expectedError := errors.New("123")
	o.segmentManager.On("FetchQualifiedSegments", o.odpManager.OdpConfig, "1", []segment.OptimizelySegmentOption{}).Return([]string{"1"}, errors.New("123"))
	segments, err := o.odpManager.FetchQualifiedSegments("1", []segment.OptimizelySegmentOption{})
	o.segmentManager.AssertExpectations(o.T())
	o.Equal(expectedSegments, segments)
	o.Equal(expectedError, err)
}

func (o *ODPManagerTestSuite) TestIdentifyUser() {
	o.eventManager.On("IdentifyUser", o.userID)
	o.odpManager.IdentifyUser(o.userID)
	o.segmentManager.AssertExpectations(o.T())
}

func (o *ODPManagerTestSuite) TestSendOdpEvent() {
	userEvent := event.Event{
		Action: "123",
		Type:   "456",
		Identifiers: map[string]string{
			"abc": "123",
		},
		Data: map[string]interface{}{
			"abc":                 nil,
			"idempotence_id":      234,
			"data_source_type":    "456",
			"data_source":         true,
			"data_source_version": 6.78,
		}}
	o.eventManager.On("ProcessEvent", userEvent).Return(true)
	o.True(o.odpManager.SendOdpEvent(userEvent.Type, userEvent.Action, userEvent.Identifiers, userEvent.Data))
	o.segmentManager.AssertExpectations(o.T())
}

func (o *ODPManagerTestSuite) TestUpdate() {
	apiKey := "1"
	apiHost := "2"
	segmentsToCheck := []string{"123"}

	o.eventManager.On("FlushEvents")
	o.config.On("Update", apiKey, apiHost, segmentsToCheck).Return(true)
	o.segmentManager.On("Reset")
	o.odpManager.Update(apiKey, apiHost, segmentsToCheck)

	o.segmentManager.AssertExpectations(o.T())
}

func TestODPManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ODPManagerTestSuite))
}
