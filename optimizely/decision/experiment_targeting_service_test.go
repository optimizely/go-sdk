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

	"github.com/optimizely/go-sdk/optimizely/decision/reasons"

	"github.com/optimizely/go-sdk/optimizely/decision/evaluator"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAudienceEvaluator struct {
	mock.Mock
}

func (m *MockAudienceEvaluator) Evaluate(audience entities.Audience, condTreeParams *entities.TreeParameters) bool {
	userContext := *condTreeParams.User
	args := m.Called(audience, userContext)
	return args.Bool(0)
}

// test with mocking
func TestExperimentTargetingGetDecisionNoAudienceCondTree(t *testing.T) {
	testAudience := entities.Audience{
		ID: "33333",
		ConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes: []*entities.TreeNode{
				&entities.TreeNode{
					Item: entities.Condition{
						Name:  "s_foo",
						Value: "foo",
					},
				},
			},
		},
	}

	// make a copy of the testExp1111 so we do not mutate the original one
	var testExperiment = testExp1111
	testExperiment.AudienceIds = []string{"33333"}
	mockProjectConfig := new(mockProjectConfig)
	mockProjectConfig.On("GetAudienceByID", "33333").Return(testAudience, nil)
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExperiment,
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
			Reason:       reasons.DoesNotQualify,
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

// Real tests with no mocking
func TestExperimentTargetingGetDecisionWithAudienceCondTree(t *testing.T) {
	testAudience := entities.Audience{
		ConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes: []*entities.TreeNode{
				{
					Item: entities.Condition{
						Name:  "s_foo",
						Type:  "custom_attribute",
						Match: "exact",
						Value: "foo",
					},
				},
			},
		},
	}

	testExperimentKey := "test_experiment"
	testExperiment := entities.Experiment{
		ID:          "111111",
		Key:         testExperimentKey,
		AudienceIds: []string{"33333"},
	}

	mockProjectConfig := new(mockProjectConfig)
	mockProjectConfig.On("GetAudienceByID", "33333").Return(testAudience, nil)
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExperiment,
		ProjectConfig: mockProjectConfig,
	}
	// test does not pass audience evaluation
	testUserContext := entities.UserContext{
		ID: "test_user_1",
		Attributes: entities.UserAttributes{
			Attributes: map[string]interface{}{
				"s_foo": "not_foo",
			},
		},
	}
	expectedExperimentDecision := ExperimentDecision{
		Decision: Decision{
			DecisionMade: true,
			Reason:       reasons.DoesNotQualify,
		},
	}

	audienceEvaluator := evaluator.NewTypedAudienceEvaluator()
	experimentTargetingService := ExperimentTargetingService{
		audienceEvaluator: audienceEvaluator,
	}

	decision, _ := experimentTargetingService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedExperimentDecision, decision) //decision made but did not pass

	/****** Perfect Match ***************/

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
		Variation: entities.Variation{},
	}

	decision, _ = experimentTargetingService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedExperimentDecision, decision) // decision not made? but it passed

}
