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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockBucketer struct {
	mock.Mock
}

func (m *MockBucketer) Bucket(bucketingID string, experiment entities.Experiment, group entities.Group) entities.Variation {
	args := m.Called(bucketingID, experiment, group)
	return args.Get(0).(entities.Variation)
}

func TestExperimentBucketerGetDecision(t *testing.T) {
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
			TrafficAllocation: []entities.Range{
				entities.Range{
					EntityID:   "22222",
					EndOfRange: 10000,
				},
			},
		},
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := ExperimentDecision{
		Variation: testVariation,
		Decision: Decision{
			DecisionMade: true,
		},
	}
	mockBucketer := new(MockBucketer)
	mockBucketer.On("Bucket", testUserContext.ID, testDecisionContext.Experiment, entities.Group{}).Return(testVariation, nil)

	experimentBucketerService := ExperimentBucketerService{
		bucketer: mockBucketer,
	}
	decision, _ := experimentBucketerService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedDecision, decision)
}
