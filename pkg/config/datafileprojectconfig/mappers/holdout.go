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

// Package mappers  ...
package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// HoldoutMaps contains the different holdout mappings for efficient lookup
type HoldoutMaps struct {
	HoldoutIDMap     map[string]entities.Holdout   // Map holdout ID to holdout
	GlobalHoldouts   []entities.Holdout            // Holdouts with no specific flag inclusion
	IncludedHoldouts map[string][]entities.Holdout // Map flag ID to holdouts that include it
	ExcludedHoldouts map[string][]entities.Holdout // Map flag ID to holdouts that exclude it
	FlagHoldoutsMap  map[string][]string           // Cached map of flag ID to holdout IDs
}

// MapHoldouts maps the raw datafile holdout entities to SDK Holdout entities
// and creates the necessary mappings for efficient holdout lookup
func MapHoldouts(holdouts []datafileEntities.Holdout, audienceMap map[string]entities.Audience) HoldoutMaps {
	holdoutMaps := HoldoutMaps{
		HoldoutIDMap:     make(map[string]entities.Holdout),
		GlobalHoldouts:   []entities.Holdout{},
		IncludedHoldouts: make(map[string][]entities.Holdout),
		ExcludedHoldouts: make(map[string][]entities.Holdout),
		FlagHoldoutsMap:  make(map[string][]string),
	}

	for _, datafileHoldout := range holdouts {
		// Create minimal runtime holdout entity - only what's needed for flag dependency
		holdout := entities.Holdout{
			ID:            datafileHoldout.ID,
			Key:           datafileHoldout.Key,
			Status:        entities.HoldoutStatus(datafileHoldout.Status),
			IncludedFlags: datafileHoldout.IncludedFlags,
			ExcludedFlags: datafileHoldout.ExcludedFlags,
		}

		// Add to ID map
		holdoutMaps.HoldoutIDMap[holdout.ID] = holdout

		// Categorize holdouts based on flag targeting
		if len(datafileHoldout.IncludedFlags) == 0 {
			// This is a global holdout (applies to all flags unless excluded)
			holdoutMaps.GlobalHoldouts = append(holdoutMaps.GlobalHoldouts, holdout)

			// Add to excluded flags map
			for _, flagID := range datafileHoldout.ExcludedFlags {
				holdoutMaps.ExcludedHoldouts[flagID] = append(holdoutMaps.ExcludedHoldouts[flagID], holdout)
			}
		} else {
			// This holdout specifically includes certain flags
			for _, flagID := range datafileHoldout.IncludedFlags {
				holdoutMaps.IncludedHoldouts[flagID] = append(holdoutMaps.IncludedHoldouts[flagID], holdout)
			}
		}
	}

	return holdoutMaps
}

// GetHoldoutsForFlag returns the holdout IDs that apply to a specific flag
// This follows the logic from JavaScript SDK: global holdouts (minus excluded) + specifically included
func GetHoldoutsForFlag(flagID string, holdoutMaps HoldoutMaps) []string {
	// Check cache first
	if cachedHoldoutIDs, exists := holdoutMaps.FlagHoldoutsMap[flagID]; exists {
		return cachedHoldoutIDs
	}

	holdoutIDs := []string{}

	// Add global holdouts that don't exclude this flag
	for _, holdout := range holdoutMaps.GlobalHoldouts {
		isExcluded := false
		for _, excludedFlagID := range holdout.ExcludedFlags {
			if excludedFlagID == flagID {
				isExcluded = true
				break
			}
		}
		if !isExcluded {
			holdoutIDs = append(holdoutIDs, holdout.ID)
		}
	}

	// Add holdouts that specifically include this flag
	if includedHoldouts, exists := holdoutMaps.IncludedHoldouts[flagID]; exists {
		for _, holdout := range includedHoldouts {
			holdoutIDs = append(holdoutIDs, holdout.ID)
		}
	}

	// Cache the result
	holdoutMaps.FlagHoldoutsMap[flagID] = holdoutIDs

	return holdoutIDs
}