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

	"github.com/stretchr/testify/mock"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

type MockFeatureDecisionService struct {
	mock.Mock
}

func (m *MockFeatureDecisionService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(FeatureDecision), args.Error(1)
}

func TestGetFeatureDecision(t *testing.T) {
	decisionContext := FeatureDecisionContext{
		Feature: entities.Feature{
			Key: "my_test_feature",
		},
	}

	userContext := entities.UserContext{
		ID: "test_user",
	}

	expectedFeatureDecision := FeatureDecision{
		FeatureEnabled: true,
		Decision:       Decision{DecisionMade: true},
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
