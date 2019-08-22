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

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/optimizely"
	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
)

type MockProjectConfig struct {
	optimizely.ProjectConfig
	mock.Mock
}

func (c *MockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := c.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (c *MockProjectConfig) GetFeatureList() []entities.Feature {
	args := c.Called()
	return args.Get(0).([]entities.Feature)
}

func (c *MockProjectConfig) GetFeatureFlagByKey(featureKey string) (datafileEntities.FeatureFlag, error) {
	args := c.Called(featureKey)
	return args.Get(0).(datafileEntities.FeatureFlag), args.Error(1)
}

type MockProjectConfigManager struct {
	mock.Mock
}

func (p *MockProjectConfigManager) GetConfig() optimizely.ProjectConfig {
	args := p.Called()
	return args.Get(0).(optimizely.ProjectConfig)
}

type MockDecisionService struct {
	decision.DecisionService
	mock.Mock
}

func (m *MockDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext) (decision.FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(decision.FeatureDecision), args.Error(1)
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
	mockConfigManager.On("GetConfig").Return(mockConfig)
	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature:       &testFeature,
		ProjectConfig: mockConfig,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		Experiment: testExperiment,
		Variation:  testVariation,
		Decision: decision.Decision{
			DecisionMade: true,
		},
	}

	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContext, testUserContext).Return(expectedFeatureDecision, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
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
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         false,
	}
	result, _ := client.IsFeatureEnabled(testFeatureKey, testUserContext)
	assert.False(t, result)
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")

	// Test invalid feature key
	expectedError := errors.New("Invalid feature key")
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(entities.Feature{}, expectedError)

	mockConfigManager = new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig)
	mockDecisionService = new(MockDecisionService)
	client = OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
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

	mockConfigManager := new(MockProjectConfigManager)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
	}

	// returning an error object will cause the Client to panic
	mockConfigManager.On("GetFeatureByKey", testFeatureKey, testUserContext).Return(errors.New("failure"))

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
	mockConfigManager.On("GetConfig").Return(mockConfig)
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
		Variation:  testVariationEnabled,
		Decision: decision.Decision{
			DecisionMade: true,
		},
	}
	expectedFeatureDecisionDisabled := decision.FeatureDecision{
		Experiment: testExperimentDisabled,
		Variation:  testVariationDisabled,
		Decision: decision.Decision{
			DecisionMade: true,
		},
	}

	mockDecisionService := new(MockDecisionService)
	mockDecisionService.On("GetFeatureDecision", testDecisionContextEnabled, testUserContext).Return(expectedFeatureDecisionEnabled, nil)
	mockDecisionService.On("GetFeatureDecision", testDecisionContextDisabled, testUserContext).Return(expectedFeatureDecisionDisabled, nil)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
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
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         false,
	}
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.Error(t, err)
	assert.Empty(t, result)
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")
}

func TestGetEnabledFeaturesPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"

	mockConfigManager := new(MockProjectConfigManager)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
	}

	// returning an error object will cause the Client to panic
	mockConfigManager.On("GetFeatureByKey", testFeatureKey, testUserContext).Return(errors.New("failure"))

	// ensure that the client calms back down and recovers
	result, err := client.GetEnabledFeatures(testUserContext)
	assert.Empty(t, result)
	assert.True(t, assert.Error(t, err))
}

func TestGetFeatureVariableStringWithValidValue(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testFeatureFlagKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := datafileEntities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := datafileEntities.Variable{
		DefaultValue: "default",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "string"}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeatureFlag := getTestFeatureFlag(testFeatureFlagKey, testVariable)
	mockConfig := getMockConfig(testFeatureKey, testFeature, testFeatureFlag)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig)

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
		isValid:         true,
	}
	result, _ := client.GetFeatureVariableString(testFeatureKey, testFeatureFlagKey, testUserContext)
	assert.Equal(t, testVariableValue, result)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringWithInvalidValueType(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testFeatureFlagKey := "test_feature_flag_key"
	testVariableValue := "true"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := datafileEntities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := datafileEntities.Variable{
		DefaultValue: "default",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "boolean"}
	testVariation := getTestVariationWithFeatureVariable(true, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeatureFlag := getTestFeatureFlag(testFeatureFlagKey, testVariable)
	mockConfig := getMockConfig(testFeatureKey, testFeature, testFeatureFlag)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig)

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
		isValid:         true,
	}
	result, err := client.GetFeatureVariableString(testFeatureKey, testFeatureFlagKey, testUserContext)
	assert.Equal(t, "", result)
	assert.NotNil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableStringReturnsDefaultValueIfFeatureNotEnabled(t *testing.T) {
	testFeatureKey := "test_feature_key"
	testFeatureFlagKey := "test_feature_flag_key"
	testVariableValue := "teststring"
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testVariationVariable := datafileEntities.VariationVariable{
		ID:    "1",
		Value: testVariableValue,
	}
	testVariable := datafileEntities.Variable{
		DefaultValue: "defaultString",
		ID:           "1",
		Key:          "test_feature_flag_key",
		Type:         "string"}
	testVariation := getTestVariationWithFeatureVariable(false, testVariationVariable)
	testExperiment := entities.Experiment{
		ID:         "111111",
		Variations: map[string]entities.Variation{"22222": testVariation},
	}
	testFeature := getTestFeature(testFeatureKey, testExperiment)
	testFeatureFlag := getTestFeatureFlag(testFeatureFlagKey, testVariable)
	mockConfig := getMockConfig(testFeatureKey, testFeature, testFeatureFlag)
	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig)

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
		isValid:         true,
	}
	result, err := client.GetFeatureVariableString(testFeatureKey, testFeatureFlagKey, testUserContext)
	assert.Equal(t, "defaultString", result)
	assert.Nil(t, err)
	mockConfig.AssertExpectations(t)
	mockConfigManager.AssertExpectations(t)
	mockDecisionService.AssertExpectations(t)
}

func TestGetFeatureVariableErrorCases(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}

	mockConfigManager := new(MockProjectConfigManager)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         false,
	}
	_, err1 := client.GetFeatureVariableString("test_feature_key", "test_variable_key", testUserContext)
	assert.Error(t, err1)
	mockConfigManager.AssertNotCalled(t, "GetFeatureFlagByKey")
	mockConfigManager.AssertNotCalled(t, "GetFeatureByKey")
	mockDecisionService.AssertNotCalled(t, "GetFeatureDecision")
}

func TestGetFeatureVariableStringPanic(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testVariableKey := "test_variable_key"

	mockConfigManager := new(MockProjectConfigManager)
	mockDecisionService := new(MockDecisionService)

	client := OptimizelyClient{
		configManager:   mockConfigManager,
		decisionService: mockDecisionService,
		isValid:         true,
	}

	// returning an error object will cause the Client to panic
	mockConfigManager.On("GetFeatureByKey", testFeatureKey, testUserContext).Return(errors.New("failure"))

	// ensure that the client calms back down and recovers
	result, err := client.GetFeatureVariableString(testFeatureKey, testVariableKey, testUserContext)
	assert.Equal(t, "", result)
	assert.True(t, assert.Error(t, err))
}

// Helper Methods
func getTestFeatureDecision(experiment entities.Experiment, variation entities.Variation, decisionMade bool) decision.FeatureDecision {
	return decision.FeatureDecision{
		Experiment: experiment,
		Variation:  variation,
		Decision: decision.Decision{
			DecisionMade: decisionMade,
		},
	}
}

func getTestVariationWithFeatureVariable(featureEnabled bool, featureVariable datafileEntities.VariationVariable) entities.Variation {
	return entities.Variation{
		ID:             "22222",
		Key:            "22222",
		FeatureEnabled: featureEnabled,
		Variables:      []datafileEntities.VariationVariable{featureVariable},
	}
}

func getTestFeatureFlag(featureFlagKey string, variable datafileEntities.Variable) datafileEntities.FeatureFlag {
	return datafileEntities.FeatureFlag{
		ID:            "21111",
		Key:           featureFlagKey,
		RolloutID:     "41111",
		ExperimentIDs: []string{"31111", "31112"},
		Variables:     []datafileEntities.Variable{variable},
	}
}

func getMockConfig(featureKey string, feature entities.Feature, featureFlag datafileEntities.FeatureFlag) *MockProjectConfig {
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", featureKey).Return(feature, nil)
	mockConfig.On("GetFeatureFlagByKey", featureKey).Return(featureFlag, nil)
	return mockConfig
}

func getTestFeature(featureKey string, experiment entities.Experiment) entities.Feature {
	return entities.Feature{
		ID:                 "22222",
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{experiment},
	}
}
