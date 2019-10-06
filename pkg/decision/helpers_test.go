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
	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/stretchr/testify/mock"
)

// Mock implementation of ProjectConfig
type mockProjectConfig struct {
	pkg.ProjectConfig
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

func (c *mockProjectConfig) GetAudienceMap() map[string]entities.Audience {
	args := c.Called()
	return args.Get(0).(map[string]entities.Audience)
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

type MockAudienceTreeEvaluator struct {
	mock.Mock
}

func (m *MockAudienceTreeEvaluator) Evaluate(node *entities.TreeNode, condTreeParams *entities.TreeParameters) bool {
	args := m.Called(node, condTreeParams)
	return args.Bool(0)
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
var testAudience5555 = entities.Audience{ID: "5555"}
var testExp1112 = entities.Experiment{
	AudienceConditionTree: &entities.TreeNode{
		Operator: "and",
		Nodes: []*entities.TreeNode{
			&entities.TreeNode{Item: "test_audience_5555"},
		},
	},
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

// Feature with test and rollout
const testFeat3335Key = "test_feature_3335_key"

// Will use this experiment for feature test
const testExp1113Key = "test_experiment_1113"

var testExp1113Var2223 = entities.Variation{ID: "2223", Key: "2223", FeatureEnabled: true}
var testExp1113Var2224 = entities.Variation{ID: "2224", Key: "2224", FeatureEnabled: false}
var testExp1113 = entities.Experiment{
	ID:      "1113",
	Key:     testExp1113Key,
	GroupID: "6666",
	Variations: map[string]entities.Variation{
		"2223": testExp1113Var2223,
		"2224": testExp1113Var2224,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2223", EndOfRange: 5000},
		entities.Range{EntityID: "2224", EndOfRange: 10000},
	},
}

const testExp1114Key = "test_experiment_1114"

var testExp1114Var2225 = entities.Variation{ID: "2225", Key: "2225", FeatureEnabled: true}
var testExp1114Var2226 = entities.Variation{ID: "2226", Key: "2226", FeatureEnabled: false}
var testExp1114 = entities.Experiment{
	ID:      "1114",
	Key:     testExp1114Key,
	GroupID: "6666",
	Variations: map[string]entities.Variation{
		"2225": testExp1114Var2225,
		"2226": testExp1114Var2226,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2225", EndOfRange: 5000},
		entities.Range{EntityID: "2226", EndOfRange: 10000},
	},
}
var testGroup6666 = entities.Group{
	ID:     "6666",
	Policy: "random",
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "1113", EndOfRange: 3000},
		entities.Range{EntityID: "1114", EndOfRange: 6000},
	},
}

// Will use this experiment for rollout
const testExp1115Key = "test_experiment_1115"

var testExp1115Var2227 = entities.Variation{ID: "2227", Key: "2227", FeatureEnabled: true}
var testExp1115 = entities.Experiment{
	ID:  "1115",
	Key: testExp1115Key,
	Variations: map[string]entities.Variation{
		"2227": testExp1115Var2227,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2227", EndOfRange: 5000},
	},
}
var testFeat3335 = entities.Feature{
	ID:                 "3335",
	Key:                testFeat3335Key,
	FeatureExperiments: []entities.Experiment{testExp1113, testExp1114},
	Rollout: entities.Rollout{
		ID:          "4445",
		Experiments: []entities.Experiment{testExp1115},
	},
}

// Targeted experiment
const testTargetedExp1116Key = "test_targeted_experiment_1116"

var testTargetedExp1116Var2228 = entities.Variation{ID: "2228", Key: "2228"}
var testTargetedExp1116 = entities.Experiment{
	AudienceConditionTree: &entities.TreeNode{Operator: "or", Item: "7771"},
	ID:                    "1116",
	Key:                   testTargetedExp1116Key,
	Variations: map[string]entities.Variation{
		"2228": testTargetedExp1116Var2228,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2228", EndOfRange: 10000},
	},
}

// Experiment with a whitelist
const testExpWhitelistKey = "test_experiment_whitelist"

var testExpWhitelistVar2229 = entities.Variation{ID: "2229", Key: "2229"}
var testExpWhitelist = entities.Experiment{
	ID:  "1117",
	Key: testExpWhitelistKey,
	Variations: map[string]entities.Variation{
		"2229": testExpWhitelistVar2229,
	},
	TrafficAllocation: []entities.Range{
		entities.Range{EntityID: "2229", EndOfRange: 10000},
	},
	Whitelist: map[string]string{
		"test_user_1": "2229",
		// Note: this is an invalid entry, there is no variation 2230 in this experiment
		"test_user_2": "2230",
	},
}
