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

// Package mappers ...
package mappers

import (
	"fmt"

	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// MapHoldouts maps the two top-level datafile holdout sections to SDK Holdout entities.
//
// Section membership is the sole signal for scope (FSSDK-12760):
//   - `holdouts` entries are global; any `includedRules` field on them is stripped.
//   - `localHoldouts` entries are local and MUST carry a non-nil `includedRules` list;
//     entries missing it are logged via `logger` and excluded (no fallback to global).
//
// Returns the combined entity list, an id->entity map, the global-only list, and a
// per-rule map for local holdouts.
func MapHoldouts(
	globalHoldoutsRaw []datafileEntities.Holdout,
	localHoldoutsRaw []datafileEntities.Holdout,
	logger logging.OptimizelyLogProducer,
) (
	holdoutList []entities.Holdout,
	holdoutIDMap map[string]entities.Holdout,
	globalHoldouts []entities.Holdout,
	ruleHoldoutsMap map[string][]entities.Holdout,
) {
	holdoutList = []entities.Holdout{}
	holdoutIDMap = make(map[string]entities.Holdout)
	globalHoldouts = []entities.Holdout{}
	ruleHoldoutsMap = make(map[string][]entities.Holdout)

	// Process global holdouts: drop any `IncludedRules` so section membership alone
	// determines scope, even if the datafile incorrectly includes one.
	for _, holdout := range globalHoldoutsRaw {
		if holdout.Status != string(entities.HoldoutStatusRunning) {
			continue
		}

		sanitized := holdout
		sanitized.IncludedRules = nil

		mappedHoldout := mapHoldout(sanitized)
		holdoutList = append(holdoutList, mappedHoldout)
		holdoutIDMap[mappedHoldout.ID] = mappedHoldout
		globalHoldouts = append(globalHoldouts, mappedHoldout)
	}

	// Process local holdouts: `IncludedRules` is REQUIRED. Entries missing it (nil
	// pointer) are invalid per spec — log and skip, never promote to global.
	for _, holdout := range localHoldoutsRaw {
		if holdout.Status != string(entities.HoldoutStatusRunning) {
			continue
		}

		if holdout.IncludedRules == nil {
			if logger != nil {
				logger.Error(
					fmt.Sprintf(
						"Local holdout %q is missing required \"includedRules\" field and will be excluded from evaluation.",
						holdoutLabel(holdout),
					),
					nil,
				)
			}
			continue
		}

		mappedHoldout := mapHoldout(holdout)
		holdoutList = append(holdoutList, mappedHoldout)
		holdoutIDMap[mappedHoldout.ID] = mappedHoldout

		// Register the local holdout for each rule it targets. An empty IncludedRules
		// slice is valid (matches no rules) but is not promoted to global.
		for _, ruleID := range *mappedHoldout.IncludedRules {
			ruleHoldoutsMap[ruleID] = append(ruleHoldoutsMap[ruleID], mappedHoldout)
		}
	}

	return holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap
}

// holdoutLabel returns a stable, user-facing label for an invalid local holdout
// log message. Prefers the holdout key (human-readable) and falls back to the id.
func holdoutLabel(h datafileEntities.Holdout) string {
	if h.Key != "" {
		return h.Key
	}
	if h.ID != "" {
		return h.ID
	}
	return "<unknown>"
}

func mapHoldout(datafileHoldout datafileEntities.Holdout) entities.Holdout {
	var audienceConditionTree *entities.TreeNode
	var err error

	// Build audience condition tree similar to experiments
	if datafileHoldout.AudienceConditions == nil && len(datafileHoldout.AudienceIds) > 0 {
		audienceConditionTree, err = buildAudienceConditionTree(datafileHoldout.AudienceIds)
	} else {
		switch audienceConditions := datafileHoldout.AudienceConditions.(type) {
		case []interface{}:
			if len(audienceConditions) > 0 {
				audienceConditionTree, err = buildAudienceConditionTree(audienceConditions)
			}
		case string:
			if audienceConditions != "" {
				audienceConditionTree, err = buildAudienceConditionTree([]string{audienceConditions})
			}
		default:
		}
	}
	if err != nil {
		// @TODO: handle error
		func() {}() // cheat the linters
	}

	// Map variations
	variations := make(map[string]entities.Variation)
	for _, datafileVariation := range datafileHoldout.Variations {
		variation := mapVariation(datafileVariation)
		variations[variation.ID] = variation
	}

	// Map traffic allocations
	trafficAllocation := make([]entities.Range, len(datafileHoldout.TrafficAllocation))
	for i, allocation := range datafileHoldout.TrafficAllocation {
		trafficAllocation[i] = entities.Range{
			EntityID:   allocation.EntityID,
			EndOfRange: allocation.EndOfRange,
		}
	}

	return entities.Holdout{
		ID:                    datafileHoldout.ID,
		Key:                   datafileHoldout.Key,
		Status:                entities.HoldoutStatus(datafileHoldout.Status),
		AudienceIds:           datafileHoldout.AudienceIds,
		AudienceConditions:    datafileHoldout.AudienceConditions,
		Variations:            variations,
		TrafficAllocation:     trafficAllocation,
		AudienceConditionTree: audienceConditionTree,
		IncludedRules:         datafileHoldout.IncludedRules,
	}
}
