/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/stretchr/testify/assert"
)

func TestMapHoldoutsEmpty(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{}
	featureMap := map[string]entities.Feature{}

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, flagHoldoutsMap)
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsGlobalHoldout(t *testing.T) {
	// Global holdout: no IncludedRules, applies to all flags
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "global_holdout",
			Status: "Running",
			// No IncludedRules - this is a global holdout
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
		"feature_2": {ID: "feature_2", Key: "feature_2"},
		"feature_3": {ID: "feature_3", Key: "feature_3"},
	}

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.Equal(t, "holdout_1", holdoutList[0].ID)
	assert.Equal(t, "global_holdout", holdoutList[0].Key)

	// Global holdout should apply to all features
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Contains(t, flagHoldoutsMap, "feature_3")

	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
	assert.Len(t, flagHoldoutsMap["feature_3"], 1)

	// Global holdout should NOT be in ruleHoldoutsMap
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsSpecificHoldout(t *testing.T) {
	// Local holdout: has IncludedRules, applies to specific rules
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "specific_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1", "rule_2"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
		"feature_2": {ID: "feature_2", Key: "feature_2"},
		"feature_3": {ID: "feature_3", Key: "feature_3"},
	}

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)

	// Local holdout should NOT be in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)

	// Local holdout should be in ruleHoldoutsMap for rule_1 and rule_2
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_2")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_2"], 1)
}

func TestMapHoldoutsNotRunning(t *testing.T) {
	// Holdout with non-Running status should be filtered out
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "paused_holdout",
			Status: "Paused",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Non-running holdouts should be filtered out
	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, flagHoldoutsMap)
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsMixed(t *testing.T) {
	// Mix of global and local holdouts
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_global",
			Key:    "global_holdout",
			Status: "Running",
			// No IncludedRules - global holdout
			Variations: []datafileEntities.Variation{
				{ID: "var_global", Key: "variation_global"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_global", EndOfRange: 5000},
			},
		},
		{
			ID:            "holdout_specific",
			Key:           "specific_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1"},
			Variations: []datafileEntities.Variation{
				{ID: "var_specific", Key: "variation_specific"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_specific", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
		"feature_2": {ID: "feature_2", Key: "feature_2"},
	}

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify both holdouts are in the list
	assert.Len(t, holdoutList, 2)
	assert.Len(t, holdoutIDMap, 2)

	// All features should get global holdout
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].Key)

	// rule_1 should get local holdout
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Equal(t, "specific_holdout", ruleHoldoutsMap["rule_1"][0].Key)
}

func TestMapHoldoutsPrecedence(t *testing.T) {
	// Test that global holdouts are in flagHoldoutsMap, local holdouts are in ruleHoldoutsMap
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_global",
			Key:    "global_holdout",
			Status: "Running",
			// No IncludedRules = global, applies to all
			Variations: []datafileEntities.Variation{
				{ID: "var_global", Key: "variation_global"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_global", EndOfRange: 5000},
			},
		},
		{
			ID:            "holdout_local",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1"}, // Local to rule_1
			Variations: []datafileEntities.Variation{
				{ID: "var_local", Key: "variation_local"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_local", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	_, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// feature_1 should have global holdout
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].Key)

	// rule_1 should have local holdout
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Equal(t, "local_holdout", ruleHoldoutsMap["rule_1"][0].Key)
}

func TestMapHoldoutsGlobalAppliestoAllFeatures(t *testing.T) {
	// Test that global holdouts apply to all features
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_global",
			Key:    "global_holdout",
			Status: "Running",
			// No IncludedRules - global holdout
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
		"feature_2": {ID: "feature_2", Key: "feature_2"},
	}

	_, _, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// All features should have the global holdout
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
}

func TestMapHoldoutsWithAudienceConditions(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:                 "holdout_1",
			Key:                "holdout_with_audience",
			Status:             "Running",
			AudienceIds:        []string{"audience_1", "audience_2"},
			AudienceConditions: []interface{}{"or", "audience_1", "audience_2"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, _, _, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify audience conditions are mapped
	assert.Len(t, holdoutList, 1)
	assert.Equal(t, []string{"audience_1", "audience_2"}, holdoutList[0].AudienceIds)
	assert.NotNil(t, holdoutList[0].AudienceConditionTree)
}

func TestMapHoldoutsVariationsMapping(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "holdout_variations",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{
					ID:             "var_1",
					Key:            "variation_1",
					FeatureEnabled: true,
					Variables: []datafileEntities.VariationVariable{
						{ID: "var_var_1", Value: "value_1"},
					},
				},
				{
					ID:             "var_2",
					Key:            "variation_2",
					FeatureEnabled: false,
				},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 5000},
				{EntityID: "var_2", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, _, _, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify variations are mapped correctly
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutList[0].Variations, 2)
	assert.Contains(t, holdoutList[0].Variations, "var_1")
	assert.Contains(t, holdoutList[0].Variations, "var_2")

	// Verify traffic allocation
	assert.Len(t, holdoutList[0].TrafficAllocation, 2)
	assert.Equal(t, "var_1", holdoutList[0].TrafficAllocation[0].EntityID)
	assert.Equal(t, 5000, holdoutList[0].TrafficAllocation[0].EndOfRange)
	assert.Equal(t, "var_2", holdoutList[0].TrafficAllocation[1].EntityID)
	assert.Equal(t, 10000, holdoutList[0].TrafficAllocation[1].EndOfRange)
}
