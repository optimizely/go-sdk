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

func TestMapHoldoutsAppliestoAllFlags(t *testing.T) {
	// Running holdouts apply to all flags
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.Equal(t, "holdout_1", holdoutList[0].ID)
	assert.Equal(t, "running_holdout", holdoutList[0].Key)

	// Running holdout should apply to all flags
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Contains(t, flagHoldoutsMap, "feature_3")

	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Len(t, flagHoldoutsMap["feature_2"], 1)
	assert.Len(t, flagHoldoutsMap["feature_3"], 1)
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

func TestMapHoldoutsMultipleHoldouts(t *testing.T) {
	// Multiple running holdouts all apply to all flags
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

	holdoutList, holdoutIDMap, flagHoldoutsMap, _ := MapHoldouts(rawHoldouts, featureMap)

	// Verify both holdouts are in the list
	assert.Len(t, holdoutList, 2)
	assert.Len(t, holdoutIDMap, 2)

	// Both features should get both holdouts
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 2)
	assert.Equal(t, "holdout_1", flagHoldoutsMap["feature_1"][0].Key)
	assert.Equal(t, "holdout_2", flagHoldoutsMap["feature_1"][1].Key)

	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_2"], 2)
	assert.Equal(t, "holdout_1", flagHoldoutsMap["feature_2"][0].Key)
	assert.Equal(t, "holdout_2", flagHoldoutsMap["feature_2"][1].Key)
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

// Level 1 tests for local holdout support (FSSDK-12369)

func TestMapHoldoutsIsGlobalNilIncludedRules(t *testing.T) {
	// A holdout with no IncludedRules field (nil pointer) should be global
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_global",
			Key:           "global_holdout",
			Status:        "Running",
			IncludedRules: nil, // nil = global
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
	// nil IncludedRules should be classified as global
	assert.True(t, holdoutList[0].IsGlobal())
	assert.Nil(t, holdoutList[0].IncludedRules)
	// Global holdout should appear in flagHoldoutsMap for every feature
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	// ruleHoldoutsMap should be empty for a global holdout
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsLocalHoldoutWithIncludedRules(t *testing.T) {
	// A holdout with IncludedRules pointing to specific rule IDs should be local
	includedRules := []string{"rule_id_1", "rule_id_2"}
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_local",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: &includedRules,
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
	// Non-nil IncludedRules should be classified as local (not global)
	assert.False(t, holdoutList[0].IsGlobal())
	assert.NotNil(t, holdoutList[0].IncludedRules)
	// Local holdout must NOT appear in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)
	// Local holdout should appear in ruleHoldoutsMap for each rule it targets
	assert.Contains(t, ruleHoldoutsMap, "rule_id_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_id_2")
	assert.Len(t, ruleHoldoutsMap["rule_id_1"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_id_2"], 1)
	assert.Equal(t, "local_holdout", ruleHoldoutsMap["rule_id_1"][0].Key)
	assert.Equal(t, "local_holdout", ruleHoldoutsMap["rule_id_2"][0].Key)
}

func TestMapHoldoutsEmptyIncludedRulesIsLocalNotGlobal(t *testing.T) {
	// An empty (non-nil) IncludedRules slice is local (targets no rules), NOT global
	emptyRules := []string{}
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_empty_local",
			Key:           "empty_local_holdout",
			Status:        "Running",
			IncludedRules: &emptyRules, // non-nil, but empty
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
	// Empty non-nil IncludedRules should be local (not global)
	assert.False(t, holdoutList[0].IsGlobal())
	// Empty local holdout must NOT appear in flagHoldoutsMap
	assert.Empty(t, flagHoldoutsMap)
	// ruleHoldoutsMap should also be empty (no rules to target)
	assert.Empty(t, ruleHoldoutsMap)
}

func TestMapHoldoutsMixedGlobalAndLocal(t *testing.T) {
	// Mix of global and local holdouts
	includedRules := []string{"rule_1"}
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_global",
			Key:           "global_holdout",
			Status:        "Running",
			IncludedRules: nil,
			Variations: []datafileEntities.Variation{
				{ID: "var_g", Key: "var_global"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_g", EndOfRange: 10000},
			},
		},
		{
			ID:            "holdout_local",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: &includedRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_l", Key: "var_local"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_l", EndOfRange: 10000},
			},
		},
	}
	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 2)

	// Global holdout in flagHoldoutsMap (only the global one)
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Equal(t, "global_holdout", flagHoldoutsMap["feature_1"][0].Key)

	// Local holdout in ruleHoldoutsMap
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Equal(t, "local_holdout", ruleHoldoutsMap["rule_1"][0].Key)
}

func TestMapHoldoutsLocalHoldoutCrossRuleTargeting(t *testing.T) {
	// A single local holdout can target rules from multiple flags
	includedRules := []string{"rule_a", "rule_b", "rule_c"}
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_cross",
			Key:           "cross_rule_holdout",
			Status:        "Running",
			IncludedRules: &includedRules,
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}
	featureMap := map[string]entities.Feature{}

	_, _, _, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Each rule should have the holdout mapped
	assert.Contains(t, ruleHoldoutsMap, "rule_a")
	assert.Contains(t, ruleHoldoutsMap, "rule_b")
	assert.Contains(t, ruleHoldoutsMap, "rule_c")
	assert.Len(t, ruleHoldoutsMap["rule_a"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_b"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_c"], 1)
}

func TestMapHoldoutsIsGlobalProperty(t *testing.T) {
	// Verify IsGlobal() works correctly for both global and local holdouts
	nilRules := (*[]string)(nil)
	emptyRules := []string{}
	ruleIDs := []string{"rule_1"}

	globalHoldout := entities.Holdout{IncludedRules: nilRules}
	localHoldoutEmpty := entities.Holdout{IncludedRules: &emptyRules}
	localHoldoutWithRules := entities.Holdout{IncludedRules: &ruleIDs}

	assert.True(t, globalHoldout.IsGlobal(), "nil IncludedRules should be global")
	assert.False(t, localHoldoutEmpty.IsGlobal(), "empty non-nil IncludedRules should NOT be global")
	assert.False(t, localHoldoutWithRules.IsGlobal(), "non-nil IncludedRules with rules should NOT be global")
}

func TestMapHoldoutsBackwardCompatibilityOldDatafile(t *testing.T) {
	// Old datafiles without IncludedRules field should be treated as global
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_old",
			Key:    "old_global_holdout",
			Status: "Running",
			// No IncludedRules field — simulates old datafile format
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
	}
	featureMap := map[string]entities.Feature{
		"my_feature": {ID: "my_feature", Key: "my_feature"},
	}

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	assert.Len(t, holdoutList, 1)
	// Old datafile holdout with no IncludedRules should default to global
	assert.True(t, holdoutList[0].IsGlobal())
	assert.Contains(t, flagHoldoutsMap, "my_feature")
	assert.Len(t, flagHoldoutsMap["my_feature"], 1)
	assert.Empty(t, ruleHoldoutsMap)
}
