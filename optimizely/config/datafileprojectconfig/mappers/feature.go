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
	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileprojectconfig/entities"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// MapFeatures maps the raw datafile feature flag entities to SDK Feature entities
func MapFeatures(featureFlags []datafileEntities.FeatureFlag, rolloutMap map[string]entities.Rollout, experimentMap map[string]entities.Experiment,
) map[string]entities.Feature {

	featureMap := make(map[string]entities.Feature)
	for _, featureFlag := range featureFlags {
		feature := entities.Feature{
			Key: featureFlag.Key,
			ID:  featureFlag.ID,
		}
		if rollout, ok := rolloutMap[featureFlag.RolloutID]; ok {
			feature.Rollout = rollout
		}
		featureExperiments := []entities.Experiment{}
		for _, experimentID := range featureFlag.ExperimentIDs {
			if experiment, ok := experimentMap[experimentID]; ok {
				featureExperiments = append(featureExperiments, experiment)
			}
		}

		feature.FeatureExperiments = featureExperiments
		featureMap[featureFlag.Key] = feature
	}
	return featureMap
}

// MapFeatureFlags maps the raw datafile feature flag entities to their keys
func MapFeatureFlags(featureFlags []datafileEntities.FeatureFlag) map[string]datafileEntities.FeatureFlag {

	featureFlagsMap := make(map[string]datafileEntities.FeatureFlag)
	for _, featureFlag := range featureFlags {
		featureFlagsMap[featureFlag.Key] = featureFlag
	}
	return featureFlagsMap
}
