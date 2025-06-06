/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

package decision

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
)

type CompositeServiceFeatureTestSuite struct {
	suite.Suite
	decisionContext    FeatureDecisionContext
	options            *decide.Options
	reasons            decide.DecisionReasons
	mockFeatureService *MockFeatureDecisionService
	testUserContext    entities.UserContext
}

type MockNotificationCenter struct {
	mock.Mock
}

func (m *MockNotificationCenter) AddHandler(notificationType notification.Type, handler func(interface{})) (int, error) {
	args := m.Called(notificationType, handler)
	return args.Int(0), args.Error(1)
}

func (m *MockNotificationCenter) RemoveHandler(id int, notificationType notification.Type) error {
	args := m.Called(id, notificationType)
	return args.Error(0)
}

func (m *MockNotificationCenter) Send(notificationType notification.Type, notification interface{}) error {
	args := m.Called(notificationType, notification)
	return args.Error(0)
}

func (s *CompositeServiceFeatureTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)

	s.decisionContext = FeatureDecisionContext{
		Feature:       &testFeat3333,
		ProjectConfig: mockConfig,
	}
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)
	s.mockFeatureService = new(MockFeatureDecisionService)
	s.testUserContext = entities.UserContext{
		ID: "test_user",
	}
}

func (s *CompositeServiceFeatureTestSuite) TestGetFeatureDecision() {
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1111,
		Variation:  &testExp1111Var2222,
	}
	decisionService := &CompositeService{
		compositeFeatureService: s.mockFeatureService,
	}
	s.mockFeatureService.On("GetDecision", s.decisionContext, s.testUserContext, s.options).Return(expectedFeatureDecision, s.reasons, nil)
	featureDecision, _, err := decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext, s.options)

	// Test assertions
	s.Equal(expectedFeatureDecision, featureDecision)
	s.NoError(err)
	s.mockFeatureService.AssertExpectations(s.T())
}

func (s *CompositeServiceFeatureTestSuite) TestNewCompositeService() {
	notificationCenter := notification.NewNotificationCenter()
	compositeService := NewCompositeService("sdk_key")
	s.Equal(notificationCenter, compositeService.notificationCenter)
	s.IsType(&CompositeExperimentService{}, compositeService.compositeExperimentService)
	s.IsType(&CompositeFeatureService{}, compositeService.compositeFeatureService)
}

func (s *CompositeServiceFeatureTestSuite) TestNewCompositeServiceWithCustomOptions() {
	compositeExperimentService := NewCompositeExperimentService("")
	compositeService := NewCompositeService("sdk_key", WithCompositeExperimentService(compositeExperimentService))
	s.IsType(compositeExperimentService, compositeService.compositeExperimentService)
	s.IsType(&CompositeFeatureService{}, compositeService.compositeFeatureService)
}

type CompositeServiceExperimentTestSuite struct {
	suite.Suite
	decisionContext       ExperimentDecisionContext
	options               *decide.Options
	reasons               decide.DecisionReasons
	mockExperimentService *MockExperimentDecisionService
	testUserContext       entities.UserContext
}

func (s *CompositeServiceExperimentTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)
	s.decisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: mockConfig,
	}
	s.options = &decide.Options{}
	s.reasons = decide.NewDecisionReasons(s.options)
	s.mockExperimentService = new(MockExperimentDecisionService)
	s.testUserContext = entities.UserContext{
		ID: "test_user",
	}
}

func (s *CompositeServiceExperimentTestSuite) TestGetExperimentDecision() {
	expectedExperimentDecision := ExperimentDecision{
		Variation: &testExp1111Var2222,
	}
	decisionService := &CompositeService{
		compositeExperimentService: s.mockExperimentService,
	}
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options).Return(expectedExperimentDecision, s.reasons, nil)
	experimentDecision, _, err := decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)

	// Test assertions
	s.Equal(expectedExperimentDecision, experimentDecision)
	s.NoError(err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *CompositeServiceExperimentTestSuite) TestDecisionListeners() {
	expectedExperimentDecision := ExperimentDecision{
		Variation: &testExp1111Var2222,
	}
	notificationCenter := notification.NewNotificationCenter()
	decisionService := &CompositeService{
		compositeExperimentService: s.mockExperimentService,
		notificationCenter:         notificationCenter,
	}
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options).Return(expectedExperimentDecision, s.reasons, nil)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)

	var numberOfCalls = 0

	callback := func(notification notification.DecisionNotification) {
		numberOfCalls++

	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(0, id)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)

	s.Equal(numberOfCalls, 1)

	err := decisionService.RemoveOnDecision(id)
	s.NoError(err)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)
	s.Equal(numberOfCalls, 1)
}

// Add these test methods to CompositeServiceExperimentTestSuite

func (s *CompositeServiceExperimentTestSuite) TestGetExperimentDecisionWithError() {
	// Test line 79: Error from compositeExperimentService.GetDecision
	expectedError := errors.New("experiment service error")
	decisionService := &CompositeService{
		compositeExperimentService: s.mockExperimentService,
	}

	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options).
		Return(ExperimentDecision{}, s.reasons, expectedError)

	_, _, err := decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)

	s.Error(err)
	s.Equal(expectedError, err)
	s.mockExperimentService.AssertExpectations(s.T())
}

func (s *CompositeServiceExperimentTestSuite) TestGetExperimentDecisionNotificationSendError() {
	// Test line 98: Error from notificationCenter.Send
	expectedExperimentDecision := ExperimentDecision{
		Variation: &testExp1111Var2222,
	}

	// Create a mock notification center that returns an error
	mockNotificationCenter := &MockNotificationCenter{}
	mockNotificationCenter.On("Send", notification.Decision, mock.AnythingOfType("notification.DecisionNotification")).
		Return(errors.New("notification send error"))

	decisionService := &CompositeService{
		compositeExperimentService: s.mockExperimentService,
		notificationCenter:         mockNotificationCenter,
		logger:                     logging.GetLogger("test", "CompositeService"),
	}

	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options).
		Return(expectedExperimentDecision, s.reasons, nil)

	experimentDecision, _, err := decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options)

	// FIX: The method DOES return the notification error at the end
	s.Error(err)
	s.Contains(err.Error(), "notification send error")
	s.Equal(expectedExperimentDecision, experimentDecision) // Decision should still be returned
	s.mockExperimentService.AssertExpectations(s.T())
	mockNotificationCenter.AssertExpectations(s.T())
}

func (s *CompositeServiceExperimentTestSuite) TestOnDecisionAddHandlerError() {
	// Test line 102: Error from notificationCenter.AddHandler
	mockNotificationCenter := &MockNotificationCenter{}
	mockNotificationCenter.On("AddHandler", notification.Decision, mock.AnythingOfType("func(interface {})")).
		Return(0, errors.New("add handler error"))

	decisionService := &CompositeService{
		notificationCenter: mockNotificationCenter,
		logger:             logging.GetLogger("test", "CompositeService"),
	}

	callback := func(notification.DecisionNotification) {}
	id, err := decisionService.OnDecision(callback)

	s.Error(err)
	s.Equal(0, id)
	mockNotificationCenter.AssertExpectations(s.T())
}

func (s *CompositeServiceExperimentTestSuite) TestRemoveOnDecisionError() {
	// Test lines 120-123: Error from notificationCenter.RemoveHandler
	mockNotificationCenter := &MockNotificationCenter{}
	mockNotificationCenter.On("RemoveHandler", 123, notification.Decision).
		Return(errors.New("remove handler error"))

	decisionService := &CompositeService{
		notificationCenter: mockNotificationCenter,
		logger:             logging.GetLogger("test", "CompositeService"),
	}

	err := decisionService.RemoveOnDecision(123)

	s.Error(err)
	mockNotificationCenter.AssertExpectations(s.T())
}

func (s *CompositeServiceExperimentTestSuite) TestOnDecisionInvalidPayload() {
	// Test lines 129-132: Invalid payload in OnDecision callback
	notificationCenter := notification.NewNotificationCenter()
	decisionService := &CompositeService{
		notificationCenter: notificationCenter,
		logger:             logging.GetLogger("test", "CompositeService"),
	}

	var callbackCalled bool
	callback := func(notification.DecisionNotification) {
		callbackCalled = true
	}

	id, err := decisionService.OnDecision(callback)
	s.NoError(err)
	s.NotEqual(0, id)

	// Send invalid payload to trigger the warning path
	err = notificationCenter.Send(notification.Decision, "invalid_payload")
	s.NoError(err)          // Send should succeed, but callback shouldn't be called
	s.False(callbackCalled) // Callback should not be called with invalid payload
}

func TestCompositeServiceTestSuites(t *testing.T) {
	suite.Run(t, new(CompositeServiceExperimentTestSuite))
	suite.Run(t, new(CompositeServiceFeatureTestSuite))
}
