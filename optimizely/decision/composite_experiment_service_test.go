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

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

func TestCompositeExperimentServiceGetDecision(t *testing.T) {
	mockProjectConfig := new(mockProjectConfig)
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1111,
		ProjectConfig: mockProjectConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1111.Variations["2222"]
	expectedExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}
	// test that we return out of the decision making and the next one doesn't get called
	mockExperimentDecisionService := new(MockExperimentDecisionService)
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 := new(MockExperimentDecisionService)
	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{mockExperimentDecisionService, mockExperimentDecisionService2},
	}
	decision, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertNotCalled(t, "GetDecision")

	// test that we move on to the next decision service if no decision is made
	mockExperimentDecisionService = new(MockExperimentDecisionService)
	expectedExperimentDecision = ExperimentDecision{}
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 = new(MockExperimentDecisionService)
	expectedExperimentDecision2 := ExperimentDecision{
		Variation: &expectedVariation,
	}
	mockExperimentDecisionService2.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService = &CompositeExperimentService{
		experimentServices: []ExperimentService{mockExperimentDecisionService, mockExperimentDecisionService2},
	}
	decision, err = compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision2, decision)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertExpectations(t)

	// test when no decisions are made
	mockExperimentDecisionService = new(MockExperimentDecisionService)
	expectedExperimentDecision = ExperimentDecision{}
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 = new(MockExperimentDecisionService)
	expectedExperimentDecision2 = ExperimentDecision{}
	mockExperimentDecisionService2.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService = &CompositeExperimentService{
		experimentServices: []ExperimentService{mockExperimentDecisionService, mockExperimentDecisionService2},
	}
	decision, err = compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision2, decision)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertExpectations(t)
}

func TestCompositeExperimentServiceGetDecisionReturnsErrorWhenExperimentNotRunning(t *testing.T) {
	experiment := &testExp1111
	experiment.Status = entities.Paused
	mockProjectConfig := new(mockProjectConfig)
	testDecisionContext := ExperimentDecisionContext{
		Experiment:    experiment,
		ProjectConfig: mockProjectConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedVariation := testExp1111.Variations["2222"]
	expectedExperimentDecision := ExperimentDecision{
		Variation: &expectedVariation,
	}

	mockExperimentDecisionService := new(MockExperimentDecisionService)
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentServices: []ExperimentService{mockExperimentDecisionService},
	}
	decision, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.Error(t, err)
	assert.Equal(t, ExperimentDecision{}, decision)
	mockExperimentDecisionService.AssertNotCalled(t, "GetDecision")
}

func TestCompositeExperimentServiceGetDecisionTargeting(t *testing.T) {
	testUserContext := entities.UserContext{
		ID: "test_user",
	}
	testAudienceMap := map[string]entities.Audience{
		"5555": testAudience5555,
	}
	mockProjectConfig := new(mockProjectConfig)
	mockProjectConfig.On("GetAudienceMap").Return(testAudienceMap)

	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	testCondTreeParams := entities.NewTreeParameters(&testUserContext, testAudienceMap)

	// Test user fails targeting
	mockAudienceTreeEvaluator := new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, testCondTreeParams).Return(false)
	mockExperimentDecisionService := new(MockExperimentDecisionService)
	testCompositeExperimentService := &CompositeExperimentService{
		audienceTreeEvaluator: mockAudienceTreeEvaluator,
		experimentServices:    []ExperimentService{mockExperimentDecisionService},
	}
	decision, _ := testCompositeExperimentService.GetDecision(testExperimentDecisionContext, testUserContext)
	assert.Nil(t, decision.Variation)
	mockAudienceTreeEvaluator.AssertExpectations(t)
	mockExperimentDecisionService.AssertNotCalled(t, "GetDecision")

	// Test user passes targeting, moves on to children decision services
	expectedExperimentDecision := ExperimentDecision{
		Variation: &testExp1112Var2222,
	}
	mockAudienceTreeEvaluator = new(MockAudienceTreeEvaluator)
	mockAudienceTreeEvaluator.On("Evaluate", testExp1112.AudienceConditionTree, testCondTreeParams).Return(true)
	mockExperimentDecisionService = new(MockExperimentDecisionService)
	mockExperimentDecisionService.On("GetDecision", testExperimentDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)
	testCompositeExperimentService = &CompositeExperimentService{
		audienceTreeEvaluator: mockAudienceTreeEvaluator,
		experimentServices:    []ExperimentService{mockExperimentDecisionService},
	}
	decision, _ = testCompositeExperimentService.GetDecision(testExperimentDecisionContext, testUserContext)
	assert.Equal(t, decision, expectedExperimentDecision)
	mockAudienceTreeEvaluator.AssertExpectations(t)
	mockExperimentDecisionService.AssertExpectations(t)
}
