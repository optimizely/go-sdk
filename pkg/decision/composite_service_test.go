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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
)

type CompositeServiceFeatureTestSuite struct {
	suite.Suite
	decisionContext    FeatureDecisionContext
	mockFeatureService *MockFeatureDecisionService
	testUserContext    entities.UserContext
}

func (s *CompositeServiceFeatureTestSuite) SetupTest() {
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

func (s *CompositeServiceFeatureTestSuite) TestGetFeatureDecision() {
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

func (s *CompositeServiceFeatureTestSuite) TestDecisionListeners() {
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1111,
		Variation:  &testExp1111Var2222,
	}
	notificationCenter := notification.NewNotificationCenter()
	decisionService := &CompositeService{
		compositeFeatureService: s.mockFeatureService,
		notificationCenter:      notificationCenter,
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

func (s *CompositeServiceFeatureTestSuite) TestDecisionListenersNotificationWithFloatVariable() {

	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureDecisionService := NewCompositeFeatureService("", compositeExperimentService)
	s.decisionContext.Variable = entities.Variable{
		DefaultValue: "23.34",
		ID:           "1",
		Key:          "Key",
		Type:         entities.Double,
	}

	decisionService := &CompositeService{
		compositeFeatureService: compositeFeatureDecisionService,
		notificationCenter:      registry.GetNotificationCenter("some_key"),
	}
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)

	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	expectedDecisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "my_test_feature_3333", "source": FeatureTest,
		"sourceInfo":  map[string]string{"experimentKey": "test_experiment_1111", "variationKey": "2222"},
		"variableKey": "Key", "variableType": entities.Double, "variableValue": 23.34}}

	s.Equal(expectedDecisionInfo, note.DecisionInfo)

}

func (s *CompositeServiceFeatureTestSuite) TestDecisionListenersNotificationWithIntegerVariable() {

	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureDecisionService := NewCompositeFeatureService("", compositeExperimentService)
	s.decisionContext.Variable = entities.Variable{
		DefaultValue: "23",
		ID:           "1",
		Key:          "Key",
		Type:         entities.Integer,
	}

	decisionService := &CompositeService{
		compositeFeatureService: compositeFeatureDecisionService,
		notificationCenter:      registry.GetNotificationCenter("some_key"),
	}
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)

	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	expectedDecisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "my_test_feature_3333", "source": FeatureTest,
		"sourceInfo":  map[string]string{"experimentKey": "test_experiment_1111", "variationKey": "2222"},
		"variableKey": "Key", "variableType": entities.Integer, "variableValue": 23}}

	s.Equal(expectedDecisionInfo, note.DecisionInfo)

}

func (s *CompositeServiceFeatureTestSuite) TestDecisionListenersNotificationWithBoolVariable() {

	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureDecisionService := NewCompositeFeatureService("", compositeExperimentService)
	s.decisionContext.Variable = entities.Variable{
		DefaultValue: "true",
		ID:           "1",
		Key:          "Key",
		Type:         entities.Boolean,
	}

	decisionService := &CompositeService{
		compositeFeatureService: compositeFeatureDecisionService,
		notificationCenter:      registry.GetNotificationCenter("some_key"),
	}
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)

	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	expectedDecisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "my_test_feature_3333", "source": FeatureTest,
		"sourceInfo":  map[string]string{"experimentKey": "test_experiment_1111", "variationKey": "2222"},
		"variableKey": "Key", "variableType": entities.Boolean, "variableValue": true}}

	s.Equal(expectedDecisionInfo, note.DecisionInfo)

}

func (s *CompositeServiceFeatureTestSuite) TestDecisionListenersNotificationWithWrongTypelVariable() {

	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureDecisionService := NewCompositeFeatureService("", compositeExperimentService)
	s.decisionContext.Variable = entities.Variable{
		DefaultValue: "string",
		ID:           "1",
		Key:          "Key",
		Type:         entities.Double,
	}

	decisionService := &CompositeService{
		compositeFeatureService: compositeFeatureDecisionService,
		notificationCenter:      registry.GetNotificationCenter("some_key"),
	}
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)

	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	expectedDecisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "my_test_feature_3333", "source": FeatureTest,
		"sourceInfo":  map[string]string{"experimentKey": "test_experiment_1111", "variationKey": "2222"},
		"variableKey": "Key", "variableType": entities.Double, "variableValue": "string"}}

	s.Equal(expectedDecisionInfo, note.DecisionInfo)

}

func (s *CompositeServiceFeatureTestSuite) TestDecisionListenersNotificationWithNoVariable() {

	compositeExperimentService := NewCompositeExperimentService("")
	compositeFeatureDecisionService := NewCompositeFeatureService("", compositeExperimentService)
	s.decisionContext.Variable = entities.Variable{} //no variable

	decisionService := &CompositeService{
		compositeFeatureService: compositeFeatureDecisionService,
		notificationCenter:      registry.GetNotificationCenter("some_key"),
	}
	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)

	decisionService.GetFeatureDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	expectedDecisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "my_test_feature_3333", "source": FeatureTest,
		"sourceInfo": map[string]string{"experimentKey": "test_experiment_1111", "variationKey": "2222"}}}

	s.Equal(expectedDecisionInfo, note.DecisionInfo)
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
	mockExperimentService *MockExperimentDecisionService
	testUserContext       entities.UserContext
}

func (s *CompositeServiceExperimentTestSuite) SetupTest() {
	mockConfig := new(mockProjectConfig)
	s.decisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: mockConfig,
	}
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
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext).Return(expectedExperimentDecision, nil)
	experimentDecision, err := decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext)

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
	s.mockExperimentService.On("GetDecision", s.decisionContext, s.testUserContext).Return(expectedExperimentDecision, nil)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext)

	var numberOfCalls = 0
	callback := func(notification notification.DecisionNotification) {
		numberOfCalls++
	}
	id, _ := decisionService.OnDecision(callback)

	s.NotEqual(id, 0)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)

	err := decisionService.RemoveOnDecision(id)
	s.NoError(err)
	decisionService.GetExperimentDecision(s.decisionContext, s.testUserContext)
	s.Equal(numberOfCalls, 1)
}

func TestCompositeServiceTestSuites(t *testing.T) {
	suite.Run(t, new(CompositeServiceExperimentTestSuite))
	suite.Run(t, new(CompositeServiceFeatureTestSuite))
}
