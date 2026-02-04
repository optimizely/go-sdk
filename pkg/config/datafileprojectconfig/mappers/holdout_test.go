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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, flagHoldoutsMap)
}

func TestMapHoldoutsGlobalHoldout(t *testing.T) {
	// Global holdout: no includedFlags, applies to all flags except excluded
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "global_holdout",
			Status:        "Running",
			ExcludedFlags: []string{"feature_2"},
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.Equal(t, "holdout_1", holdoutList[0].ID)
	assert.Equal(t, "global_holdout", holdoutList[0].Key)

	// Global holdout should apply to feature_1 and feature_3, but NOT feature_2 (excluded)
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.NotContains(t, flagHoldoutsMap, "feature_2")
	assert.Contains(t, flagHoldoutsMap, "feature_3")

	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_3"], 1)
}

func TestMapHoldoutsSpecificHoldout(t *testing.T) {
	// Specific holdout: has includedFlags, only applies to those flags
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "specific_holdout",
			Status:        "Running",
			IncludedFlags: []string{"feature_1", "feature_2"},
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)

	// Specific holdout should only apply to feature_1 and feature_2
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.NotContains(t, flagHoldoutsMap, "feature_3")

	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Non-running holdouts should be filtered out
	assert.Empty(t, holdoutList)
	assert.Empty(t, holdoutIDMap)
	assert.Empty(t, flagHoldoutsMap)
}

func TestMapHoldoutsMixed(t *testing.T) {
	// Mix of global and specific holdouts
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_global",
			Key:           "global_holdout",
			Status:        "Running",
			ExcludedFlags: []string{"feature_2"},
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
			IncludedFlags: []string{"feature_2"},
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify both holdouts are in the list
	assert.Len(t, holdoutList, 2)
	assert.Len(t, holdoutIDMap, 2)

	// feature_1: should get global holdout only (not excluded, not specifically included)
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].Key)

	// feature_2: should get specific holdout only (excluded from global)
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
	assert.Equal(t, "specific_holdout", flagHoldoutsMap["feature_2"][0].Key)
}

func TestMapHoldoutsPrecedence(t *testing.T) {
	// Test that global holdouts take precedence over specific holdouts
	// When both apply to the same flag, global should come first in the slice
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_global",
			Key:    "global_holdout",
			Status: "Running",
			// No includedFlags = global, applies to all
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
			IncludedFlags: []string{"feature_1"}, // Specific to feature_1
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
	}

	_, _, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// feature_1 should have BOTH holdouts, with global FIRST (precedence)
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 2)

	// Global holdout should be first (takes precedence)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].Key, "Global holdout should take precedence (be first)")
	// Specific holdout should be second
	assert.Equal(t, "specific_holdout", flagHoldoutsMap["feature_1"][1].Key, "Specific holdout should be second")
}

func TestMapHoldoutsExcludedFlagsNotInMap(t *testing.T) {
	// Test that excluded flags do not get global holdouts
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_global",
			Key:           "global_holdout",
			Status:        "Running",
			ExcludedFlags: []string{"feature_excluded"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_included": {ID: "feature_included", Key: "feature_included"},
		"feature_excluded": {ID: "feature_excluded", Key: "feature_excluded"},
	}

	_, _, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// feature_included should have the global holdout
	assert.Contains(t, flagHoldoutsMap, "feature_included")
	assert.Len(t, flagHoldoutsMap["feature_included"], 1)

	// feature_excluded should NOT be in the map (no holdouts apply)
	assert.NotContains(t, flagHoldoutsMap, "feature_excluded")
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

func TestMapHoldoutsWithExperiments(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:          "holdout_1",
			Key:         "local_holdout_1",
			Status:      string(entities.HoldoutStatusRunning),
			Experiments: []string{"exp_1"},
		},
		{
			ID:          "holdout_2",
			Key:         "local_holdout_2",
			Status:      string(entities.HoldoutStatusRunning),
			Experiments: []string{"exp_1", "exp_2"},
		},
		{
			ID:          "holdout_3",
			Key:         "global_holdout",
			Status:      string(entities.HoldoutStatusRunning),
			Experiments: []string{},
		},
	}
	featureMap := map[string]entities.Feature{}

	holdoutList, _, _, experimentHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify all holdouts are in the list
	assert.Len(t, holdoutList, 3)

	// Verify experiments field is mapped correctly
	assert.Len(t, holdoutList[0].Experiments, 1)
	assert.Contains(t, holdoutList[0].Experiments, "exp_1")
	assert.Len(t, holdoutList[1].Experiments, 2)
	assert.Contains(t, holdoutList[1].Experiments, "exp_1")
	assert.Contains(t, holdoutList[1].Experiments, "exp_2")
	assert.Empty(t, holdoutList[2].Experiments)

	// Verify experiment holdouts map
	assert.Len(t, experimentHoldoutsMap, 2)
	assert.Len(t, experimentHoldoutsMap["exp_1"], 2)
	assert.Len(t, experimentHoldoutsMap["exp_2"], 1)

	// Check exp_1 has holdout_1 and holdout_2
	exp1HoldoutIDs := []string{experimentHoldoutsMap["exp_1"][0].ID, experimentHoldoutsMap["exp_1"][1].ID}
	assert.Contains(t, exp1HoldoutIDs, "holdout_1")
	assert.Contains(t, exp1HoldoutIDs, "holdout_2")

	// Check exp_2 has only holdout_2
	assert.Equal(t, "holdout_2", experimentHoldoutsMap["exp_2"][0].ID)
}

func TestMapHoldoutsMultipleExperiments(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:          "holdout_multi",
			Key:         "multi_exp_holdout",
			Status:      string(entities.HoldoutStatusRunning),
			Experiments: []string{"exp_a", "exp_b", "exp_c"},
		},
	}
	featureMap := map[string]entities.Feature{}

	_, _, _, experimentHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// All three experiments should map to the same holdout
	assert.Len(t, experimentHoldoutsMap, 3)
	for _, expID := range []string{"exp_a", "exp_b", "exp_c"} {
		assert.Len(t, experimentHoldoutsMap[expID], 1)
		assert.Equal(t, "holdout_multi", experimentHoldoutsMap[expID][0].ID)
	}
}
