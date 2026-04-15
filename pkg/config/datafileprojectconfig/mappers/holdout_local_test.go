/****************************************************************************
 * Copyright 2025-2026, Optimizely, Inc. and contributors                        *
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

// TestMapHoldoutsLocalSingleRule tests local holdout with a single rule
func TestMapHoldoutsLocalSingleRule(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "local_holdout_single",
			Status:        "Running",
			IncludedRules: []string{"rule_1"},
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

	// Verify holdout list and ID map
	assert.Len(t, holdoutList, 1)
	assert.Len(t, holdoutIDMap, 1)
	assert.Equal(t, "holdout_1", holdoutList[0].ID)
	assert.Equal(t, "local_holdout_single", holdoutList[0].Key)
	assert.NotNil(t, holdoutList[0].IncludedRules)
	assert.Len(t, holdoutList[0].IncludedRules, 1)
	assert.Equal(t, "rule_1", holdoutList[0].IncludedRules[0])

	// Local holdout should NOT appear in flagHoldoutsMap (only global holdouts)
	assert.Empty(t, flagHoldoutsMap)

	// Local holdout should appear in ruleHoldoutsMap for rule_1
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Equal(t, "holdout_1", ruleHoldoutsMap["rule_1"][0].ID)
}

// TestMapHoldoutsLocalMultipleRules tests local holdout with multiple rules
func TestMapHoldoutsLocalMultipleRules(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "local_holdout_multi",
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

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	holdoutList, _, _, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify holdout has all included rules
	assert.Len(t, holdoutList[0].IncludedRules, 3)
	assert.Contains(t, holdoutList[0].IncludedRules, "rule_1")
	assert.Contains(t, holdoutList[0].IncludedRules, "rule_2")
	assert.Contains(t, holdoutList[0].IncludedRules, "rule_3")

	// Verify ruleHoldoutsMap contains entries for all rules
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_2")
	assert.Contains(t, ruleHoldoutsMap, "rule_3")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_2"], 1)
	assert.Len(t, ruleHoldoutsMap["rule_3"], 1)
}

// TestMapHoldoutsGlobalVsLocal tests global holdout (nil IncludedRules) vs local holdout
func TestMapHoldoutsGlobalVsLocal(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:     "holdout_global",
			Key:    "global_holdout",
			Status: "Running",
			// IncludedRules is nil (not present in JSON) - this is a global holdout
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
		{
			ID:            "holdout_local",
			Key:           "local_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1"},
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

	holdoutList, _, flagHoldoutsMap, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Verify both holdouts are in the list
	assert.Len(t, holdoutList, 2)

	// Global holdout should be in flagHoldoutsMap for all features
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Contains(t, flagHoldoutsMap, "feature_2")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)
	assert.Equal(t, "holdout_global", flagHoldoutsMap["feature_1"][0].ID)
	assert.Nil(t, flagHoldoutsMap["feature_1"][0].IncludedRules)

	// Local holdout should be in ruleHoldoutsMap for rule_1
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 1)
	assert.Equal(t, "holdout_local", ruleHoldoutsMap["rule_1"][0].ID)
	assert.NotNil(t, ruleHoldoutsMap["rule_1"][0].IncludedRules)
}

// TestMapHoldoutsEmptyIncludedRules tests that empty IncludedRules array is treated as local with no rules
func TestMapHoldoutsEmptyIncludedRules(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_empty",
			Key:           "empty_rules_holdout",
			Status:        "Running",
			IncludedRules: []string{}, // Empty slice, not nil
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

	// Empty IncludedRules should be treated as global holdout (same as nil)
	assert.Len(t, holdoutList, 1)

	// Global holdouts should appear in flagHoldoutsMap
	assert.Contains(t, flagHoldoutsMap, "feature_1")
	assert.Len(t, flagHoldoutsMap["feature_1"], 1)

	// Should NOT appear in ruleHoldoutsMap since empty means no specific rules
	assert.Empty(t, ruleHoldoutsMap)
}

// TestMapHoldoutsMultipleLocalHoldoutsSameRule tests multiple local holdouts targeting the same rule
func TestMapHoldoutsMultipleLocalHoldoutsSameRule(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_1",
			Key:           "local_holdout_1",
			Status:        "Running",
			IncludedRules: []string{"rule_1"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 5000},
			},
		},
		{
			ID:            "holdout_2",
			Key:           "local_holdout_2",
			Status:        "Running",
			IncludedRules: []string{"rule_1"},
			Variations: []datafileEntities.Variation{
				{ID: "var_2", Key: "variation_2"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_2", EndOfRange: 5000},
			},
		},
	}

	featureMap := map[string]entities.Feature{
		"feature_1": {ID: "feature_1", Key: "feature_1"},
	}

	_, _, _, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Both holdouts should be in ruleHoldoutsMap for rule_1
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Len(t, ruleHoldoutsMap["rule_1"], 2)

	// Verify both holdouts are present
	holdoutIDs := []string{ruleHoldoutsMap["rule_1"][0].ID, ruleHoldoutsMap["rule_1"][1].ID}
	assert.Contains(t, holdoutIDs, "holdout_1")
	assert.Contains(t, holdoutIDs, "holdout_2")
}

// TestMapHoldoutsNonRunningHoldouts tests that non-running holdouts are excluded
func TestMapHoldoutsNonRunningLocalHoldouts(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_draft",
			Key:           "draft_holdout",
			Status:        "Draft", // Not "Running"
			IncludedRules: []string{"rule_1"},
			Variations: []datafileEntities.Variation{
				{ID: "var_1", Key: "variation_1"},
			},
			TrafficAllocation: []datafileEntities.TrafficAllocation{
				{EntityID: "var_1", EndOfRange: 10000},
			},
		},
		{
			ID:            "holdout_running",
			Key:           "running_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_2"},
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
	}

	holdoutList, _, _, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// Only running holdout should be included
	assert.Len(t, holdoutList, 1)
	assert.Equal(t, "holdout_running", holdoutList[0].ID)

	// Only running holdout should appear in ruleHoldoutsMap
	assert.NotContains(t, ruleHoldoutsMap, "rule_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_2")
}

// TestMapHoldoutsCrossFlag tests local holdout with rules from multiple flags
func TestMapHoldoutsCrossFlag(t *testing.T) {
	rawHoldouts := []datafileEntities.Holdout{
		{
			ID:            "holdout_cross_flag",
			Key:           "cross_flag_holdout",
			Status:        "Running",
			IncludedRules: []string{"rule_1", "rule_2", "rule_3"}, // Rules from different flags
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

	_, _, _, ruleHoldoutsMap := MapHoldouts(rawHoldouts, featureMap)

	// All three rules should have the cross-flag holdout
	assert.Contains(t, ruleHoldoutsMap, "rule_1")
	assert.Contains(t, ruleHoldoutsMap, "rule_2")
	assert.Contains(t, ruleHoldoutsMap, "rule_3")

	assert.Equal(t, "holdout_cross_flag", ruleHoldoutsMap["rule_1"][0].ID)
	assert.Equal(t, "holdout_cross_flag", ruleHoldoutsMap["rule_2"][0].ID)
	assert.Equal(t, "holdout_cross_flag", ruleHoldoutsMap["rule_3"][0].ID)
}
