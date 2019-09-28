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

package client

import (
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockProjectConfigManager struct {
	projectConfig optimizely.ProjectConfig
	mock.Mock
}

func (p *MockProjectConfigManager) GetConfig() (optimizely.ProjectConfig, error) {
	if p.projectConfig != nil {
		return p.projectConfig, nil
	}

	args := p.Called()
	return args.Get(0).(optimizely.ProjectConfig), args.Error(1)
}

func ValidProjectConfigManager() *MockProjectConfigManager {
	p := new(MockProjectConfigManager)
	p.projectConfig = new(TestConfig)
	return p
}

type MockProcessor struct {
	Events []event.UserEvent
}

func (f *MockProcessor) ProcessEvent(event event.UserEvent) {
	f.Events = append(f.Events, event)
}

type TestConfig struct {
	optimizely.ProjectConfig
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
	mockProcessor := &MockProcessor{}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   ValidProjectConfigManager(),
		decisionService: mockDecisionService,
		eventProcessor:  mockProcessor,
	}

	err := client.Track("sample_conversion", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	assert.Nil(t, err)
	assert.True(t, len(mockProcessor.Events) == 1)
	assert.True(t, mockProcessor.Events[0].VisitorID == "1212121")
	assert.True(t, mockProcessor.Events[0].EventContext.ProjectID == "15389410617")

}

func TestTrackFail(t *testing.T) {
	mockProcessor := &MockProcessor{}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   ValidProjectConfigManager(),
		decisionService: mockDecisionService,
		eventProcessor:  mockProcessor,
	}

	err := client.Track("bob", entities.UserContext{ID: "1212121", Attributes: map[string]interface{}{}}, map[string]interface{}{})

	assert.Error(t, err)
	assert.True(t, len(mockProcessor.Events) == 0)

}

func TestIsFeatureEnabled(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariation := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
	}
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeatureKey := "test_feature_key"
	testFeature := entities.Feature{
		ID:                 "22222",
		Key:                testFeatureKey,
		FeatureExperiments: []entities.Experiment{testExperiment},
	}
	// Test happy path
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  &testVariation,
	}

	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	assert.True(t, result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestIsFeatureEnabledErrorCases(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"

	// Test instance invalid
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(nil, errors.New("no project config available"))
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	assert.False(t, result)
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")

	// Test invalid feature key
	expectedError := errors.New("Invalid feature key")
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(entities.Feature{}, expectedError)

	mockConfigManager = new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	mockDecisionService = new(MockDecisionService)
	client = OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	if assert.Error(t, err) {
		assert.Equal(t, expectedError, err)
	}
	assert.False(t, result)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertNotCalled(t, "GetDecision")
}

func TestIsFeatureEnabledPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	assert.False(t, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetEnabledFeatures(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationEnabled := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: true,
	}
	testVariationDisabled := entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: false,
	}
	testExperimentEnabled := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariationEnabled},
	}
	testExperimentDisabled := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariationDisabled},
	}
	testFeatureEnabledKey := "test_feature_enabled_key"
	testFeatureEnabled := entities.Feature{
		ID:                 "22222",
		Key:                testFeatureEnabledKey,
		FeatureExperiments: []entities.Experiment{testExperimentEnabled},
	}
	testFeatureDisabledKey := "test_feature_disabled_key"
	testFeatureDisabled := entities.Feature{
		ID:                 "22222",
		Key:                testFeatureDisabledKey,
		FeatureExperiments: []entities.Experiment{testExperimentDisabled},
	}
	featureList := []entities.Feature{testFeatureEnabled, testFeatureDisabled}
	// Test happy path
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", testFeatureEnabledKey).Return(testFeatureEnabled, nil)
	mockConfig.On("GetFeatureByKey", testFeatureDisabledKey).Return(testFeatureDisabled, nil)
	mockConfig.On("GetFeatureList").Return(featureList)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)
	// Set up the mock decision service and its return value
	testDecisionContextEnabled := decision.FeatureDecisionContext{
		Feature:       &testFeatureEnabled,
		ProjectConfig: mockConfig,
	}
	testDecisionContextDisabled := decision.FeatureDecisionContext{
		Feature:       &testFeatureDisabled,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecisionEnabled := decision.FeatureDecision{
		Experiment: testExperimentEnabled,
		Variation:  &testVariationEnabled,
	}
	expectedFeatureDecisionDisabled := decision.FeatureDecision{
		Experiment: testExperimentDisabled,
		Variation:  &testVariationDisabled,
	}

	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContextEnabled, testUserContext).Return(expectedFeatureDecisionEnabled, nil)
	mockDecisionService.On("GetFeatureDecision", testDecisionContextDisabled, testUserContext).Return(expectedFeatureDecisionDisabled, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.NoError(t, err)
	assert.ElementsMatch(t, result, []string{testFeatureEnabledKey})
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetEnabledFeaturesErrorCases(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	// Test instance invalid
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(nil, errors.New("no project config available"))
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.Error(t, err)
	assert.Empty(t, result)
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")
}

func TestGetEnabledFeaturesPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.Empty(t, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableBooleanWithValidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "false",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Boolean,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, true, result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableBooleanWithInvalidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "stringvalue"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "false",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Boolean,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableBooleanWithInvalidValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Integer,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableBooleanWithEmptyValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "",
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableBooleanReturnsDefaultValueIfFeatureNotEnabled(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "false",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Boolean,
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.Nil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableBoolPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableBoolean(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, false, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableDoubleWithValidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Double,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(5), result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableDoubleWithInvalidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "stringvalue"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Double,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(0), result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableDoubleWithInvalidValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Integer,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(0), result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableDoubleWithEmptyValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "",
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(0), result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableDoubleReturnsDefaultValueIfFeatureNotEnabled(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Double,
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(4), result)
	assert.Nil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableDoublePanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableDouble(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, float64(0), result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableIntegerWithValidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Integer,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 5, result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableIntegerWithInvalidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "stringvalue"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Integer,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 0, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableIntegerWithInvalidValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "false",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Boolean,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 0, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableIntegerWithEmptyValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "false",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "",
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 0, result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableIntegerReturnsDefaultValueIfFeatureNotEnabled(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "5"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "4",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Integer,
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 4, result)
	assert.Nil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableIntegerPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableInteger(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, 0, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableStringWithValidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "default",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.String,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, _ := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, testVariableValue, result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringWithInvalidValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "default",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         entities.Boolean,
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "", result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringWithEmptyValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := entities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := entities.Variable{
		DefaultValue: "default",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "",
	}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "", result)
	assert.Error(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringReturnsDefaultValueIfFeatureNotEnabled(t *testing.T) {
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "defaultString", result)
	assert.Nil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "", result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableErrorCases(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(nil, errors.New("no project config available"))
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}
	_, err1 := client.GetFeatureVariableBoolean("test_feature_key", "test_variable_key", testUserContext)
	_, err2 := client.GetFeatureVariableDouble("test_feature_key", "test_variable_key", testUserContext)
	_, err3 := client.GetFeatureVariableInteger("test_feature_key", "test_variable_key", testUserContext)
	_, err4 := client.GetFeatureVariableString("test_feature_key", "test_variable_key", testUserContext)
	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Error(t, err3)
	assert.Error(t, err4)
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockConfigManager.AssertNotCalled(t, "GetVariableByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")
}

func TestGetProjectConfigIsValid(t *testing.T) {
	mockConfigManager := ValidProjectConfigManager()

	client := OptimizelyClient{
		configManager: mockConfigManager,
	}

	actual, err := client.GetProjectConfig()

	assert.Nil(t, err)
	assert.Equal(t, mockConfigManager.projectConfig, actual)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}

	_, featureDecision, err := client.getFeatureDecision(testFeatureKey, testUserContext)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testUserContext)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)

	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   &PanickingConfigManager{},
		decisionService: mockDecisionService,
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testUserContext)
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
		configManager:   mockConfigManager,
		decisionService: &PanickingDecisionService{},
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testUserContext)
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
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, errors.New("error feaure"))

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}

	_, _, err := client.getFeatureDecision(testFeatureKey, testUserContext)
	assert.Nil(t, err)
}

func TestGetAllFeatureVariables(t *testing.T) {
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
	testVariation.FeatureEnabled = true
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeature.Variables = make([]entities.Variable, 1)
	testFeature.Variables[0] = testVariable
	mockConfig := getMockConfig(testFeatureKey, testVariableKey, testFeature, testVariable)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig, nil)

	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := getTestFeatureDecision(testExperiment, testVariation, true)
	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
	}

	enabled, variationMap, err := client.GetAllFeatureVariables(testFeatureKey, testUserContext)
	assert.True(t, enabled)
	assert.Equal(t, testVariableValue, variationMap[testVariableKey])
	assert.Nil(t, err)
}

// Helper Methods
func getTestFeatureDecision(experiment entities.Experiment, variation entities.Variation, decisionMade bool) decision.FeatureDecision {
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
	testExperiment := getTestExperiment("test_exp_1")
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
	s.mockEventProcessor.On("ProcessEvent", mock.AnythingOfType("event.UserEvent"))

	testClient := OptimizelyClient{
		configManager:   s.mockConfigManager,
		decisionService: s.mockDecisionService,
		eventProcessor:  s.mockEventProcessor,
	}

	variationKey, err := testClient.Activate("test_exp_1", testUserContext)
	s.NoError(err)
	s.Equal(expectedVariation.Key, variationKey)
	s.mockConfig.AssertExpectations(s.T())
	s.mockDecisionService.AssertExpectations(s.T())
	s.mockEventProcessor.AssertExpectations(s.T())
}

func (s *ClientTestSuiteAB) TestGetVariation() {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testExperiment := getTestExperiment("test_exp_1")
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
		configManager:   s.mockConfigManager,
		decisionService: s.mockDecisionService,
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
		configManager:   new(PanickingConfigManager),
		decisionService: s.mockDecisionService,
	}

	variationKey, err := testClient.GetVariation("test_exp_1", testUserContext)
	s.Equal("", variationKey)
	s.EqualError(err, "I'm panicking")
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuiteAB))
}
