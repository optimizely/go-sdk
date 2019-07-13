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
	testExperimentKey := "test_experiment"
	testExperiment := entities.Experiment{
		ID:  "111111",
		Key: testExperimentKey,
		Variations: map[string]entities.Variation{
			"22222": entities.Variation{
				ID:  "22222",
				Key: "22222",
			},
		},
	}
	mockProjectConfig := new(MockProjectConfig)
	mockProjectConfig.On("GetExperimentByKey", testExperimentKey).Return(testExperiment, nil)
	testDecisionContext := ExperimentDecisionContext{
		ExperimentKey: testExperimentKey,
		ProjectConfig: mockProjectConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedExperimentDecision := ExperimentDecision{
		Variation: testExperiment.Variations["22222"],
		Decision: Decision{
			DecisionMade: true,
		},
	}
	// test that we return out of the decision making and the next one doesn't get called
	mockExperimentDecisionService := new(MockExperimentDecisionService)
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 := new(MockExperimentDecisionService)
	compositeExperimentService := &CompositeExperimentService{
		experimentDecisionServices: []ExperimentDecisionService{
			mockExperimentDecisionService,
			mockExperimentDecisionService2,
		},
	}
	decision, err := compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision, decision)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertNotCalled(t, "GetDecision")

	// test that we move on to the next decision service if no decision is made
	mockExperimentDecisionService = new(MockExperimentDecisionService)
	expectedExperimentDecision = ExperimentDecision{
		Decision: Decision{
			DecisionMade: false,
		},
	}
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 = new(MockExperimentDecisionService)
	expectedExperimentDecision2 := ExperimentDecision{
		Variation: testExperiment.Variations["22222"],
		Decision: Decision{
			DecisionMade: true,
		},
	}
	mockExperimentDecisionService2.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService = &CompositeExperimentService{
		experimentDecisionServices: []ExperimentDecisionService{
			mockExperimentDecisionService,
			mockExperimentDecisionService2,
		},
	}
	decision, err = compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision2, decision)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertExpectations(t)

	// test when no decisions are made
	mockExperimentDecisionService = new(MockExperimentDecisionService)
	expectedExperimentDecision = ExperimentDecision{
		Decision: Decision{
			DecisionMade: false,
		},
	}
	mockExperimentDecisionService.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision, nil)

	mockExperimentDecisionService2 = new(MockExperimentDecisionService)
	expectedExperimentDecision2 = ExperimentDecision{
		Decision: Decision{
			DecisionMade: false,
		},
	}
	mockExperimentDecisionService2.On("GetDecision", testDecisionContext, testUserContext).Return(expectedExperimentDecision2, nil)

	compositeExperimentService = &CompositeExperimentService{
		experimentDecisionServices: []ExperimentDecisionService{
			mockExperimentDecisionService,
			mockExperimentDecisionService2,
		},
	}
	decision, err = compositeExperimentService.GetDecision(testDecisionContext, testUserContext)

	assert.NoError(t, err)
	assert.Equal(t, expectedExperimentDecision2, decision)
	mockExperimentDecisionService.AssertExpectations(t)
	mockExperimentDecisionService2.AssertExpectations(t)
}
