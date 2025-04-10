/****************************************************************************
 * Copyright 2019,2021-2025, Optimizely, Inc. and contributors                   *
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

	jsoniter "github.com/json-iterator/go"
	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
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
	experimentGroupMap := map[string]string{"11111": "15"}

	experimentsIDMap, experimentKeyMap := MapExperiments(rawExperiments, experimentGroupMap)
	expectedExperiments := map[string]entities.Experiment{
		"11111": {
			AudienceIds: []string{"31111"},
			ID:          "11111",
			GroupID:     "15",
			Key:         "test_experiment_11111",
			Variations: map[string]entities.Variation{
				"21111": {
					ID:             "21111",
					Variables:      map[string]entities.VariationVariable{"1": {ID: "1", Value: "1"}},
					Key:            "variation_1",
					FeatureEnabled: true,
				},
				"21112": {
					ID:             "21112",
					Variables:      map[string]entities.VariationVariable{"2": {ID: "2", Value: "2"}},
					Key:            "variation_2",
					FeatureEnabled: false,
				},
			},
			VariationKeyToIDMap: map[string]string{
				"variation_1": "21111",
				"variation_2": "21112",
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
			AudienceConditions: []interface{}{"or", "31111"},
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

	assert.Equal(t, expectedExperiments, experimentsIDMap)
	assert.Equal(t, expectedExperimentKeyMap, experimentKeyMap)
}

func TestMapExperimentsWithStringAudienceCondition(t *testing.T) {

	rawExperiment := datafileEntities.Experiment{
		ID:                 "11111",
		AudienceIds:        []string{"31111"},
		Key:                "test_experiment_11111",
		AudienceConditions: "31111",
	}

	rawExperiments := []datafileEntities.Experiment{rawExperiment}
	experimentGroupMap := map[string]string{"11111": "15"}

	experimentsIDMap, experimentKeyMap := MapExperiments(rawExperiments, experimentGroupMap)
	expectedExperiments := map[string]entities.Experiment{
		"11111": {
			AudienceIds:         []string{"31111"},
			ID:                  "11111",
			GroupID:             "15",
			Key:                 "test_experiment_11111",
			Variations:          map[string]entities.Variation{},
			VariationKeyToIDMap: map[string]string{},
			TrafficAllocation:   []entities.Range{},
			AudienceConditions:  "31111",
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

	assert.Equal(t, expectedExperiments, experimentsIDMap)
	assert.Equal(t, expectedExperimentKeyMap, experimentKeyMap)
}

func TestMergeExperiments(t *testing.T) {

	rawExperiment := datafileEntities.Experiment{
		ID: "11111",
	}
	rawGroup := datafileEntities.Group{
		Policy: "random",
		ID:     "11112",
		TrafficAllocation: []datafileEntities.TrafficAllocation{
			{
				EntityID:   "21113",
				EndOfRange: 7000,
			},
			{
				EntityID:   "21114",
				EndOfRange: 10000,
			},
		},
		Experiments: []datafileEntities.Experiment{
			{
				ID: "11112",
			},
		},
	}

	rawExperiments := []datafileEntities.Experiment{rawExperiment}
	rawGroups := []datafileEntities.Group{rawGroup}
	mergedExperiments := MergeExperiments(rawExperiments, rawGroups)

	expectedExperiments := []datafileEntities.Experiment{
		{
			ID: "11111",
		},
		{
			ID: "11112",
		},
	}

	assert.Equal(t, expectedExperiments, mergedExperiments)
}

func TestMapExperimentsAudienceIdsOnly(t *testing.T) {
	var rawExperiment datafileEntities.Experiment
	rawExperiment.AudienceIds = []string{"11111", "11112"}
	rawExperiment.Key = "test_experiment_1"
	rawExperiment.ID = "22222"

	expectedExperiment := entities.Experiment{
		AudienceIds:         rawExperiment.AudienceIds,
		ID:                  rawExperiment.ID,
		Key:                 rawExperiment.Key,
		Variations:          map[string]entities.Variation{},
		VariationKeyToIDMap: map[string]string{},
		TrafficAllocation:   make([]entities.Range, 0),
		AudienceConditionTree: &entities.TreeNode{
			Operator: "or",
			Nodes: []*entities.TreeNode{
				{
					Operator: "",
					Item:     "11111",
				},
				{
					Operator: "",
					Item:     "11112",
				},
			},
		},
	}

	experimentsIDMap, _ := MapExperiments([]datafileEntities.Experiment{rawExperiment}, map[string]string{})
	assert.Equal(t, expectedExperiment.AudienceConditionTree, experimentsIDMap[rawExperiment.ID].AudienceConditionTree)
}

func TestMapCmab(t *testing.T) {
    tests := []struct {
        name     string
        input    *datafileEntities.Cmab
        expected *entities.Cmab
    }{
        {
            name:     "nil input",
            input:    nil,
            expected: nil,
        },
        {
            name: "with attributes only",
            input: &datafileEntities.Cmab{
                AttributeIds:      []string{"attr1", "attr2"},
                TrafficAllocation: []datafileEntities.TrafficAllocation{},
            },
            expected: &entities.Cmab{
                AttributeIds:      []string{"attr1", "attr2"},
                TrafficAllocation: []entities.Range{},
            },
        },
        {
            name: "with traffic allocation only",
            input: &datafileEntities.Cmab{
                AttributeIds: []string{},
                TrafficAllocation: []datafileEntities.TrafficAllocation{
                    {EntityID: "var1", EndOfRange: 5000},
                    {EntityID: "var2", EndOfRange: 10000},
                },
            },
            expected: &entities.Cmab{
                AttributeIds: []string{},
                TrafficAllocation: []entities.Range{
                    {EntityID: "var1", EndOfRange: 5000},
                    {EntityID: "var2", EndOfRange: 10000},
                },
            },
        },
        {
            name: "with both attributes and traffic allocation",
            input: &datafileEntities.Cmab{
                AttributeIds: []string{"attr1", "attr2"},
                TrafficAllocation: []datafileEntities.TrafficAllocation{
                    {EntityID: "var1", EndOfRange: 5000},
                    {EntityID: "var2", EndOfRange: 10000},
                },
            },
            expected: &entities.Cmab{
                AttributeIds: []string{"attr1", "attr2"},
                TrafficAllocation: []entities.Range{
                    {EntityID: "var1", EndOfRange: 5000},
                    {EntityID: "var2", EndOfRange: 10000},
                },
            },
        },
        {
            name: "with empty traffic allocation array",
            input: &datafileEntities.Cmab{
                AttributeIds:      []string{"attr1", "attr2"},
                TrafficAllocation: []datafileEntities.TrafficAllocation{},
            },
            expected: &entities.Cmab{
                AttributeIds:      []string{"attr1", "attr2"},
                TrafficAllocation: []entities.Range{},
            },
        },
    }

    // Run tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := mapCmab(tt.input)

            if tt.expected == nil {
                assert.Nil(t, result)
            } else {
                assert.NotNil(t, result)
                assert.Equal(t, tt.expected.AttributeIds, result.AttributeIds)
                assert.Equal(t, len(tt.expected.TrafficAllocation), len(result.TrafficAllocation))

                for i, expectedTA := range tt.expected.TrafficAllocation {
                    assert.Equal(t, expectedTA.EntityID, result.TrafficAllocation[i].EntityID)
                    assert.Equal(t, expectedTA.EndOfRange, result.TrafficAllocation[i].EndOfRange)
                }
            }
        })
    }
}

func TestMapExperimentWithCmab(t *testing.T) {
    // Create a raw experiment with CMAB configuration
    rawExperiment := datafileEntities.Experiment{
        ID:      "exp1",
        Key:     "experiment_1",
        LayerID: "layer1",
        Variations: []datafileEntities.Variation{
            {ID: "var1", Key: "variation_1"},
        },
        TrafficAllocation: []datafileEntities.TrafficAllocation{
            {EntityID: "var1", EndOfRange: 10000},
        },
        Cmab: &datafileEntities.Cmab{
            AttributeIds: []string{"attr1", "attr2"},
            TrafficAllocation: []datafileEntities.TrafficAllocation{
                {EntityID: "var1", EndOfRange: 5000},
                {EntityID: "var2", EndOfRange: 10000},
            },
        },
    }

    // Map the experiment
    experiment := mapExperiment(rawExperiment)

    // Verify CMAB mapping
    assert.NotNil(t, experiment.Cmab)
    assert.Equal(t, []string{"attr1", "attr2"}, experiment.Cmab.AttributeIds)
    assert.Equal(t, 2, len(experiment.Cmab.TrafficAllocation))
    assert.Equal(t, "var1", experiment.Cmab.TrafficAllocation[0].EntityID)
    assert.Equal(t, 5000, experiment.Cmab.TrafficAllocation[0].EndOfRange)
    assert.Equal(t, "var2", experiment.Cmab.TrafficAllocation[1].EntityID)
    assert.Equal(t, 10000, experiment.Cmab.TrafficAllocation[1].EndOfRange)
}
