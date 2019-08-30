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

/**
 * This file contains mocks and test data to be used by test files throughout this package.
 */

// Package decision //
package decision

import (
	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
)

// Mock implementation of ProjectConfig
type mockProjectConfig struct {
	optimizely.ProjectConfig
	mock.Mock
}

func (c *mockProjectConfig) GetFeatureByKey(featureKey string) (entities.Feature, error) {
	args := c.Called(featureKey)
	return args.Get(0).(entities.Feature), args.Error(1)
}

func (c *mockProjectConfig) GetExperimentByKey(experimentKey string) (entities.Experiment, error) {
	args := c.Called(experimentKey)
	return args.Get(0).(entities.Experiment), args.Error(1)
}

func (c *mockProjectConfig) GetAudienceByID(audienceID string) (entities.Audience, error) {
	args := c.Called(audienceID)
	return args.Get(0).(entities.Audience), args.Error(1)
}

type MockExperimentDecisionService struct {
	mock.Mock
}

func (m *MockExperimentDecisionService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(ExperimentDecision), args.Error(1)
}

type MockFeatureDecisionService struct {
	mock.Mock
}

func (m *MockFeatureDecisionService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	args := m.Called(decisionContext, userContext)
	return args.Get(0).(FeatureDecision), args.Error(1)
}

// Single variation experiment
const testExp1111Key = "test_experiment_1111"

var testExp1111Var2222 = entities.Variation{ID: "2222", Key: "2222"}
var testExp1111 = entities.Experiment{
	ID:  "1111",
	Key: testExp1111Key,
	Variations: map[string]entities.Variation{
		"2222": testExp1111Var2222,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2222", EndOfRange: 10000},
	},
	Status: entities.Running,
}

// Simple feature test
const testFeat3333Key = "my_test_feature_3333"

var testFeat3333 = entities.Feature{
	ID:                 "3333",
	Key:                testFeat3333Key,
	FeatureExperiments: []entities.Experiment{testExp1111},
}

// Feature rollout
var testExp1112Var2222 = entities.Variation{ID: "2222", Key: "2222"}
var testExp1112 = entities.Experiment{
	ID:  "1112",
	Key: testExp1111Key,
	Variations: map[string]entities.Variation{
		"2222": testExp1111Var2222,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2222", EndOfRange: 10000},
	},
}

const testFeatRollout3334Key = "test_feature_rollout_3334_key"

var testFeatRollout3334 = entities.Feature{
	ID:  "3334",
	Key: testFeatRollout3334Key,
	Rollout: entities.Rollout{
		ID:          "4444",
		Experiments: []entities.Experiment{testExp1112},
	},
}
