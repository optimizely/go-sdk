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

package mappers

import (
	"encoding/json"
	"testing"

	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapExperiments(t *testing.T) {
	const testExperimentString = `{
		"audienceIds": ["31111"],
		"id": "11111",
		"key": "test_experiment_11111",
		"variations": [
			{
				"id": "21111",
				"key": "variation_1",
				"featureEnabled": true
			},
			{
				"id": "21112",
				"key": "variation_2",
				"featureEnabled": false
			}
		],
		"trafficAllocation": [
			{
				"entityId": "21111",
				"endOfRange": 7000
			},
			{
				"entityId": "21112",
				"endOfRange": 10000
			}
		],
		"audienceConditions": [
			"or",
			"31111"
		]
	}`

	var rawExperiment datafileEntities.Experiment
	json.Unmarshal([]byte(testExperimentString), &rawExperiment)

	rawExperiments := []datafileEntities.Experiment{rawExperiment}
	experiments, experimentKeyMap := MapExperiments(rawExperiments)
	expectedExperiments := map[string]entities.Experiment{
		"11111": {
			AudienceIds: []string{"31111"},
			ID:          "11111",
			Key:         "test_experiment_11111",
			Variations: map[string]entities.Variation{
				"21111": {
					ID:             "21111",
					Key:            "variation_1",
					FeatureEnabled: true,
				},
				"21112": {
					ID:             "21112",
					Key:            "variation_2",
					FeatureEnabled: false,
				},
			},
			TrafficAllocation: []entities.Range{
				{
					EntityID:   "21111",
					EndOfRange: 7000,
				},
				{
					EntityID:   "21112",
					EndOfRange: 10000,
				},
			},
			AudienceConditionTree: &entities.ConditionTreeNode{
				Operator: "or",
				Nodes: []*entities.ConditionTreeNode{
					{
						Operator: "",
						Condition: entities.Condition{
							Name:  "optimizely_populated",
							Match: "",
							Type:  "audience_condition",
							Value: "31111",
						},
					},
				},
			},
		},
	}
	expectedExperimentKeyMap := map[string]string{
		"test_experiment_11111": "11111",
	}

	assert.Equal(t, expectedExperiments, experiments)
	assert.Equal(t, expectedExperimentKeyMap, experimentKeyMap)
}
