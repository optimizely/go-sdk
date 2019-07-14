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

func TestGetFeatureDecision(t *testing.T) {
	mockProjectConfig := new(mockProjectConfig)
	mockProjectConfig.On("GetFeatureByKey", testFeat3333Key).Return(testFeat3333, nil)
	decisionContext := FeatureDecisionContext{
		Feature:       &testFeat3333,
		ProjectConfig: mockProjectConfig,
	}

	userContext := entities.UserContext{
		ID: "test_user",
	}

	expectedFeatureDecision := FeatureDecision{
		Experiment: testExp1111,
		Variation:  testExp1111Var2222,
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
