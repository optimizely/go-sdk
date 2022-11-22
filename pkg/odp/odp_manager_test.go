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
	"context"
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/pkg/odp/cache"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/odp/segment"
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

func (m *MockSegmentManager) FetchQualifiedSegments(userID string, options []segment.OptimizelySegmentOption) (segments []string, err error) {
	args := m.Called(userID, options)
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
	o.odpManager = NewOdpManager("", false, 0, 0, WithOdpConfig(o.config), WithEventManager(o.eventManager), WithSegmentManager(o.segmentManager))
}

func (o *ODPManagerTestSuite) TestNewODPManagerNilParametersWithDisableFalse() {
	odpManager := NewOdpManager("", false, 0, 0)
	o.NotNil(odpManager.segmentsCache)
	o.NotNil(odpManager.OdpConfig)
	o.NotNil(odpManager.logger)
	o.NotNil(odpManager.processing)
	o.NotNil(odpManager.SegmentManager)
	o.NotNil(odpManager.EventManager)
}

func (o *ODPManagerTestSuite) TestNewODPManagerWithOptionssWithDisableFalse() {
	odpConfig := config.NewConfig("123", "456", []string{"123"})
	segmentsCache := cache.NewLRUCache(1, 2)
	segmentManager := segment.NewSegmentManager("", 1, 2, segment.WithSegmentsCache(segmentsCache))
	eventManager := event.NewBatchEventManager()
	odpManager := NewOdpManager("", false, -1, -1, WithOdpConfig(odpConfig), WithSegmentsCache(segmentsCache), WithSegmentManager(segmentManager), WithEventManager(eventManager))
	o.Equal(odpConfig, odpManager.OdpConfig)
	o.Equal(segmentsCache, odpManager.segmentsCache)
	o.Equal(segmentManager, odpManager.SegmentManager)
	o.Equal(eventManager, odpManager.EventManager)
	o.NotNil(odpManager.logger)
	o.NotNil(odpManager.processing)
}

func (o *ODPManagerTestSuite) TestNewODPManagersWithDisableTrue() {
	odpManager := NewOdpManager("", true, 0, 0)
	o.NotNil(odpManager.logger)
	o.False(odpManager.enabled)
	o.Nil(odpManager.segmentsCache)
	o.Nil(odpManager.OdpConfig)
	o.Nil(odpManager.processing)
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
	odpConfig := config.NewConfig("1", "2", []string{"123"})
	odpManager := NewOdpManager("", false, 0, 0, WithOdpConfig(odpConfig))
	o.Equal(odpConfig, odpManager.OdpConfig)
	o.Equal(odpConfig, odpManager.EventManager.(*event.BatchEventManager).OdpConfig)
	o.Equal(odpConfig, odpManager.SegmentManager.(*segment.DefaultSegmentManager).OdpConfig)

	// Check changes reflected after update
	odpConfig.Update("a", "b", []string{"1234"})
	o.Equal(odpConfig.GetAPIKey(), "a")
	o.Equal(odpConfig.GetAPIHost(), "b")
	o.Equal(odpConfig.GetSegmentsToCheck(), []string{"1234"})

	o.Equal(odpConfig.GetAPIKey(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetAPIKey())
	o.Equal(odpConfig.GetAPIHost(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetAPIHost())
	o.Equal(odpConfig.GetSegmentsToCheck(), odpManager.EventManager.(*event.BatchEventManager).OdpConfig.GetSegmentsToCheck())
}

func (o *ODPManagerTestSuite) TestFetchQualifiedSegments() {
	expectedSegments := []string{"1"}
	expectedError := errors.New("123")
	o.segmentManager.On("FetchQualifiedSegments", "1", []segment.OptimizelySegmentOption{}).Return([]string{"1"}, errors.New("123"))
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
