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

func TestMapHoldoutsAppliestoAllFlags(t *testing.T) {
	// Running holdouts with no includedRules are global — apply to all flags
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "running_holdout",
			Status: "Running",
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
	assert.Equal(t, "running_holdout", holdoutList[0].Key)
	assert.True(t, holdoutList[0].IsGlobal())

	// Global holdout should apply to all flags
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Contains(t, flagHoldoutsMap, "feature_3")

	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
	assert.Len(t, flagHoldoutsMap["feature_3"], 1)

	// Global holdout should NOT appear in ruleHoldoutsMap
	assert.Empty(t, ruleHoldoutsMap)
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

func TestMapHoldoutsMultipleHoldouts(t *testing.T) {
	// Multiple running global holdouts all apply to all flags
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_1",
			Key:    "holdout_1",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 5000},
			},
		},
		{
			ID:     "holdout_2",
			Key:    "holdout_2",
			Status: "Running",
			Variations: []datafileEntities.Variation{
				{ID: "var_2", Key: "variation_2"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_2", EndOfRange: 10000},
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

	// Both features should get both global holdouts
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 2)
	assert.Equal(t, "holdout_1", flagHoldoutsMap["feature_1"][0].Key)
	assert.Equal(t, "holdout_2", flagHoldoutsMap["feature_1"][1].Key)

	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_2"], 2)

	// No local holdouts
	assert.Empty(t, ruleHoldoutsMap)
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

// TestMapLocalHoldoutSingleRule tests a local holdout targeting a single rule.
func TestMapLocalHoldoutSingleRule(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "local_holdout_1",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_123"},
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.False(t, holdoutList[0].IsGlobal())
	assert.Equal(t, []string{"rule_123"}, holdoutList[0].IncludedRules)

	// Local holdout should NOT appear in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)

	// Local holdout SHOULD appear in ruleHoldoutsMap under its rule ID
	assert.Len(t, ruleHoldoutsMap, 1)
	assert.Contains(t, ruleHoldoutsMap, "rule_123")
	assert.Len(t, ruleHoldoutsMap["rule_123"], 1)
	assert.Equal(t, "local_holdout_1", ruleHoldoutsMap["rule_123"][0].ID)
}

// TestMapLocalHoldoutMultipleRules tests a local holdout targeting multiple rules.
func TestMapLocalHoldoutMultipleRules(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "local_holdout_1",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1", "rule_2", "rule_3"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}

	featureMap := map[string]entities.Feature{}

	_, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Local holdout should NOT appear in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)

	// Each rule gets an entry
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_2")
	assert.Contains(t, ruleHoldoutsMap, "rule_3")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_2"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_3"], 1)
}

// TestMapEmptyIncludedRulesIsLocal tests that an empty (non-nil) IncludedRules is local.
func TestMapEmptyIncludedRulesIsLocal(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_empty_rules",
			Key:           "empty_rules_holdout",
			Status:        "Running",
			IncludedRules: []string{}, // empty slice = local holdout with no rules
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

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 1)
	// Empty IncludedRules is NOT nil, so it is local (not global)
	assert.False(t, holdoutList[0].IsGlobal())

	// Should not appear in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)
	// No rule IDs to register, so ruleHoldoutsMap is also empty
	assert.Empty(t, ruleHoldoutsMap)
}

// TestMapMixedGlobalAndLocalHoldouts tests a mix of global and local holdouts.
func TestMapMixedGlobalAndLocalHoldouts(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "global_holdout",
			Key:    "global",
			Status: "Running",
			// nil IncludedRules = global
			Variations: []datafileEntities.Variation{
				{ID: "gvar", Key: "gvar"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "gvar", EndOfRange: 5000},
			},
		},
		{
			ID:            "local_holdout",
			Key:           "local",
			Status:        "Running",
			IncludedRules: []string{"rule_abc"},
			Variations: []datafileEntities.Variation{
				{ID: "lvar", Key: "lvar"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "lvar", EndOfRange: 3000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 2)

	// Global holdout in flagHoldoutsMap
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].ID)

	// Local holdout in ruleHoldoutsMap
	assert.Contains(t, ruleHoldoutsMap, "rule_abc")
	assert.Len(t, ruleHoldoutsMap["rule_abc"], 1)
	assert.Equal(t, "local_holdout", ruleHoldoutsMap["rule_abc"][0].ID)
}

// TestMapBackwardCompatibilityOldDatafile tests that old datafiles without
// includedRules field still parse correctly as global holdouts.
func TestMapBackwardCompatibilityOldDatafile(t *testing.T) {
	// Old datafile holdout has no IncludedRules field (nil by default)
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "old_holdout",
			Key:    "old_holdout",
			Status: "Running",
			// IncludedRules not set, defaults to nil = global
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

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 1)
	assert.Nil(t, holdoutList[0].IncludedRules)
	assert.True(t, holdoutList[0].IsGlobal())

	// Should appear in flagHoldoutsMap as before
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Empty(t, ruleHoldoutsMap)
}
