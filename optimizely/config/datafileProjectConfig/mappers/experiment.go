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

package mappers

import (
	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// MapExperiments maps the raw experiments entities from the datafile to SDK Experiment entities and also returns a map of experiment key to experiment ID
func MapExperiments(rawExperiments []datafileEntities.Experiment) (map[string]entities.Experiment, map[string]string) {

	experimentMap := make(map[string]entities.Experiment)
	experimentKeyMap := make(map[string]string)
	for _, rawExperiment := range rawExperiments {
		experiment := entities.Experiment{
			AudienceIds:       rawExperiment.AudienceIds,
			ID:                rawExperiment.ID,
			Key:               rawExperiment.Key,
			TrafficAllocation: make([]entities.Range, len(rawExperiment.TrafficAllocation)),
			Variations:        make(map[string]entities.Variation),
		}

		for _, variation := range rawExperiment.Variations {
			experiment.Variations[variation.ID] = mapVariation(variation)
		}

		for i, allocation := range rawExperiment.TrafficAllocation {
			experiment.TrafficAllocation[i] = entities.Range(allocation)
		}

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
	return variation
}
