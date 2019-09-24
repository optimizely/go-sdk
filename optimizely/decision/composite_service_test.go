/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/notification"
)

type CompositeServiceTestSuite struct {
	suite.Suite
	decisionContext    FeatureDecisionContext
	mockFeatureService *MockFeatureDecisionService
	testUserContext    entities.UserContext
}

func (s *CompositeServiceTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)
	s.decisionContext = FeatureDecisionContext{
		Feature:       &testFeat3333,
		ProjectConfig: mockConfig,
	}
	s.mockFeatureService = new(MockFeatureDecisionService)
	s.testUserContext = entities.UserContext{
		ID: "test_user",
	}
}

func (s *CompositeServiceTestSuite) TestGetFeatureDecision() {
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1111,
		Variation:  &testExp1111Var2222,
	}
	decisionService := &CompositeService{
		compositeFeatureService: s.mockFeatureService,
	}
	s.mockFeatureService.On("GetDecision", s.decisionContext, s.testUserContext).Return(expectedFeatureDecision, nil)
	featureDecision, err := decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	// Test assertions
	s.Equal(expectedFeatureDecision, featureDecision)
	s.NoError(err)
	s.mockFeatureService.AssertExpectations(s.T())
}

func (s *CompositeServiceTestSuite) TestDecisionListeners() {
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1111,
		Variation:  &testExp1111Var2222,
	}

	testFeatureDecisionService := new(MockFeatureDecisionService)
	testFeatureDecisionService.On("GetDecision", decisionContext, userContext).Return(expectedFeatureDecision, nil)

	decisionService := &CompositeService{
		featureDecisionServices: []FeatureService{testFeatureDecisionService},
		sdkKey:                  "sdkKey",
	}
	s.mockFeatureService.On("GetDecision", s.decisionContext, s.testUserContext).Return(expectedFeatureDecision, nil)
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	callback := func(notification notification.DecisionNotification) {
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	err := decisionService.RemoveOnDecision(id)
	s.NoError(err)
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)
}

func (s *CompositeServiceTestSuite) TestNewCompositeService() {
	notificationCenter := notification.NewNotificationCenter()
	compositeService := NewCompositeService(notificationCenter)
	s.Equal(notificationCenter, compositeService.notificationCenter)
	s.IsType(&CompositeExperimentService{}, compositeService.compositeExperimentService)
	s.IsType(&CompositeFeatureService{}, compositeService.compositeFeatureService)
}

func TestCompositeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CompositeServiceTestSuite))
}
