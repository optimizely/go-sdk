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
	"errors"
	"testing"

	"github.com/optimizely/go-sdk/optimizely/decision/reasons"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/optimizely/entities"
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

	// Test experiment passes targeting and bucketing
	testExperimentTargetingDecision := ExperimentDecision{} // zero-value decision means the user passed targeting
	testExperimentTargetingDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	testExperimentBucketerDecision := ExperimentDecision{
		Variation: &testExp1112Var2222,
	}
	testExperimentBucketerDecisionContext := ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	mockExperimentTargetingService := new(MockExperimentDecisionService)
	mockExperimentTargetingService.On("GetDecision", testExperimentTargetingDecisionContext, testUserContext).Return(testExperimentTargetingDecision, nil)
	mockExperimentBucketerService := new(MockExperimentDecisionService)
	mockExperimentBucketerService.On("GetDecision", testExperimentBucketerDecisionContext, testUserContext).Return(testExperimentBucketerDecision, nil)
	testRolloutService := RolloutService{
		experimentTargetingService: mockExperimentTargetingService,
		experimentBucketerService:  mockExperimentBucketerService,
	}
	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1112,
		Variation:  &testExp1112Var2222,
	}
	decision, _ := testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.Equal(t, expectedFeatureDecision, decision)
	mockExperimentTargetingService.AssertExpectations(t)
	mockExperimentBucketerService.AssertExpectations(t)

	// Test experiment passes targeting but not bucketing
	testExperimentTargetingDecision = ExperimentDecision{} // zero-value decision means the user passed targeting
	testExperimentTargetingDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	testExperimentBucketerDecision = ExperimentDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
	}
	testExperimentBucketerDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	mockExperimentTargetingService = new(MockExperimentDecisionService)
	mockExperimentTargetingService.On("GetDecision", testExperimentTargetingDecisionContext, testUserContext).Return(testExperimentTargetingDecision, nil)
	mockExperimentBucketerService = new(MockExperimentDecisionService)
	mockExperimentBucketerService.On("GetDecision", testExperimentBucketerDecisionContext, testUserContext).Return(testExperimentBucketerDecision, nil)
	testRolloutService = RolloutService{
		experimentTargetingService: mockExperimentTargetingService,
		experimentBucketerService:  mockExperimentBucketerService,
	}
	expectedFeatureDecision = FeatureDecision{
		Decision: Decision{
			Reason: reasons.NotBucketedIntoVariation,
		},
		Experiment: testExp1112,
	}
	decision, _ = testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	assert.Equal(t, expectedFeatureDecision, decision)
	mockExperimentBucketerService.AssertExpectations(t)
	mockExperimentTargetingService.AssertExpectations(t)

	// Test experiment fails targeting
	testExperimentTargetingDecision = ExperimentDecision{} // zero-value variation means the user failed targeting
	testExperimentTargetingDecisionContext = ExperimentDecisionContext{
		Experiment:    &testExp1112,
		ProjectConfig: mockProjectConfig,
	}
	expectedErr := errors.New("User failed targeting")
	mockExperimentTargetingService = new(MockExperimentDecisionService)
	mockExperimentTargetingService.On("GetDecision", testExperimentTargetingDecisionContext, testUserContext).Return(testExperimentTargetingDecision, expectedErr)
	testRolloutService = RolloutService{
		experimentTargetingService: mockExperimentTargetingService,
		experimentBucketerService:  mockExperimentBucketerService,
	}
	expectedFeatureDecision = FeatureDecision{
		Decision: Decision{
			Reason: reasons.FailedRolloutTargeting,
		},
	}
	_, err := testRolloutService.GetDecision(testFeatureDecisionContext, testUserContext)
	if assert.Error(t, err) {
		assert.Equal(t, expectedErr, err)
	} else {
		panic("Error expected")
	}
	mockExperimentTargetingService.AssertExpectations(t)
	mockExperimentBucketerService.AssertNotCalled(t, "GetDecision")
}
