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
	"github.com/stretchr/testify/mock"
)

type MockExperimentDecisionService struct {
	mock.Mock
}

func (m *MockExperimentDecisionService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(ExperimentDecision), args.Error(1)
}

func TestCompositeExperimentServiceGetDecision(t *testing.T) {
	testDecisionContext := ExperimentDecisionContext{
		Experiment: entities.Experiment{
			ID: "111111",
			Variations: map[string]entities.Variation{
				"22222": entities.Variation{
					ID:  "22222",
					Key: "22222",
				},
			},
		},
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedExperimentDecision := ExperimentDecision{
		Variation: testDecisionContext.Experiment.Variations["22222"],
		Decision: Decision{
			DecisionMade: true,
		},
	}
	mockExperimentDecisionService := new(MockExperimentDecisionService)
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	compositeExperimentService := &CompositeExperimentService{
		experimentDecisionServices: []ExperimentDecisionService{mockExperimentDecisionService},
	}
	decision, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision, decision)
	mockExperimentDecisionService.AssertExpectations(t)
}
