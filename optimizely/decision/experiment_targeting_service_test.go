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
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type bMockProjectConfig struct {
	optimizely.ProjectConfig
	mock.Mock
}

func (c *bMockProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	args := c.Called(audienceID)
	return args.Get(0).(entities.Audience), args.Error(1)
}

type MockAudienceEvaluator struct {
	mock.Mock
}

func (m *MockAudienceEvaluator) Evaluate(audience entities.Audience, userContext entities.UserContext) bool {
	args := m.Called(audience, userContext)
	return args.Bool(0)
}

func TestExperimentTargetingGetDecision(t *testing.T) {
	testAudience := entities.Audience{
		ConditionTree: &entities.ConditionTreeNode{
			Operator: "or",
			Nodes: []*entities.ConditionTreeNode{
				&entities.ConditionTreeNode{
					Condition: entities.Condition{
						Name:  "s_foo",
						Value: "foo",
					},
				},
			},
		},
	}
	mockProjectConfig := new(bMockProjectConfig)
	mockProjectConfig.On("GetAudienceByID", "33333").Return(testAudience, nil)
	testVariation := entities.Variation{
		ID:  "22222",
		Key: "22222",
	}
	testDecisionContext := ExperimentDecisionContext{
		Experiment: entities.Experiment{
			ID: "111111",
			Variations: map[string]entities.Variation{
				"22222": testVariation,
			},
			AudienceIds: []string{"33333"},
		},
		ProjectConfig: mockProjectConfig,
	}

	// test does not pass audience evaluation
	testUserContext := entities.UserContext{
		ID: "test_user_1",
		Attributes: entities.UserAttributes{
			Attributes: map[string]interface{}{
				"s_foo": "not foo",
			},
		},
	}
	expectedExperimentDecision := ExperimentDecision{
		Decision: Decision{
			DecisionMade: true,
		},
	}

	mockAudienceEvaluator := new(MockAudienceEvaluator)
	mockAudienceEvaluator.On("Evaluate", testAudience, testUserContext).Return(false)
	experimentTargetingService := ExperimentTargetingService{
		audienceEvaluator: mockAudienceEvaluator,
	}
	decision, _ := experimentTargetingService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedExperimentDecision, decision)
	mockAudienceEvaluator.AssertExpectations(t)

	// test passes evaluation, no decision is made
	testUserContext = entities.UserContext{
		ID: "test_user_1",
		Attributes: entities.UserAttributes{
			Attributes: map[string]interface{}{
				"s_foo": "foo",
			},
		},
	}
	expectedExperimentDecision = ExperimentDecision{
		Decision: Decision{
			DecisionMade: false,
		},
	}

	mockAudienceEvaluator = new(MockAudienceEvaluator)
	mockAudienceEvaluator.On("Evaluate", testAudience, testUserContext).Return(true)
	experimentTargetingService = ExperimentTargetingService{
		audienceEvaluator: mockAudienceEvaluator,
	}
	decision, _ = experimentTargetingService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedExperimentDecision, decision)
	mockAudienceEvaluator.AssertExpectations(t)
	mockProjectConfig.AssertExpectations(t)
}
