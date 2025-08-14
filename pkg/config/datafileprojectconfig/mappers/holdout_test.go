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

func TestMapHoldouts(t *testing.T) {
	// Create test holdouts - minimal fields only
	holdouts := []datafileEntities.Holdout{
		{
			ExperimentCore: datafileEntities.ExperimentCore{
				ID:  "holdout_1",
				Key: "global_holdout",
			},
			Status:        datafileEntities.HoldoutStatusRunning,
			IncludedFlags: []string{}, // Global holdout
			ExcludedFlags: []string{"feature_3"},
		},
		{
			ExperimentCore: datafileEntities.ExperimentCore{
				ID:  "holdout_2",
				Key: "feature_specific_holdout",
			},
			Status:        datafileEntities.HoldoutStatusRunning,
			IncludedFlags: []string{"feature_1", "feature_2"},
			ExcludedFlags: []string{},
		},
	}

	audienceMap := map[string]entities.Audience{
		"audience1": {
			ID:   "audience1",
			Name: "Test Audience",
		},
	}

	// Map holdouts
	holdoutMaps := MapHoldouts(holdouts, audienceMap)

	// Verify holdout ID map
	assert.Len(t, holdoutMaps.HoldoutIDMap, 2)
	assert.Contains(t, holdoutMaps.HoldoutIDMap, "holdout_1")
	assert.Contains(t, holdoutMaps.HoldoutIDMap, "holdout_2")

	// Verify global holdouts
	assert.Len(t, holdoutMaps.GlobalHoldouts, 1)
	assert.Equal(t, "holdout_1", holdoutMaps.GlobalHoldouts[0].ID)

	// Verify included holdouts
	assert.Len(t, holdoutMaps.IncludedHoldouts["feature_1"], 1)
	assert.Equal(t, "holdout_2", holdoutMaps.IncludedHoldouts["feature_1"][0].ID)
	assert.Len(t, holdoutMaps.IncludedHoldouts["feature_2"], 1)
	assert.Equal(t, "holdout_2", holdoutMaps.IncludedHoldouts["feature_2"][0].ID)

	// Verify excluded holdouts
	assert.Len(t, holdoutMaps.ExcludedHoldouts["feature_3"], 1)
	assert.Equal(t, "holdout_1", holdoutMaps.ExcludedHoldouts["feature_3"][0].ID)
}

func TestGetHoldoutsForFlag(t *testing.T) {
	// Create test holdouts similar to JavaScript SDK test
	holdout1 := entities.Holdout{
		ID:            "holdout_1",
		Key:           "global_holdout",
		IncludedFlags: []string{}, // Global
		ExcludedFlags: []string{},
	}

	holdout2 := entities.Holdout{
		ID:            "holdout_2",
		Key:           "global_with_exclusion",
		IncludedFlags: []string{}, // Global
		ExcludedFlags: []string{"feature_3"},
	}

	holdout3 := entities.Holdout{
		ID:            "holdout_3",
		Key:           "feature_specific",
		IncludedFlags: []string{"feature_1"},
		ExcludedFlags: []string{},
	}

	holdoutMaps := HoldoutMaps{
		HoldoutIDMap: map[string]entities.Holdout{
			"holdout_1": holdout1,
			"holdout_2": holdout2,
			"holdout_3": holdout3,
		},
		GlobalHoldouts: []entities.Holdout{holdout1, holdout2},
		IncludedHoldouts: map[string][]entities.Holdout{
			"feature_1": {holdout3},
		},
		ExcludedHoldouts: map[string][]entities.Holdout{
			"feature_3": {holdout2},
		},
		FlagHoldoutsMap: make(map[string][]string),
	}

	// Test feature_1: should get global holdouts + specifically included
	holdoutIDs := GetHoldoutsForFlag("feature_1", holdoutMaps)
	assert.Len(t, holdoutIDs, 3)
	assert.Contains(t, holdoutIDs, "holdout_1")
	assert.Contains(t, holdoutIDs, "holdout_2")
	assert.Contains(t, holdoutIDs, "holdout_3")

	// Test feature_2: should get only global holdouts
	holdoutIDs = GetHoldoutsForFlag("feature_2", holdoutMaps)
	assert.Len(t, holdoutIDs, 2)
	assert.Contains(t, holdoutIDs, "holdout_1")
	assert.Contains(t, holdoutIDs, "holdout_2")

	// Test feature_3: should get global holdouts minus excluded
	holdoutIDs = GetHoldoutsForFlag("feature_3", holdoutMaps)
	assert.Len(t, holdoutIDs, 1)
	assert.Contains(t, holdoutIDs, "holdout_1")
	assert.NotContains(t, holdoutIDs, "holdout_2") // Excluded

	// Test caching - second call should return cached result
	cachedHoldoutIDs := GetHoldoutsForFlag("feature_3", holdoutMaps)
	assert.Equal(t, holdoutIDs, cachedHoldoutIDs)
}
