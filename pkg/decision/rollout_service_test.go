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

	"github.com/optimizely/go-sdk/pkg/decision/evaluator"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/pkg/entities"
)

func TestRolloutServiceGetDecision(t *testing.T) {
	testUserContext := entities.UserContext{
		ID: "test_user",
	}
	mockProjectConfig := new(mockProjectConfig)
	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &testFeatRollout3334,
		ProjectConfig: mockProjectConfig,
	}
	testAudienceMap := map[string]entities.Audience{
		"5555": testAudience5555,
	}
	mockProjectConfig.On("GetAudienceMap").Return(testAudienceMap)
	testCondTreeParams := entities.NewTreeParameters(&testUserContext, testAudienceMap)

	// Test experiment passes targeting and bucketing
	testExperimentBucketerDecision := ExperimentDecision{
		Variation: &testExp1112Var2222,
	}
	testExperimentBucketerDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}

	testAudienceConditionTree := testExp1112.AudienceConditionTree
	mockAudienceTreeEvaluator := new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", testAudienceConditionTree, testCondTreeParams).Return(true)
	mockExperimentBucketerService := new(MockExperimentDecisionService)
	mockExperimentBucketerService.On("GetDecision", testExperimentBucketerDecisionContext, testUserContext).Return(testExperimentBucketerDecision, nil)
	testRolloutService := RolloutService{
		audienceTreeEvaluator:     mockAudienceTreeEvaluator,
		experimentBucketerService: mockExperimentBucketerService,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1112,
		Variation:  &testExp1112Var2222,
		Source:     Rollout,
	}
	decision, _ := testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.Equal(t, expectedFeatureDecision, decision)
	mockAudienceTreeEvaluator.AssertExpectations(t)
	mockExperimentBucketerService.AssertExpectations(t)

	// Test experiment passes targeting but not bucketing
	testExperimentBucketerDecision = ExperimentDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
	}
	testExperimentBucketerDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}

	mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", testAudienceConditionTree, testCondTreeParams).Return(true)
	mockExperimentBucketerService = new(MockExperimentDecisionService)
	mockExperimentBucketerService.On("GetDecision", testExperimentBucketerDecisionContext, testUserContext).Return(testExperimentBucketerDecision, nil)
	testRolloutService = RolloutService{
		audienceTreeEvaluator:     mockAudienceTreeEvaluator,
		experimentBucketerService: mockExperimentBucketerService,
	}
	expectedFeatureDecision = FeatureDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
		Experiment: testExp1112,
		Source:     Rollout,
	}
	decision, _ = testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.Equal(t, expectedFeatureDecision, decision)
	mockAudienceTreeEvaluator.AssertExpectations(t)
	mockExperimentBucketerService.AssertExpectations(t)

	// Test experiment fails targeting
	mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", testAudienceConditionTree, testCondTreeParams).Return(false)
	testRolloutService = RolloutService{
		audienceTreeEvaluator:     mockAudienceTreeEvaluator,
		experimentBucketerService: mockExperimentBucketerService,
	}
	expectedFeatureDecision = FeatureDecision{
		Decision: Decision{
			Reason: reasons.FailedRolloutTargeting,
		},
	}
	decision, _ = testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.Nil(t, decision.Variation)
	mockAudienceTreeEvaluator.AssertExpectations(t)
	mockExperimentBucketerService.AssertNotCalled(t, "GetDecision")
}

func TestNewRolloutService(t *testing.T) {
	rolloutService := NewRolloutService()
	assert.IsType(t, &evaluator.MixedTreeEvaluator{}, rolloutService.audienceTreeEvaluator)
	assert.IsType(t, &ExperimentBucketerService{}, rolloutService.experimentBucketerService)
}
