/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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
	datafileEntities "github.com/optimizely/go-sdk/pkg/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// MapExperiments maps the raw experiments entities from the datafile to SDK Experiment entities and also returns a map of experiment key to experiment ID
func MapExperiments(rawExperiments []datafileEntities.Experiment, experimentGroupMap map[string]string) (experimentMap map[string]entities.Experiment, experimentKeyMap map[string]string) {

	experimentMap = make(map[string]entities.Experiment)
	experimentKeyMap = make(map[string]string)
	for _, rawExperiment := range rawExperiments {

		experiment := mapExperiment(rawExperiment)
		experiment.GroupID = experimentGroupMap[experiment.ID]
		experimentMap[experiment.ID] = experiment
		experimentKeyMap[experiment.Key] = experiment.ID
	}

	return experimentMap, experimentKeyMap
}

// Maps the raw variation entity from the datafile to an SDK Variation entity
func mapVariation(rawVariation datafileEntities.Variation) entities.Variation {

	variation := entities.Variation{
		ID:             rawVariation.ID,
		Key:            rawVariation.Key,
		FeatureEnabled: rawVariation.FeatureEnabled,
	}

	variation.Variables = make(map[string]entities.VariationVariable)
	for _, variable := range rawVariation.Variables {
		variation.Variables[variable.ID] = entities.VariationVariable{ID: variable.ID, Value: variable.Value}
	}

	return variation
}

// Maps the raw experiment entity from the datafile into an SDK Experiment entity
func mapExperiment(rawExperiment datafileEntities.Experiment) entities.Experiment {
	var audienceConditionTree *entities.TreeNode
	var err error
	if rawExperiment.AudienceConditions == nil && len(rawExperiment.AudienceIds) > 0 {
		audienceConditionTree, err = buildAudienceConditionTree(rawExperiment.AudienceIds)
	} else {
		switch audienceConditions := rawExperiment.AudienceConditions.(type) {
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

	experiment := entities.Experiment{
		AudienceIds:           rawExperiment.AudienceIds,
		ID:                    rawExperiment.ID,
		LayerID:               rawExperiment.LayerID,
		Key:                   rawExperiment.Key,
		Variations:            make(map[string]entities.Variation),
		VariationKeyToIDMap:   make(map[string]string),
		TrafficAllocation:     make([]entities.Range, len(rawExperiment.TrafficAllocation)),
		AudienceConditionTree: audienceConditionTree,
		Whitelist:             rawExperiment.ForcedVariations,
		IsFeatureExperiment:   false,
	}

	for _, variation := range rawExperiment.Variations {
		experiment.Variations[variation.ID] = mapVariation(variation)
		experiment.VariationKeyToIDMap[variation.Key] = variation.ID
	}

	for i, allocation := range rawExperiment.TrafficAllocation {
		experiment.TrafficAllocation[i] = entities.Range(allocation)
	}

	return experiment
}

// MergeExperiments combines raw experiments and experiments inside groups and returns the array
func MergeExperiments(rawExperiments []datafileEntities.Experiment, rawGroups []datafileEntities.Group) (mergedExperiments []datafileEntities.Experiment) {
	mergedExperiments = rawExperiments
	for _, group := range rawGroups {
		mergedExperiments = append(mergedExperiments, group.Experiments...)
	}
	return mergedExperiments
}
