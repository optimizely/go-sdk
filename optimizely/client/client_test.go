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

	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
)

type MockProjectConfig struct {
	config.ProjectConfig
	mock.Mock
}

func (c *MockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := c.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

type MockProjectConfigManager struct {
	mock.Mock
}

func (p *MockProjectConfigManager) GetConfig() config.ProjectConfig {
	args := p.Called()
	return args.Get(0).(config.ProjectConfig)
}

type MockDecisionService struct {
	mock.Mock
}

func (m *MockDecisionService) GetFeatureDecision(decisionContext decision.FeatureDecisionContext, userContext entities.UserContext) (decision.FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(decision.FeatureDecision), args.Error(1)
}

func TestIsFeatureEnabled(t *testing.T) {
	testUserContext := entities.UserContext{ID: "test_user_1"}
	testFeatureKey := "test_feature_key"
	testFeature := entities.Feature{
		Key: testFeatureKey,
		FeatureExperiments: []entities.Experiment{
			entities.Experiment{
				ID: "111111",
				Variations: map[string]entities.Variation{
					"22222": entities.Variation{
						ID:             "22222",
						Key:            "22222",
						FeatureEnabled: true,
					},
				},
			},
		},
	}

	// Test happy path
	mockConfig := new(MockProjectConfig)
	mockConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)

	mockConfigManager := new(MockProjectConfigManager)
	mockConfigManager.On("GetConfig").Return(mockConfig)

	// Set up the mock decision service and its return value
	testDecisionContext := decision.FeatureDecisionContext{
		Feature: testFeature,
	}

	expectedFeatureDecision := decision.FeatureDecision{
		FeatureEnabled: true,
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
}
