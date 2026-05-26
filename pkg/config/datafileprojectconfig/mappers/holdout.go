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
	datafileEntities "github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// MapHoldouts maps the raw datafile holdout entities to SDK Holdout entities.
// Global holdouts (IncludedRules == nil) are returned in globalHoldouts for flag-level evaluation.
// Local holdouts (IncludedRules != nil) are indexed by rule ID in ruleHoldoutsMap.
func MapHoldouts(holdouts []datafileEntities.Holdout) (
	holdoutList []entities.Holdout,
	holdoutIDMap map[string]entities.Holdout,
	globalHoldouts []entities.Holdout,
	ruleHoldoutsMap map[string][]entities.Holdout,
) {
	holdoutList = []entities.Holdout{}
	holdoutIDMap = make(map[string]entities.Holdout)
	globalHoldouts = []entities.Holdout{}
	ruleHoldoutsMap = make(map[string][]entities.Holdout)

	for _, holdout := range holdouts {
		// Only process running holdouts
		if holdout.Status != string(entities.HoldoutStatusRunning) {
			continue
		}

		mappedHoldout := mapHoldout(holdout)
		holdoutList = append(holdoutList, mappedHoldout)
		holdoutIDMap[holdout.ID] = mappedHoldout

		if mappedHoldout.IsGlobal() {
			globalHoldouts = append(globalHoldouts, mappedHoldout)
		} else {
			// Local holdout: applies only to the specified rule IDs
			for _, ruleID := range *mappedHoldout.IncludedRules {
				ruleHoldoutsMap[ruleID] = append(ruleHoldoutsMap[ruleID], mappedHoldout)
			}
		}
	}

	return holdoutList, holdoutIDMap, globalHoldouts, ruleHoldoutsMap
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
