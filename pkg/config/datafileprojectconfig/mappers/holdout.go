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

// MapHoldouts maps the raw datafile holdout entities to SDK Holdout entities
// and organizes them by flag relationships
func MapHoldouts(holdouts []datafileEntities.Holdout, featureMap map[string]entities.Feature) (
	holdoutList []entities.Holdout,
	holdoutIDMap map[string]entities.Holdout,
	flagHoldoutsMap map[string][]entities.Holdout,
	experimentHoldoutsMap map[string][]entities.Holdout,
) {
	holdoutList = []entities.Holdout{}
	holdoutIDMap = make(map[string]entities.Holdout)
	flagHoldoutsMap = make(map[string][]entities.Holdout)
	experimentHoldoutsMap = make(map[string][]entities.Holdout)

	globalHoldouts := []entities.Holdout{}
	includedHoldouts := make(map[string][]entities.Holdout)
	excludedHoldouts := make(map[string][]entities.Holdout)

	for _, holdout := range holdouts {
		// Only process running holdouts
		if holdout.Status != string(entities.HoldoutStatusRunning) {
			continue
		}

		mappedHoldout := mapHoldout(holdout)
		holdoutList = append(holdoutList, mappedHoldout)
		holdoutIDMap[holdout.ID] = mappedHoldout

		// Classify holdout by flag relationships
		if len(holdout.IncludedFlags) == 0 {
			// Global holdout - applies to all flags except excluded
			globalHoldouts = append(globalHoldouts, mappedHoldout)

			// Track exclusions
			for _, flagID := range holdout.ExcludedFlags {
				excludedHoldouts[flagID] = append(excludedHoldouts[flagID], mappedHoldout)
			}
		} else {
			// Specific holdout - applies only to included flags
			for _, flagID := range holdout.IncludedFlags {
				includedHoldouts[flagID] = append(includedHoldouts[flagID], mappedHoldout)
			}
		}

		// Build experiment-to-holdout mapping
		for _, experimentID := range holdout.Experiments {
			experimentHoldoutsMap[experimentID] = append(experimentHoldoutsMap[experimentID], mappedHoldout)
		}
	}

	// Build flagHoldoutsMap by combining global and specific holdouts
	// Global holdouts take precedence (evaluated first), then specific holdouts
	for _, feature := range featureMap {
		flagKey := feature.Key
		flagID := feature.ID
		applicableHoldouts := []entities.Holdout{}

		// Add global holdouts first (if not excluded) - they take precedence
		if _, exists := excludedHoldouts[flagID]; !exists {
			applicableHoldouts = append(applicableHoldouts, globalHoldouts...)
		}

		// Add specifically included holdouts second
		if included, exists := includedHoldouts[flagID]; exists {
			applicableHoldouts = append(applicableHoldouts, included...)
		}

		if len(applicableHoldouts) > 0 {
			flagHoldoutsMap[flagKey] = applicableHoldouts
		}
	}

	return holdoutList, holdoutIDMap, flagHoldoutsMap, experimentHoldoutsMap
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
		Experiments:           datafileHoldout.Experiments,
	}
}
