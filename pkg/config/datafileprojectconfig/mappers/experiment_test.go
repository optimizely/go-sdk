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
	"testing"

	"github.com/json-iterator/go"
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
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
				"featureEnabled": true,
				"variables": [{"id":"1","value":"1"}]
			},
			{
				"id": "21112",
				"key": "variation_2",
				"featureEnabled": false,
				"variables": [{"id":"2","value":"2"}]
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
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
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
					Variables:      map[string]entities.VariationVariable{"1": entities.VariationVariable{ID: "1", Value: "1"}},
					Key:            "variation_1",
					FeatureEnabled: true,
				},
				"21112": {
					ID:             "21112",
					Variables:      map[string]entities.VariationVariable{"2": entities.VariationVariable{ID: "2", Value: "2"}},
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
			AudienceConditionTree: &entities.TreeNode{
				Operator: "or",
				Nodes: []*entities.TreeNode{
					{
						Operator: "",
						Item:     "31111",
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
