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

	"github.com/optimizely/go-sdk/optimizely"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

type cMockProjectConfig struct {
	optimizely.ProjectConfig
	mock.Mock
}

func (c *cMockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := c.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

type MockFeatureDecisionService struct {
	mock.Mock
}

func (m *MockFeatureDecisionService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(FeatureDecision), args.Error(1)
}

func TestGetFeatureDecision(t *testing.T) {
	testFeatureKey := "my_test_feature"
	testVariation := entities.Variation{
		ID:             "11111",
		FeatureEnabled: true,
	}
	testExperiment := entities.Experiment{
		Variations: map[string]entities.Variation{"11111": testVariation},
	}
	testFeature := entities.Feature{
		Key:                testFeatureKey,
		FeatureExperiments: []entities.Experiment{testExperiment},
	}
	mockProjectConfig := new(cMockProjectConfig)
	mockProjectConfig.On("GetFeatureByKey", testFeatureKey).Return(testFeature, nil)
	decisionContext := FeatureDecisionContext{
		FeatureKey:    testFeatureKey,
		ProjectConfig: mockProjectConfig,
	}

	userContext := entities.UserContext{
		ID: "test_user",
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: testExperiment,
		Variation:  testVariation,
		Decision:   Decision{DecisionMade: true},
	}

	testFeatureDecisionService := new(MockFeatureDecisionService)
	testFeatureDecisionService.On("GetDecision", decisionContext, userContext).Return(expectedFeatureDecision, nil)

	decisionService := &CompositeService{
		featureDecisionServices: []FeatureDecisionService{testFeatureDecisionService},
	}
	featureDecision, err := decisionService.GetFeatureDecision(decisionContext, userContext)
	if err != nil {
	}

	// Test assertions
	assert.Equal(t, expectedFeatureDecision, featureDecision)
	testFeatureDecisionService.AssertExpectations(t)
}
