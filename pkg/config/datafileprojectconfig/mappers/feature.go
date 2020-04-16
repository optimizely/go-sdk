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

// MapFeatures maps the raw datafile feature flag entities to SDK Feature entities
func MapFeatures(featureFlags []datafileEntities.FeatureFlag, rolloutMap map[string]entities.Rollout, experimentMap map[string]entities.Experiment,
) (featureMap map[string]entities.Feature) {

	featureMap = make(map[string]entities.Feature)
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
				experiment.IsFeatureExperiment = true
				featureExperiments = append(featureExperiments, experiment)
				experimentMap[experimentID] = experiment
			}
		}

		variableMap := map[string]entities.Variable{}
		for _, variable := range featureFlag.Variables {

			realType := variable.Type
			if variable.Type == entities.String && variable.SubType == entities.JSON {
				realType = entities.JSON
			}
			variableMap[variable.Key] = entities.Variable{
				DefaultValue: variable.DefaultValue,
				ID:           variable.ID,
				Key:          variable.Key,
				Type:         realType}

		}

		feature.FeatureExperiments = featureExperiments
		feature.VariableMap = variableMap
		featureMap[featureFlag.Key] = feature
	}
	return featureMap
}
