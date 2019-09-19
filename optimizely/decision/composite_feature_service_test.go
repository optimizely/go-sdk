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
	"github.com/optimizely/go-sdk/optimizely/decision/reasons"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompositeFeatureServiceGetDecisionFeatureExperiment(t *testing.T) {
	mockProjectConfig := new(mockProjectConfig)
	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: mockProjectConfig,
	}
	testExperimentDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: mockProjectConfig,
	}
	mockExperimentDecision := ExperimentDecision{
		Decision:  Decision{reasons.BucketedIntoVariation},
		Variation: &testExp1113Var2223,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	mockExperimentService := new(MockExperimentDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)
	// Mock to return decision from feature experiment service
	mockExperimentService.On("GetDecision", testExperimentDecisionContext, testUserContext).Return(mockExperimentDecision, nil)

	// Decision is returned from feature test evaluation
	expectedDecision := FeatureDecision{
		Decision:   Decision{reasons.BucketedIntoVariation},
		Source:     FeatureTest,
		Experiment: testExp1113,
		Variation:  &testExp1113Var2223,
	}

	compositeFeatureService := &CompositeFeatureService{
		featureExperimentService: mockExperimentService,
		rolloutDecisionService:   mockRolloutService,
	}
	actualDecision, err := compositeFeatureService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, actualDecision)
}

func TestCompositeFeatureServiceGetDecisionRollout(t *testing.T) {
	mockProjectConfig := new(mockProjectConfig)
	testFeatureDecisionContext := FeatureDecisionContext{
		Feature:       &testFeat3335,
		ProjectConfig: mockProjectConfig,
	}
	testExperimentDecisionContext1 := ExperimentDecisionContext{
		Experiment:    &testExp1113,
		ProjectConfig: mockProjectConfig,
	}
	testExperimentDecisionContext2 := ExperimentDecisionContext{
		Experiment:    &testExp1114,
		ProjectConfig: mockProjectConfig,
	}

	// Mock to not bucket user in feature experiment
	mockExperimentDecision := ExperimentDecision{
		Decision:  Decision{reasons.NotBucketedIntoVariation},
		Variation: nil,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	mockFeatureDecision := FeatureDecision{
		Decision:   Decision{reasons.BucketedIntoVariation},
		Source:     Rollout,
		Experiment: testExp1115,
		Variation:  &testExp1115Var2227,
	}

	mockExperimentService := new(MockExperimentDecisionService)
	mockRolloutService := new(MockFeatureDecisionService)
	// Mock to return decision from feature experiment service which causes rollout service to be called
	mockExperimentService.On("GetDecision", testExperimentDecisionContext1, testUserContext).Return(mockExperimentDecision, nil)
	mockExperimentService.On("GetDecision", testExperimentDecisionContext2, testUserContext).Return(mockExperimentDecision, nil)
	mockRolloutService.On("GetDecision", testFeatureDecisionContext, testUserContext).Return(mockFeatureDecision, nil)

	// Decision is returned from rollout evaluation
	expectedDecision := mockFeatureDecision

	compositeFeatureService := &CompositeFeatureService{
		featureExperimentService: mockExperimentService,
		rolloutDecisionService:   mockRolloutService,
	}
	actualDecision, err := compositeFeatureService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.NoError(t, err)
	assert.Equal(t, expectedDecision, actualDecision)
}
