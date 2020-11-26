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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

type CompositeServiceFeatureTestSuite struct {
	suite.Suite
	decisionContext    FeatureDecisionContext
	options            *decide.OptimizelyDecideOptions
	reasons            decide.DecisionReasons
	mockFeatureService *MockFeatureDecisionService
	testUserContext    entities.UserContext
}

func (s *CompositeServiceFeatureTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)

	s.decisionContext = FeatureDecisionContext{
		Feature:       &testFeat3333,
		ProjectConfig: mockConfig,
	}
	s.options = &decide.OptimizelyDecideOptions{}
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
	s.mockFeatureService.On("GetDecision", s.decisionContext, s.testUserContext, s.options, s.reasons).Return(expectedFeatureDecision, nil)
	featureDecision, err := decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext, s.options, s.reasons)

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
	options               *decide.OptimizelyDecideOptions
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
	s.options = &decide.OptimizelyDecideOptions{}
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
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options, s.reasons).Return(expectedExperimentDecision, nil)
	experimentDecision, err := decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options, s.reasons)

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
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext, s.options, s.reasons).Return(expectedExperimentDecision, nil)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options, s.reasons)

	var numberOfCalls = 0

	callback := func(notification notification.DecisionNotification) {
		numberOfCalls++

	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(0, id)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options, s.reasons)

	s.Equal(numberOfCalls, 1)

	err := decisionService.RemoveOnDecision(id)
	s.NoError(err)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext, s.options, s.reasons)
	s.Equal(numberOfCalls, 1)
}

func TestCompositeServiceTestSuites(t *testing.T) {
	suite.Run(t, new(CompositeServiceExperimentTestSuite))
	suite.Run(t, new(CompositeServiceFeatureTestSuite))
}
