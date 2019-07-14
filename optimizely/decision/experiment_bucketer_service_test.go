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
	mockProjectConfig := new(mockProjectConfig)
	mockProjectConfig.On("GetExperimentByKey", testExp1111Key).Return(testExp1111, nil)
	testDecisionContext := ExperimentDecisionContext{
		ExperimentKey: testExp1111Key,
		ProjectConfig: mockProjectConfig,
	}

	testUserContext := entities.UserContext{
		ID: "test_user_1",
	}

	expectedDecision := ExperimentDecision{
		Variation: testExp1111Var2222,
		Decision: Decision{
			DecisionMade: true,
		},
	}
	mockBucketer := new(MockBucketer)
	mockBucketer.On("Bucket", testUserContext.ID, testExp1111, entities.Group{}).Return(testExp1111Var2222, nil)

	experimentBucketerService := ExperimentBucketerService{
		bucketer: mockBucketer,
	}
	decision, _ := experimentBucketerService.GetDecision(testDecisionContext, testUserContext)
	assert.Equal(t, expectedDecision, decision)
}
