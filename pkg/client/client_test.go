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

package client

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func ValidProjectConfigManager() *MockProjectConfigManager {
	p := new(MockProjectConfigManager)
	p.projectConfig = new(TestConfig)
	return p
}

func InValidProjectConfigManager() *MockProjectConfigManager {
	return nil
}

func getMockConfigAndMapsForVariables(featureKey string, variables []variable) (mockConfig *MockProjectConfig, variableMap map[string]entities.Variable, varVariableMap map[string]entities.VariationVariable) {
	mockConfig = new(MockProjectConfig)
	variableMap = make(map[string]entities.Variable)
	varVariableMap = make(map[string]entities.VariationVariable)

	for i, v := range variables {
		id := strconv.Itoa(i)
		varVariableMap[id] = entities.VariationVariable{
			ID:    id,
			Value: v.varVal,
		}

		variableMap[id] = entities.Variable{
			DefaultValue: v.defaultVal,
			ID:           id,
			Key:          v.key,
			Type:         v.varType,
		}

		mockConfig.On("GetVariableByKey", featureKey, v.key).Return(v.varVal, nil)
	}
	return
}

type variable struct {
	key        string
	defaultVal string
	varVal     string
	varType    entities.VariableType
	expected   interface{}
}

type MockProcessor struct {
	Events []event.UserEvent
	mock.Mock
}

func (m *MockProcessor) ProcessEvent(event event.UserEvent) bool {
	result := m.Called(event).Get(0).(bool)
	if result {
		m.Events = append(m.Events, event)
	}
	return result
}

func (m *MockProcessor) OnEventDispatch(callback func(logEvent event.LogEvent)) (int, error) {
	return 0, nil
}

func (m *MockProcessor) RemoveOnEventDispatch(id int) error {
	return nil
}

type MockNotificationCenter struct {
	notification.Center
	mock.Mock
}

func (m *MockNotificationCenter) AddHandler(notificationType notification.Type, callback func(interface{})) (int, error) {
	args := m.Called(notificationType, callback)
	var err error
	if tmpError, ok := args.Get(1).(error); ok {
		err = tmpError
	}
	return args.Get(0).(int), err
}

func (m *MockNotificationCenter) RemoveHandler(id int, notificationType notification.Type) error {
	args := m.Called(id, notificationType)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (m *MockNotificationCenter) Send(notificationType notification.Type, notification interface{}) error {
	args := m.Called(notificationType, notification)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

type TestConfig struct {
	config.ProjectConfig
}

func (TestConfig) GetEventByKey(key string) (entities.Event, error) {
	if key == "sample_conversion" {
		return entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, nil
	}

	return entities.Event{}, errors.New("No conversion")
}

func (TestConfig) GetFeatureByKey(string) (entities.Feature, error) {
	return entities.Feature{}, nil
}

func (TestConfig) GetProjectID() string {
	return "15389410617"
}
func (TestConfig) GetRevision() string {
	return "7"
}
func (TestConfig) GetAccountID() string {
	return "8362480420"
}
func (TestConfig) GetAnonymizeIP() bool {
	return true
}
func (TestConfig) GetAttributeID(key string) string { // returns "" if there is no id
	return ""
}
func (TestConfig) GetBotFiltering() bool {
	return false
}
func (TestConfig) GetClientName() string {
	return "go-sdk"
}
func (TestConfig) GetClientVersion() string {
	return "1.0.0"
}

func TestTrack(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockDecisionService := new(MockDecisionService)
	mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)

	client := OptimizelyClient{
		ConfigManager:   ValidProjectConfigManager(),
		DecisionService: mockDecisionService,
		EventProcessor:  mockProcessor,
		logger:          logging.GetLogger("", ""),
	}

	err := client.Track("sample_conversion", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	assert.NoError(t, err)
	assert.True(t, len(mockProcessor.Events) == 1)
	assert.True(t, mockProcessor.Events[0].VisitorID == "1212121")
	assert.True(t, mockProcessor.Events[0].EventContext.ProjectID == "15389410617")

}

func TestTrackFailEventNotFound(t *testing.T) {
	mockProcessor := &MockProcessor{}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   ValidProjectConfigManager(),
		DecisionService: mockDecisionService,
		EventProcessor:  mockProcessor,
		logger:          logging.GetLogger("", ""),
	}

	err := client.Track("bob", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	assert.NoError(t, err)
	assert.True(t, len(mockProcessor.Events) == 0)

}

func TestTrackPanics(t *testing.T) {
	mockProcessor := &MockProcessor{}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   new(PanickingConfigManager),
		DecisionService: mockDecisionService,
		EventProcessor:  mockProcessor,
		logger:          logging.GetLogger("", ""),
	}

	err := client.Track("bob", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	assert.Error(t, err)
	assert.True(t, len(mockProcessor.Events) == 0)

}

func TestGetEnabledFeaturesPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.Empty(t, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableBool(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		validBool         bool
		result            bool
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "true", varType: entities.Boolean, validBool: true,
			featureEnabled: true, result: true},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Boolean, validBool: false,
			featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "5", varType: entities.Integer, validBool: false,
			featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", validBool: false,
			featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "true", varType: entities.Boolean, validBool: true,
			featureEnabled: false, result: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "false",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		client := OptimizelyClient{
			ConfigManager:   mockConfigManager,
			DecisionService: mockDecisionService,
			logger:          logging.GetLogger("", ""),
		}
		result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)

		} else {
			assert.Error(t, err)
			assert.False(t, result)
		}
		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableBoolWithNotification(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		decisionInfo      map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "true", varType: entities.Boolean, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Boolean, "variableValue": true}}, featureEnabled: true},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Boolean, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Boolean, "variableValue": "stringvalue"}}, featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "5", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": "5"}}, featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.VariableType(""), "variableValue": "true"}}, featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "true", varType: entities.Boolean, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Boolean, "variableValue": false}}, featureEnabled: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "false",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		notificationCenter := notification.NewNotificationCenter()
		client := OptimizelyClient{
			ConfigManager:      mockConfigManager,
			DecisionService:    mockDecisionService,
			logger:             logging.GetLogger("", ""),
			notificationCenter: notificationCenter,
		}
		var numberOfCalls = 0
		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfCalls++
		}
		mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
		mockDecisionService.notificationCenter = notificationCenter
		id, _ := mockDecisionService.OnDecision(callback)

		assert.NotEqual(t, id, 0)
		client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)

		assert.Equal(t, numberOfCalls, 1)
		assert.Equal(t, ts.decisionInfo, note.DecisionInfo)

		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableBoolPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableDouble(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		validDouble       bool
		result            float64
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "5", varType: entities.Double, validDouble: true,
			featureEnabled: true, result: 5.0},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Double, validDouble: false,
			featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "5", varType: entities.Integer, validDouble: false,
			featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", validDouble: false,
			featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "5", varType: entities.Double, validDouble: true,
			featureEnabled: false, result: 4.0},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "4",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		client := OptimizelyClient{
			ConfigManager:   mockConfigManager,
			DecisionService: mockDecisionService,
			logger:          logging.GetLogger("", ""),
		}
		result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
		if ts.validDouble {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)

		} else {
			assert.Error(t, err)
			assert.Equal(t, float64(0), result)
		}
		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableDoubleWithNotification(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		decisionInfo      map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "5", varType: entities.Double, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Double, "variableValue": 5.0}}, featureEnabled: true},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Double, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Double, "variableValue": "stringvalue"}}, featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "5", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": "5"}}, featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.VariableType(""), "variableValue": "true"}}, featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "5", varType: entities.Double, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Double, "variableValue": 4.0}}, featureEnabled: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "4",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		notificationCenter := notification.NewNotificationCenter()
		client := OptimizelyClient{
			ConfigManager:      mockConfigManager,
			DecisionService:    mockDecisionService,
			logger:             logging.GetLogger("", ""),
			notificationCenter: notificationCenter,
		}
		var numberOfCalls = 0
		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfCalls++
		}
		mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
		mockDecisionService.notificationCenter = notificationCenter
		id, _ := mockDecisionService.OnDecision(callback)

		assert.NotEqual(t, id, 0)
		client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)

		assert.Equal(t, numberOfCalls, 1)
		assert.Equal(t, ts.decisionInfo, note.DecisionInfo)

		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableDoublePanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(0), result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableInteger(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		validInteger      bool
		result            int
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "5", varType: entities.Integer, validInteger: true,
			featureEnabled: true, result: 5},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Integer, validInteger: false,
			featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "true", varType: entities.Boolean, validInteger: false,
			featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", validInteger: false,
			featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "5", varType: entities.Integer, validInteger: true,
			featureEnabled: false, result: 4},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "4",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		client := OptimizelyClient{
			ConfigManager:   mockConfigManager,
			DecisionService: mockDecisionService,
			logger:          logging.GetLogger("", ""),
		}
		result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
		if ts.validInteger {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)

		} else {
			assert.Error(t, err)
			assert.Equal(t, 0, result)
		}
		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableIntegerWithNotification(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		decisionInfo      map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "5", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": 5}}, featureEnabled: true},
		{name: "InvalidValue", testVariableValue: "stringvalue", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": "stringvalue"}}, featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "5", varType: entities.Boolean, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Boolean, "variableValue": "5"}}, featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.VariableType(""), "variableValue": "true"}}, featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "5", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": 4}}, featureEnabled: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "4",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		notificationCenter := notification.NewNotificationCenter()
		client := OptimizelyClient{
			ConfigManager:      mockConfigManager,
			DecisionService:    mockDecisionService,
			logger:             logging.GetLogger("", ""),
			notificationCenter: notificationCenter,
		}
		var numberOfCalls = 0
		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfCalls++
		}
		mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
		mockDecisionService.notificationCenter = notificationCenter
		id, _ := mockDecisionService.OnDecision(callback)

		assert.NotEqual(t, id, 0)
		client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)

		assert.Equal(t, numberOfCalls, 1)
		assert.Equal(t, ts.decisionInfo, note.DecisionInfo)

		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableIntegerPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 0, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableSting(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		validString       bool
		result            string
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "teststring", varType: entities.String, validString: true,
			featureEnabled: true, result: "teststring"},
		{name: "InvalidVariableType", testVariableValue: "true", varType: entities.Boolean, validString: false,
			featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", validString: false,
			featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "some_value", varType: entities.String, validString: true,
			featureEnabled: false, result: "default"},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "default",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		client := OptimizelyClient{
			ConfigManager:   mockConfigManager,
			DecisionService: mockDecisionService,
			logger:          logging.GetLogger("", ""),
		}
		result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
		if ts.validString {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)

		} else {
			assert.Error(t, err)
			assert.Equal(t, "", result)
		}
		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableStringWithNotification(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		decisionInfo      map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "teststring", varType: entities.String, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.String, "variableValue": "teststring"}}, featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "true", varType: entities.Boolean, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Boolean, "variableValue": ""}}, featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "true", varType: "", decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.VariableType(""), "variableValue": ""}}, featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "some_value", varType: entities.String, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.String, "variableValue": "default"}}, featureEnabled: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "default",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		notificationCenter := notification.NewNotificationCenter()
		client := OptimizelyClient{
			ConfigManager:      mockConfigManager,
			DecisionService:    mockDecisionService,
			logger:             logging.GetLogger("", ""),
			notificationCenter: notificationCenter,
		}
		var numberOfCalls = 0
		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfCalls++
		}
		mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
		mockDecisionService.notificationCenter = notificationCenter
		id, _ := mockDecisionService.OnDecision(callback)

		assert.NotEqual(t, id, 0)
		client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)

		assert.Equal(t, numberOfCalls, 1)
		assert.Equal(t, ts.decisionInfo, note.DecisionInfo)

		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}
func TestGetFeatureVariableStringPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "", result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableJSON(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		stringRepr        string
		varType           entities.VariableType
		featureEnabled    bool
		validJson         bool
		mapRepr           map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "{\"test\":12}", varType: entities.JSON, validJson: true,
			featureEnabled: true, mapRepr: map[string]interface{}{"test": 12.0}, stringRepr: "{\"test\":12}"},
		{name: "InvalidValue", testVariableValue: "{\"test\": }", varType: entities.JSON, validJson: false,
			featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "{}", varType: entities.Integer, validJson: false,
			featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "{}", varType: "", validJson: false,
			featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "{\"test\":12}", varType: entities.JSON, validJson: true,
			featureEnabled: false, mapRepr: map[string]interface{}{}, stringRepr: "{}"},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "{}",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		client := OptimizelyClient{
			ConfigManager:   mockConfigManager,
			DecisionService: mockDecisionService,
			logger:          logging.GetLogger("", ""),
		}
		result, err := client.GetFeatureVariableJSON(testFeatureKey, testVariableKey, testUserContext)
		if ts.validJson {
			assert.NoError(t, err)
			assert.NotNil(t, result)

			resultStr, err := result.ToString()
			assert.NoError(t, err)
			assert.Equal(t, ts.stringRepr, resultStr)

			resultMap := result.ToMap()
			assert.Equal(t, ts.mapRepr, resultMap)
		} else {
			assert.Error(t, err)
			assert.Nil(t, result)
		}
		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}

func TestGetFeatureVariableJSONWithNotification(t *testing.T) {

	type test struct {
		name              string
		testVariableValue string
		varType           entities.VariableType
		featureEnabled    bool
		decisionInfo      map[string]interface{}
	}

	testSuite := []test{
		{name: "ValidValue", testVariableValue: "{\"test\":12}", varType: entities.JSON, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.JSON, "variableValue": map[string]interface{}{"test": 12.0}}}, featureEnabled: true},
		{name: "InvalidValue", testVariableValue: "{\"test\": }", varType: entities.JSON, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.JSON, "variableValue": "{\"test\": }"}}, featureEnabled: true},
		{name: "InvalidVariableType", testVariableValue: "{}", varType: entities.Integer, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.Integer, "variableValue": "{}"}}, featureEnabled: true},
		{name: "EmptyVariableType", testVariableValue: "{}", varType: "", decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.VariableType(""), "variableValue": "{}"}}, featureEnabled: true},
		{name: "DefaultValueIfFeatureNotEnabled", testVariableValue: "{\"test\":12}", varType: entities.JSON, decisionInfo: map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": false, "featureKey": "test_feature_key", "source": decision.Source(""),
			"sourceInfo": map[string]string{}, "variableKey": "test_feature_flag_key", "variableType": entities.JSON, "variableValue": map[string]interface{}{}}}, featureEnabled: false},
	}

	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	for _, ts := range testSuite {
		testVariationVariable := entities.VariationVariable{
			ID:    "1",
			Value: ts.testVariableValue,
		}
		testVariable := entities.Variable{
			DefaultValue: "{}",
			ID:           "1",
			Key:          "test_feature_flag_key",
			Type:         ts.varType,
		}
		testVariation := getTestVariationWithFeatureVariable(ts.featureEnabled, testVariationVariable)
		testExperiment := entities.Experiment{
			ID:         "111111",
			Variations: map[string]entities.Variation{"22222": testVariation},
		}
		testFeature := getTestFeature(testFeatureKey, testExperiment)
		mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
		mockConfigManager := new(MockProjectConfigManager)
		mockConfigManager.On("GetConfig").Return(mockConfig, nil)

		testDecisionContext := decision.FeatureDecisionContext{
			Feature:       &testFeature,
			ProjectConfig: mockConfig,
			Variable:      testVariable,
		}

		expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
		mockDecisionService := new(MockDecisionService)
		mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

		notificationCenter := notification.NewNotificationCenter()
		client := OptimizelyClient{
			ConfigManager:      mockConfigManager,
			DecisionService:    mockDecisionService,
			logger:             logging.GetLogger("", ""),
			notificationCenter: notificationCenter,
		}
		var numberOfCalls = 0
		note := notification.DecisionNotification{}
		callback := func(notification notification.DecisionNotification) {
			note = notification
			numberOfCalls++
		}
		mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
		mockDecisionService.notificationCenter = notificationCenter
		id, _ := mockDecisionService.OnDecision(callback)

		assert.NotEqual(t, id, 0)
		client.GetFeatureVariableJSON(testFeatureKey, testVariableKey, testUserContext)

		assert.Equal(t, numberOfCalls, 1)
		assert.Equal(t, ts.decisionInfo, note.DecisionInfo)

		mockConfig.AssertExpectations(t)
		mockConfigManager.AssertExpectations(t)
		mockDecisionService.AssertExpectations(t)
	}
}
func TestGetFeatureVariableJSONPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableJSON(testFeatureKey, testVariableKey, testUserContext)
	assert.Nil(t, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableErrorCases(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(nil, errors.New("no project config available"))
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}
	_, err1 := client.GetFeatureVariableBoolean("test_feature_key", "test_variable_key", testUserContext)
	_, err2 := client.GetFeatureVariableDouble("test_feature_key", "test_variable_key", testUserContext)
	_, err3 := client.GetFeatureVariableInteger("test_feature_key", "test_variable_key", testUserContext)
	_, err4 := client.GetFeatureVariableString("test_feature_key", "test_variable_key", testUserContext)
	_, err5 := client.GetFeatureVariableJSON("test_feature_key", "test_variable_key", testUserContext)
	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Error(t, err3)
	assert.Error(t, err4)
	assert.Error(t, err5)
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockConfigManager.AssertNotCalled(t, "GetVariableByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")
}

func TestGetProjectConfigIsValid(t *testing.T) {
	mockConfigManager := ValidProjectConfigManager()

	client := OptimizelyClient{
		ConfigManager: mockConfigManager,
		logger:        logging.GetLogger("", ""),
	}

	actual, err := client.getProjectConfig()

	assert.Nil(t, err)
	assert.Equal(t, mockConfigManager.projectConfig, actual)
}

func TestGetProjectConfigIsInValid(t *testing.T) {

	client := OptimizelyClient{
		ConfigManager: InValidProjectConfigManager(),
		logger:        logging.GetLogger("", ""),
	}

	actual, err := client.getProjectConfig()

	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestGetOptimizelyConfig(t *testing.T) {
	mockConfigManager := ValidProjectConfigManager()

	client := OptimizelyClient{
		ConfigManager: mockConfigManager,
		logger:        logging.GetLogger("", ""),
	}

	optimizelyConfig := client.GetOptimizelyConfig()

	assert.Equal(t, &config.OptimizelyConfig{Revision: "232"}, optimizelyConfig)
}

func TestGetFeatureDecisionValid(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
		Variable:      testVariable,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	_, featureDecision, err := client.getFeatureDecision(testFeatureKey, testVariableKey, testUserContext)
	assert.Nil(t, err)
	assert.Equal(t, expectedFeatureDecision, featureDecision)
}

func TestGetFeatureDecisionErrProjectConfig(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          testVariableKey,
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, errors.New("project config error"))

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
		Variable:      testVariable,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testVariableKey, testUserContext)
	assert.Error(t, err)
}

func TestGetFeatureDecisionPanicProjectConfig(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          testVariableKey,
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
		Variable:      testVariable,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)

	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   &PanickingConfigManager{},
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testVariableKey, testUserContext)
	assert.Error(t, err)
}

func TestGetFeatureDecisionPanicDecisionService(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          testVariableKey,
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: &PanickingDecisionService{},
		logger:          logging.GetLogger("", ""),
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testVariableKey, testUserContext)
	assert.Error(t, err)
	assert.EqualError(t, err, "I'm panicking")
}

func TestGetFeatureDecisionErrFeatureDecision(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          testVariableKey,
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
		Variable:      testVariable,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, errors.New("error feature"))

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	_, decision, err := client.getFeatureDecision(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, expectedFeatureDecision, decision)
	assert.NoError(t, err)
}

func TestGetAllFeatureVariablesWithDecision(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	variables := []variable{
		{key: "var_str", defaultVal: "default", varVal: "var", varType: entities.String, expected: "var"},
		{key: "var_bool", defaultVal: "false", varVal: "true", varType: entities.Boolean, expected: true},
		{key: "var_int", defaultVal: "10", varVal: "20", varType: entities.Integer, expected: 20},
		{key: "var_double", defaultVal: "1.0", varVal: "2.0", varType: entities.Double, expected: 2.0},
		{key: "var_json", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: entities.JSON,
			expected: map[string]interface{}{"field1": 12.0, "field2": "some_value"}},
		{key: "var_unknown", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: "",
			expected: "{\"field1\":12.0, \"field2\": \"some_value\"}"},
	}

	mockConfig, variableMap, varVariableMap := getMockConfigAndMapsForVariables(testFeatureKey, variables)
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
		Variables:      varVariableMap,
	}

	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = variableMap
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	enabled, variationMap, err := client.GetAllFeatureVariablesWithDecision(testFeatureKey, testUserContext)
	assert.NoError(t, err)
	assert.True(t, enabled)

	for _, v := range variables {
		assert.Equal(t, v.expected, variationMap[v.key])
	}
}

func TestGetAllFeatureVariablesWithDecisionWithNotification(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	variables := []variable{
		{key: "var_str", defaultVal: "default", varVal: "var", varType: entities.String, expected: "var"},
		{key: "var_bool", defaultVal: "false", varVal: "true", varType: entities.Boolean, expected: true},
		{key: "var_int", defaultVal: "10", varVal: "20", varType: entities.Integer, expected: 20},
		{key: "var_double", defaultVal: "1.0", varVal: "2.0", varType: entities.Double, expected: 2.0},
		{key: "var_json", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: entities.JSON,
			expected: map[string]interface{}{"field1": 12.0, "field2": "some_value"}},
	}

	mockConfig, variableMap, varVariableMap := getMockConfigAndMapsForVariables(testFeatureKey, variables)
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
		Variables:      varVariableMap,
	}

	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = variableMap
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	notificationCenter := notification.NewNotificationCenter()
	client := OptimizelyClient{
		ConfigManager:      mockConfigManager,
		DecisionService:    mockDecisionService,
		logger:             logging.GetLogger("", ""),
		notificationCenter: notificationCenter,
	}
	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
	mockDecisionService.notificationCenter = notificationCenter
	id, _ := mockDecisionService.OnDecision(callback)

	assert.NotEqual(t, id, 0)
	client.GetAllFeatureVariablesWithDecision(testFeatureKey, testUserContext)

	decisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
		"sourceInfo": map[string]string{}, "variableValues": map[string]interface{}{"var_bool": true, "var_double": 2.0, "var_int": 20,
			"var_json": map[string]interface{}{"field1": 12.0, "field2": "some_value"}, "var_str": "var"}}}
	assert.Equal(t, numberOfCalls, 1)
	assert.Equal(t, decisionInfo, note.DecisionInfo)

}
func TestGetAllFeatureVariablesWithDecisionWithError(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          testVariableKey,
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = map[string]entities.Variable{testVariable.Key: testVariable}
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, errors.New(""))

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	enabled, variationMap, err := client.GetAllFeatureVariablesWithDecision(testFeatureKey, testUserContext)

	// if we have a decision, but also a non-fatal error, we should return the decision
	assert.True(t, enabled)
	assert.Equal(t, testVariableValue, variationMap[testVariableKey])
	assert.NoError(t, err)
}

func TestGetAllFeatureVariablesWithDecisionWithoutFeature(t *testing.T) {
	invalidFeatureKey := "non-existent-feature"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", invalidFeatureKey).Return(entities.Feature{}, errors.New(""))
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	enabled, variationMap, err := client.GetAllFeatureVariablesWithDecision(invalidFeatureKey, testUserContext)

	// if we have a decision, but also a non-fatal error, we should return the decision
	assert.False(t, enabled)
	assert.Equal(t, 0, len(variationMap))
	assert.NoError(t, err)
}

func TestGetDetailedFeatureDecisionUnsafeWithNotification(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	variables := []variable{
		{key: "var_str", defaultVal: "default", varVal: "var", varType: entities.String, expected: "var"},
		{key: "var_bool", defaultVal: "false", varVal: "true", varType: entities.Boolean, expected: true},
		{key: "var_int", defaultVal: "10", varVal: "20", varType: entities.Integer, expected: 20},
		{key: "var_double", defaultVal: "1.0", varVal: "2.0", varType: entities.Double, expected: 2.0},
		{key: "var_json", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: entities.JSON,
			expected: map[string]interface{}{"field1": 12.0, "field2": "some_value"}},
	}

	mockConfig, variableMap, varVariableMap := getMockConfigAndMapsForVariables(testFeatureKey, variables)
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
		Variables:      varVariableMap,
	}

	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = variableMap
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	notificationCenter := notification.NewNotificationCenter()
	client := OptimizelyClient{
		ConfigManager:      mockConfigManager,
		DecisionService:    mockDecisionService,
		logger:             logging.GetLogger("", ""),
		notificationCenter: notificationCenter,
	}
	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
	mockDecisionService.notificationCenter = notificationCenter
	id, _ := mockDecisionService.OnDecision(callback)

	assert.NotEqual(t, id, 0)
	client.GetDetailedFeatureDecisionUnsafe(testFeatureKey, testUserContext, true)

	decisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "test_feature_key", "source": decision.Source(""),
		"sourceInfo": map[string]string{}, "variableValues": map[string]interface{}{"var_bool": true, "var_double": 2.0, "var_int": 20,
			"var_json": map[string]interface{}{"field1": 12.0, "field2": "some_value"}, "var_str": "var"}}}
	assert.Equal(t, numberOfCalls, 1)
	assert.Equal(t, decisionInfo, note.DecisionInfo)
}

func TestGetDetailedFeatureDecisionUnsafeWithTrackingDisabled(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	variables := []variable{
		{key: "var_str", defaultVal: "default", varVal: "var", varType: entities.String, expected: "var"},
		{key: "var_bool", defaultVal: "false", varVal: "true", varType: entities.Boolean, expected: true},
		{key: "var_int", defaultVal: "10", varVal: "20", varType: entities.Integer, expected: 20},
		{key: "var_double", defaultVal: "1.0", varVal: "2.0", varType: entities.Double, expected: 2.0},
		{key: "var_json", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: entities.JSON,
			expected: map[string]interface{}{"field1": 12.0, "field2": "some_value"}},
		{key: "var_unknown", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: "",
			expected: "{\"field1\":12.0, \"field2\": \"some_value\"}"},
	}

	mockConfig, variableMap, varVariableMap := getMockConfigAndMapsForVariables(testFeatureKey, variables)
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
		Variables:      varVariableMap,
	}

	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = variableMap
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	decision, err := client.GetDetailedFeatureDecisionUnsafe(testFeatureKey, testUserContext, true)
	assert.NoError(t, err)
	assert.True(t, decision.Enabled)

	for _, v := range variables {
		assert.Equal(t, v.expected, decision.VariableMap[v.key])
	}
	assert.Equal(t, decision.ExperimentKey, "")
	assert.Equal(t, decision.VariationKey, "")
}

func TestGetDetailedFeatureDecisionUnsafeWithoutFeature(t *testing.T) {
	invalidFeatureKey := "non-existent-feature"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", invalidFeatureKey).Return(entities.Feature{}, errors.New(""))
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	decision, err := client.GetDetailedFeatureDecisionUnsafe(invalidFeatureKey, testUserContext, true)

	// if we have a decision, but also a non-fatal error, we should return the decision
	assert.False(t, decision.Enabled)
	assert.Equal(t, 0, len(decision.VariableMap))
	assert.NoError(t, err)
}

func TestGetDetailedFeatureDecisionUnsafeWithError(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariation := getTestVariationWithFeatureVariable(true, entities.VariationVariable{})
	testExperiment := entities.Experiment{}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	mockConfig := getMockConfig(testFeatureKey, "", testFeature, entities.Variable{})
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, errors.New(""))

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, errors.New(""))

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	decision, err := client.GetDetailedFeatureDecisionUnsafe(testFeatureKey, testUserContext, true)
	assert.False(t, decision.Enabled)
	assert.Error(t, err)
}

func TestGetDetailedFeatureDecisionUnsafeWithFeatureTestAndTrackingEnabled(t *testing.T) {
	mockConfig := new(MockProjectConfig)
	mockConfigManager := new(MockProjectConfigManager)
	mockDecisionService := new(MockDecisionService)
	mockEventProcessor := new(MockEventProcessor)
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test happy path
	testVariation := makeTestVariation("green", true)
	testExperiment := makeTestExperimentWithVariations("number_1", []entities.Variation{testVariation})
	testFeature := makeTestFeatureWithExperiment("feature_1", testExperiment)
	mockConfig.On("GetFeatureByKey", testFeature.Key).Return(testFeature, nil)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	mockEventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent"))

	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  &testVariation,
		Source:     decision.FeatureTest,
	}

	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		EventProcessor:  mockEventProcessor,
		logger:          logging.GetLogger("", ""),
	}

	decision, err := client.GetDetailedFeatureDecisionUnsafe(testFeature.Key, testUserContext, false)
	assert.NoError(t, err)
	assert.True(t, decision.Enabled)
	assert.Equal(t, decision.ExperimentKey, "number_1")
	assert.Equal(t, decision.VariationKey, "green")

	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
	mockEventProcessor.AssertExpectations(t)
}

func TestGetAllFeatureVariables(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	variables := []variable{
		{key: "var_str", defaultVal: "default", varVal: "var", varType: entities.String, expected: "var"},
		{key: "var_bool", defaultVal: "false", varVal: "true", varType: entities.Boolean, expected: true},
		{key: "var_int", defaultVal: "10", varVal: "20", varType: entities.Integer, expected: 20},
		{key: "var_double", defaultVal: "1.0", varVal: "2.0", varType: entities.Double, expected: 2.0},
		{key: "var_json", defaultVal: "{}", varVal: "{\"field1\":12.0, \"field2\": \"some_value\"}", varType: entities.JSON,
			expected: map[string]interface{}{"field1": 12.0, "field2": "some_value"}},
	}

	mockConfig, variableMap, varVariableMap := getMockConfigAndMapsForVariables(testFeatureKey, variables)
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
		Variables:      varVariableMap,
	}

	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.VariableMap = variableMap
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	optlyJson, err := client.GetAllFeatureVariables(testFeatureKey, testUserContext)
	assert.NoError(t, err)
	assert.NotNil(t, optlyJson)
	variationMap := optlyJson.ToMap()
	assert.NoError(t, err)

	for _, v := range variables {
		assert.Equal(t, v.expected, variationMap[v.key])
	}

	jsonVarMap, ok := variationMap["var_json"].(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, 12.0, jsonVarMap["field1"])
	assert.Equal(t, "some_value", jsonVarMap["field2"])
}

func TestGetAllFeatureVariablesWithoutFeature(t *testing.T) {
	invalidFeatureKey := "non-existent-feature"
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", invalidFeatureKey).Return(entities.Feature{}, errors.New(""))
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	optlyJson, err := client.GetAllFeatureVariables(invalidFeatureKey, testUserContext)
	assert.NoError(t, err)
	assert.NotNil(t, optlyJson)

	variationMap := optlyJson.ToMap()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(variationMap))

	variationString, err := optlyJson.ToString()
	assert.Equal(t, "{}", variationString)
}

// Helper Methods
func getTestFeatureDecision(experiment entities.Experiment, variation entities.Variation) decision.FeatureDecision {
	return decision.FeatureDecision{
		Experiment: experiment,
		Variation:  &variation,
	}
}

func getTestVariationWithFeatureVariable(featureEnabled bool, variable entities.VariationVariable) entities.Variation {
	return entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: featureEnabled,
		Variables:      map[string]entities.VariationVariable{variable.ID: variable},
	}
}

func getMockConfig(featureKey string, variableKey string, feature entities.Feature, variable entities.Variable) *MockProjectConfig {
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", featureKey).Return(feature, nil)
	mockConfig.On("GetVariableByKey", featureKey, variableKey).Return(variable, nil)
	return mockConfig
}

func getTestFeature(featureKey string, experiment entities.Experiment) entities.Feature {
	return entities.Feature{
		ID:                 "22222",
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{experiment},
	}
}

type ClientTestSuiteAB struct {
	suite.Suite
	mockConfig          *MockProjectConfig
	mockConfigManager   *MockProjectConfigManager
	mockDecisionService *MockDecisionService
	mockEventProcessor  *MockEventProcessor
}

func (s *ClientTestSuiteAB) SetupTest() {
	s.mockConfig = new(MockProjectConfig)
	s.mockConfigManager = new(MockProjectConfigManager)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)
	s.mockDecisionService = new(MockDecisionService)
	s.mockEventProcessor = new(MockEventProcessor)
}

func (s *ClientTestSuiteAB) TestActivate() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testExperiment := makeTestExperiment("test_exp_1")
	s.mockConfig.On("GetExperimentByKey", "test_exp_1").Return(testExperiment, nil)
	s.mockConfig.On("GetExperimentByKey", "test_exp_2").Return(testExperiment, errors.New("Experiment not found"))

	testDecisionContext := decision.ExperimentDecisionContext{
		Experiment:    &testExperiment,
		ProjectConfig: s.mockConfig,
	}

	expectedVariation := testExperiment.Variations["v2"]
	expectedExperimentDecision := decision.ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockDecisionService.On("GetExperimentDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)
	s.mockEventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent"))

	testClient := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		EventProcessor:  s.mockEventProcessor,
		logger:          logging.GetLogger("", ""),
	}

	variationKey1, err1 := testClient.Activate("test_exp_1", testUserContext)
	s.NoError(err1)
	s.Equal(expectedVariation.Key, variationKey1)

	// should not return error for experiment not found.
	variationKey2, err2 := testClient.Activate("test_exp_2", testUserContext)
	s.NoError(err2)
	s.Equal("", variationKey2)

	s.mockConfig.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
	s.mockEventProcessor.AssertExpectations(s.T())
}

func (s *ClientTestSuiteAB) TestActivatePanics() {
	// ensure that we recover if the SDK panics while getting variation
	testUserContext := entities.UserContext{}
	testClient := OptimizelyClient{
		ConfigManager:   new(PanickingConfigManager),
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	variationKey, err := testClient.Activate("test_exp_1", testUserContext)
	s.Equal("", variationKey)
	s.EqualError(err, "I'm panicking")
}

func (s *ClientTestSuiteAB) TestActivateInvalidConfig() {
	testUserContext := entities.UserContext{}

	mockConfigManager := new(MockProjectConfigManager)
	expectedError := errors.New("no project config available")
	mockConfigManager.On("GetConfig").Return(s.mockConfig, expectedError)
	testClient := OptimizelyClient{
		ConfigManager: mockConfigManager,
		logger:        logging.GetLogger("", ""),
	}

	variationKey, err := testClient.Activate("test_exp_1", testUserContext)
	s.Equal("", variationKey)
	s.Error(err)
	s.Equal(expectedError, err)
}

func (s *ClientTestSuiteAB) TestGetVariation() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testExperiment := makeTestExperiment("test_exp_1")
	s.mockConfig.On("GetExperimentByKey", "test_exp_1").Return(testExperiment, nil)

	testDecisionContext := decision.ExperimentDecisionContext{
		Experiment:    &testExperiment,
		ProjectConfig: s.mockConfig,
	}

	expectedVariation := testExperiment.Variations["v2"]
	expectedExperimentDecision := decision.ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockDecisionService.On("GetExperimentDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	testClient := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	variationKey, err := testClient.GetVariation("test_exp_1", testUserContext)
	s.NoError(err)
	s.Equal(expectedVariation.Key, variationKey)
	s.mockConfig.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
	s.mockEventProcessor.AssertNotCalled(s.T(), "ProcessEvent", mock.AnythingOfType("event.UserEvent"))
}

func (s *ClientTestSuiteAB) TestGetVariationWithDecisionError() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testExperiment := makeTestExperiment("test_exp_1")
	s.mockConfig.On("GetExperimentByKey", "test_exp_1").Return(testExperiment, nil)

	testDecisionContext := decision.ExperimentDecisionContext{
		Experiment:    &testExperiment,
		ProjectConfig: s.mockConfig,
	}

	expectedVariation := testExperiment.Variations["v2"]
	expectedExperimentDecision := decision.ExperimentDecision{
		Variation: &expectedVariation,
	}
	s.mockDecisionService.On("GetExperimentDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, errors.New(""))

	testClient := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	variationKey, err := testClient.GetVariation("test_exp_1", testUserContext)
	s.NoError(err)
	s.Equal(expectedVariation.Key, variationKey)
	s.mockConfig.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
	s.mockEventProcessor.AssertNotCalled(s.T(), "ProcessEvent", mock.AnythingOfType("event.UserEvent"))
}

func (s *ClientTestSuiteAB) TestGetVariationPanics() {
	// ensure that we recover if the SDK panics while getting variation
	testUserContext := entities.UserContext{}
	testClient := OptimizelyClient{
		ConfigManager:   new(PanickingConfigManager),
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}

	variationKey, err := testClient.GetVariation("test_exp_1", testUserContext)
	s.Equal("", variationKey)
	s.EqualError(err, "I'm panicking")
}

type ClientTestSuiteFM struct {
	suite.Suite
	mockConfig          *MockProjectConfig
	mockConfigManager   *MockProjectConfigManager
	mockDecisionService *MockDecisionService
	mockEventProcessor  *MockEventProcessor
}

func (s *ClientTestSuiteFM) SetupTest() {
	s.mockConfig = new(MockProjectConfig)
	s.mockConfigManager = new(MockProjectConfigManager)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)
	s.mockDecisionService = new(MockDecisionService)
	s.mockEventProcessor = new(MockEventProcessor)
}

func (s *ClientTestSuiteFM) TestIsFeatureEnabled() {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test happy path
	testVariation := makeTestVariation("green", true)
	testExperiment := makeTestExperimentWithVariations("number_1", []entities.Variation{testVariation})
	testFeature := makeTestFeatureWithExperiment("feature_1", testExperiment)
	s.mockConfig.On("GetFeatureByKey", testFeature.Key).Return(testFeature, nil)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)

	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: s.mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  &testVariation,
		Source:     decision.FeatureTest,
	}

	s.mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		EventProcessor:  s.mockEventProcessor,
		logger:          logging.GetLogger("", ""),
	}
	result, _ := client.IsFeatureEnabled(testFeature.Key, testUserContext)
	s.True(result)
	s.mockConfig.AssertExpectations(s.T())
	s.mockConfigManager.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
}

func (s *ClientTestSuiteFM) TestIsFeatureEnabledWithNotification() {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test happy path
	testVariation := makeTestVariation("green", true)
	testExperiment := makeTestExperimentWithVariations("number_1", []entities.Variation{testVariation})
	testFeature := makeTestFeatureWithExperiment("feature_1", testExperiment)
	s.mockConfig.On("GetFeatureByKey", testFeature.Key).Return(testFeature, nil)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)

	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: s.mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  &testVariation,
		Source:     decision.FeatureTest,
	}

	s.mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	notificationCenter := notification.NewNotificationCenter()
	client := OptimizelyClient{
		ConfigManager:      s.mockConfigManager,
		DecisionService:    s.mockDecisionService,
		logger:             logging.GetLogger("", ""),
		notificationCenter: notificationCenter,
	}
	var numberOfCalls = 0
	note := notification.DecisionNotification{}
	callback := func(notification notification.DecisionNotification) {
		note = notification
		numberOfCalls++
	}
	s.mockDecisionService.On("OnDecision", mock.AnythingOfType("func(notification.DecisionNotification)")).Return(1, nil)
	s.mockDecisionService.notificationCenter = notificationCenter
	id, _ := s.mockDecisionService.OnDecision(callback)

	s.NotEqual(id, 0)
	client.IsFeatureEnabled(testFeature.Key, testUserContext)

	decisionInfo := map[string]interface{}{"feature": map[string]interface{}{"featureEnabled": true, "featureKey": "feature_1",
		"source": decision.FeatureTest, "sourceInfo": map[string]string{"experimentKey": "number_1", "variationKey": "green"}}}
	s.Equal(numberOfCalls, 1)
	s.Equal(decisionInfo, note.DecisionInfo)

	s.mockConfig.AssertExpectations(s.T())
	s.mockConfigManager.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
}

func (s *ClientTestSuiteFM) TestIsFeatureEnabledWithDecisionError() {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test happy path
	testVariation := makeTestVariation("green", true)
	testExperiment := makeTestExperimentWithVariations("number_1", []entities.Variation{testVariation})
	testFeature := makeTestFeatureWithExperiment("feature_1", testExperiment)
	s.mockConfig.On("GetFeatureByKey", testFeature.Key).Return(testFeature, nil)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)

	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: s.mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  &testVariation,
		Source:     decision.FeatureTest,
	}

	s.mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, errors.New(""))
	s.mockEventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent"))

	client := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		EventProcessor:  s.mockEventProcessor,
		logger:          logging.GetLogger("", ""),
	}

	// should still return the decision because the error is non-fatal
	result, err := client.IsFeatureEnabled(testFeature.Key, testUserContext)
	s.True(result)
	s.NoError(err)
	s.mockConfig.AssertExpectations(s.T())
	s.mockConfigManager.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
}

func (s *ClientTestSuiteFM) TestIsFeatureEnabledErrorCases() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"

	// Test instance invalid
	s.mockConfigManager.On("GetConfig").Return(nil, errors.New("no project config available"))

	client := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}
	result, _ := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	s.False(result)
	s.mockDecisionService.AssertNotCalled(s.T(), "GetFeatureDecision")

	// Test invalid feature key
	expectedError := errors.New("Invalid feature key")
	s.mockConfig.On("GetFeatureByKey", testFeatureKey).Return(entities.Feature{}, expectedError)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)

	client = OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}
	result, err := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	s.NoError(err)
	s.False(result)
	s.mockConfigManager.AssertExpectations(s.T())
	s.mockDecisionService.AssertNotCalled(s.T(), "GetDecision")
}

func (s *ClientTestSuiteFM) TestIsFeatureEnabledPanic() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"

	client := OptimizelyClient{
		ConfigManager: &PanickingConfigManager{},
		logger:        logging.GetLogger("", ""),
	}

	// ensure that the client calms back down and recovers
	result, err := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	s.False(result)
	s.Error(err)
}

func (s *ClientTestSuiteFM) TestGetEnabledFeatures() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationEnabled := makeTestVariation("a", true)
	testVariationDisabled := makeTestVariation("b", false)
	testExperimentEnabled := makeTestExperimentWithVariations("enabled_exp", []entities.Variation{testVariationEnabled})
	testExperimentDisabled := makeTestExperimentWithVariations("disabled_exp", []entities.Variation{testVariationDisabled})
	testFeatureEnabled := makeTestFeatureWithExperiment("enabled_feat", testExperimentEnabled)
	testFeatureDisabled := makeTestFeatureWithExperiment("disabled_feat", testExperimentDisabled)

	featureList := []entities.Feature{testFeatureEnabled, testFeatureDisabled}
	s.mockConfig.On("GetFeatureByKey", testFeatureEnabled.Key).Return(testFeatureEnabled, nil)
	s.mockConfig.On("GetFeatureByKey", testFeatureDisabled.Key).Return(testFeatureDisabled, nil)
	s.mockConfig.On("GetFeatureList").Return(featureList)
	s.mockConfigManager.On("GetConfig").Return(s.mockConfig, nil)

	testDecisionContextEnabled := decision.FeatureDecisionContext{
		Feature:       &testFeatureEnabled,
		ProjectConfig: s.mockConfig,
	}
	testDecisionContextDisabled := decision.FeatureDecisionContext{
		Feature:       &testFeatureDisabled,
		ProjectConfig: s.mockConfig,
	}

	expectedFeatureDecisionEnabled := decision.FeatureDecision{
		Experiment: testExperimentEnabled,
		Variation:  &testVariationEnabled,
	}
	expectedFeatureDecisionDisabled := decision.FeatureDecision{
		Experiment: testExperimentDisabled,
		Variation:  &testVariationDisabled,
	}

	s.mockDecisionService.On("GetFeatureDecision", testDecisionContextEnabled, testUserContext).Return(expectedFeatureDecisionEnabled, nil)
	s.mockDecisionService.On("GetFeatureDecision", testDecisionContextDisabled, testUserContext).Return(expectedFeatureDecisionDisabled, nil)

	client := OptimizelyClient{
		ConfigManager:   s.mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}
	result, err := client.GetEnabledFeatures(testUserContext)
	s.NoError(err)
	s.ElementsMatch(result, []string{testFeatureEnabled.Key})
	s.mockConfig.AssertExpectations(s.T())
	s.mockConfigManager.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
}

func (s *ClientTestSuiteFM) TestGetEnabledFeaturesErrorCases() {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test instance invalid
	expectedError := errors.New("no project config available")
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(s.mockConfig, expectedError)

	client := OptimizelyClient{
		ConfigManager:   mockConfigManager,
		DecisionService: s.mockDecisionService,
		logger:          logging.GetLogger("", ""),
	}
	result, err := client.GetEnabledFeatures(testUserContext)
	s.Error(err)
	s.Equal(expectedError, err)
	s.Empty(result)
	mockConfigManager.AssertNotCalled(s.T(), "GetFeatureByKey")
	s.mockDecisionService.AssertNotCalled(s.T(), "GetFeatureDecision")
}

func TestClose(t *testing.T) {
	mockProcessor := &MockProcessor{}
	mockDecisionService := new(MockDecisionService)

	eg := utils.NewExecGroup(context.Background(), logging.GetLogger("", "ExecGroup"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	eg.Go(func(ctx context.Context) {
		<-ctx.Done()
		wg.Done()
	})

	client := OptimizelyClient{
		ConfigManager:   ValidProjectConfigManager(),
		DecisionService: mockDecisionService,
		EventProcessor:  mockProcessor,
		execGroup:       eg,
		logger:          logging.GetLogger("", ""),
	}

	client.Close()
	wg.Wait()
}

type ClientTestSuiteTrackEvent struct {
	suite.Suite
	mockProcessor       *MockProcessor
	mockDecisionService *MockDecisionService
	client              OptimizelyClient
}

func (s *ClientTestSuiteTrackEvent) SetupTest() {
	mockProcessor := new(MockProcessor)
	s.mockProcessor = mockProcessor
	s.mockDecisionService = new(MockDecisionService)

	s.client = OptimizelyClient{
		ConfigManager:      ValidProjectConfigManager(),
		DecisionService:    s.mockDecisionService,
		EventProcessor:     s.mockProcessor,
		notificationCenter: notification.NewNotificationCenter(),
		logger:             logging.GetLogger("", ""),
	}
}

func (s *ClientTestSuiteTrackEvent) TestTrackWithNotification() {

	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	expectedUserContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
		s.Equal("sample_conversion", eventKey)
		s.Equal(expectedUserContext, userContext)
		s.Equal(*s.mockProcessor.Events[0].Conversion, conversionEvent)
	}

	id, err := s.client.OnTrack(onTrack)
	s.Equal(1, id)
	s.NoError(err)

	err = s.client.Track("sample_conversion", expectedUserContext, map[string]interface{}{})
	s.NoError(err)
	s.True(isTrackCalled)
	s.Equal(1, len(s.mockProcessor.Events))
	s.Equal("1212121", s.mockProcessor.Events[0].VisitorID)
	s.Equal("15389410617", s.mockProcessor.Events[0].EventContext.ProjectID)
}

func (s *ClientTestSuiteTrackEvent) TestTrackWithNotificationAndEventTag() {

	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	expectedUserContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}
	expectedEvenTags := map[string]interface{}{
		"client":  "ios",
		"version": "7.0",
	}
	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
		s.Equal("sample_conversion", eventKey)
		s.Equal(expectedUserContext, userContext)
		s.Equal(expectedEvenTags, eventTags)
		s.Equal(*s.mockProcessor.Events[0].Conversion, conversionEvent)
	}

	s.client.OnTrack(onTrack)
	err := s.client.Track("sample_conversion", expectedUserContext, expectedEvenTags)

	s.NoError(err)
	s.True(isTrackCalled)
	s.Equal(1, len(s.mockProcessor.Events))
	s.Equal("1212121", s.mockProcessor.Events[0].VisitorID)
	s.Equal("15389410617", s.mockProcessor.Events[0].EventContext.ProjectID)
}

func (s *ClientTestSuiteTrackEvent) TestTrackWithNotificationAndUserEvent() {

	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	expectedUserContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}
	expectedEventTags := map[string]interface{}{
		"client":  "ios",
		"version": "7.0",
	}
	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
		s.Equal("sample_conversion", eventKey)
		s.Equal(expectedUserContext, userContext)
		s.Equal(expectedEventTags, eventTags)
		s.Equal(1, len(s.mockProcessor.Events))
		s.Equal(*s.mockProcessor.Events[0].Conversion, conversionEvent)
	}

	s.client.OnTrack(onTrack)
	err := s.client.Track("sample_conversion", expectedUserContext, expectedEventTags)

	s.NoError(err)
	s.True(isTrackCalled)
	s.Equal(1, len(s.mockProcessor.Events))
	s.Equal("1212121", s.mockProcessor.Events[0].VisitorID)
	s.Equal("15389410617", s.mockProcessor.Events[0].EventContext.ProjectID)
}

func (s *ClientTestSuiteTrackEvent) TestTrackNotificationNotCalledWhenEventProcessorReturnsFalse() {
	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(false)

	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
	}

	s.client.OnTrack(onTrack)
	err := s.client.Track("sample_conversion", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})
	s.NoError(err)
	s.Equal(0, len(s.mockProcessor.Events))
	s.False(isTrackCalled)
	s.mockProcessor.AssertExpectations(s.T())
}

func (s *ClientTestSuiteTrackEvent) TestTrackNotificationNotCalledWhenNoNotificationCenterProvided() {

	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
	}
	s.client.notificationCenter = nil
	s.client.OnTrack(onTrack)
	err := s.client.Track("sample_conversion", userContext, map[string]interface{}{})

	s.NoError(err)
	s.False(isTrackCalled)
}

func (s *ClientTestSuiteTrackEvent) TestTrackNotificationNotCalledWhenInvalidEventKeyProvided() {

	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
	}

	s.client.OnTrack(onTrack)
	err := s.client.Track("bob", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	s.NoError(err)
	s.Equal(0, len(s.mockProcessor.Events))
	s.False(isTrackCalled)
}

func (s *ClientTestSuiteTrackEvent) TestTrackNotificationNotCalledWhenSendThrowsError() {

	s.mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	expectedUserContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	isTrackCalled := false
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
		isTrackCalled = true
	}

	mockNotificationCenter := new(MockNotificationCenter)
	config, err := s.client.getProjectConfig()
	s.NoError(err)
	configEvent, err := config.GetEventByKey("sample_conversion")
	s.NoError(err)
	userEvent := event.CreateConversionUserEvent(config, configEvent, expectedUserContext, map[string]interface{}{})
	expectedTrackNotification := notification.TrackNotification{EventKey: "sample_conversion", UserContext: expectedUserContext, EventTags: map[string]interface{}{}, ConversionEvent: *userEvent.Conversion}

	mockNotificationCenter.On("Send", notification.Track, expectedTrackNotification).Return(fmt.Errorf(""))
	mockNotificationCenter.On("AddHandler", notification.Track, mock.AnythingOfType("func(interface {})")).Return(1, nil)
	s.client.notificationCenter = mockNotificationCenter
	s.client.OnTrack(onTrack)
	err = s.client.Track("sample_conversion", expectedUserContext, map[string]interface{}{})

	s.NoError(err)
	s.Equal(1, len(s.mockProcessor.Events))
	s.False(isTrackCalled)
	mockNotificationCenter.AssertExpectations(s.T())
}

type ClientTestSuiteTrackNotification struct {
	suite.Suite
	mockProcessor       *MockProcessor
	mockDecisionService *MockDecisionService
	client              OptimizelyClient
}

func (s *ClientTestSuiteTrackNotification) SetupTest() {
	mockProcessor := new(MockProcessor)
	mockProcessor.On("ProcessEvent", mock.AnythingOfType("UserEvent")).Return(true)
	s.mockProcessor = mockProcessor
	s.mockDecisionService = new(MockDecisionService)
	s.client = OptimizelyClient{
		ConfigManager:      ValidProjectConfigManager(),
		DecisionService:    s.mockDecisionService,
		EventProcessor:     s.mockProcessor,
		notificationCenter: notification.NewNotificationCenter(),
		logger:             logging.GetLogger("", ""),
	}
}

func (s *ClientTestSuiteTrackNotification) TestMultipleOnTrack() {

	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	numberOfCalls := 0
	addOnTrack := func(count int) {
		for i := 1; i <= count; i++ {
			onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
				numberOfCalls++
			}
			id, err := s.client.OnTrack(onTrack)
			// To check if id's are unique and increasing
			s.Equal(i, id)
			s.NoError(err)
		}
	}

	// Add 5 on track callbacks
	addOnTrack(5)
	err := s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(5, numberOfCalls)
}

func (s *ClientTestSuiteTrackNotification) TestMultipleRemoveOnTrack() {

	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	numberOfCalls := 0
	callbackIds := []int{}

	// Add 5 on track callbacks
	for i := 1; i <= 5; i++ {
		onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
			numberOfCalls++
		}
		id, err := s.client.OnTrack(onTrack)
		callbackIds = append(callbackIds, id)
		s.Equal(i, id)
		s.NoError(err)
	}

	err := s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(5, numberOfCalls)

	// Remove all track callbacks
	numberOfCalls = 0
	for i := 0; i < 5; i++ {
		err = s.client.RemoveOnTrack(callbackIds[i])
		s.NoError(err)
	}

	err = s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(0, numberOfCalls)
}

func (s *ClientTestSuiteTrackNotification) TestOnTrackAfterRemoveOnTrack() {

	userContext := entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}

	numberOfCalls := 0
	callbackIds := []int{}
	addOnTrack := func(count int) {
		for i := 0; i < count; i++ {
			onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
				numberOfCalls++
			}
			id, err := s.client.OnTrack(onTrack)
			callbackIds = append(callbackIds, id)
			s.NoError(err)
		}
	}

	// Add 5 on track callbacks
	addOnTrack(5)
	err := s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(5, numberOfCalls)

	// Remove all track callbacks
	numberOfCalls = 0
	for i := 0; i < 5; i++ {
		err = s.client.RemoveOnTrack(callbackIds[i])
		s.NoError(err)
	}
	err = s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(0, numberOfCalls)

	// Add 2 on track callbacks
	addOnTrack(2)
	err = s.client.Track("sample_conversion", userContext, map[string]interface{}{})
	s.NoError(err)
	s.Equal(2, numberOfCalls)
}

func (s *ClientTestSuiteTrackNotification) TestOnTrackThrowsErrorWithoutNotificationCenter() {

	s.client.notificationCenter = nil
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	}
	id, err := s.client.OnTrack(onTrack)
	s.Equal(0, id)
	s.Error(err)
}

func (s *ClientTestSuiteTrackNotification) TestRemoveOnTrackThrowsErrorWithoutNotificationCenter() {

	s.client.notificationCenter = nil
	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	}
	id, _ := s.client.OnTrack(onTrack)
	err := s.client.RemoveOnTrack(id)
	s.Error(err)
}

func (s *ClientTestSuiteTrackNotification) TestOnTrackThrowsErrorWhenAddHandlerFails() {

	mockNotificationCenter := new(MockNotificationCenter)
	mockNotificationCenter.On("AddHandler", notification.Track, mock.AnythingOfType("func(interface {})")).Return(-1, fmt.Errorf(""))
	s.client.notificationCenter = mockNotificationCenter

	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	}
	id, err := s.client.OnTrack(onTrack)
	s.Equal(0, id)
	s.Error(err)
	mockNotificationCenter.AssertExpectations(s.T())
}

func (s *ClientTestSuiteTrackNotification) TestRemoveOnTrackThrowsErrorWhenRemoveHandlerFails() {

	mockNotificationCenter := new(MockNotificationCenter)
	mockNotificationCenter.On("AddHandler", notification.Track, mock.AnythingOfType("func(interface {})")).Return(1, nil)
	mockNotificationCenter.On("RemoveHandler", 1, notification.Track).Return(fmt.Errorf(""))
	s.client.notificationCenter = mockNotificationCenter

	onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	}
	id, err := s.client.OnTrack(onTrack)
	s.Equal(1, id)
	s.NoError(err)

	err = s.client.RemoveOnTrack(id)
	s.Error(err)
	mockNotificationCenter.AssertExpectations(s.T())
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuiteAB))
	suite.Run(t, new(ClientTestSuiteFM))
	suite.Run(t, new(ClientTestSuiteTrackEvent))
	suite.Run(t, new(ClientTestSuiteTrackNotification))
}
