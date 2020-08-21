/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

// Package config //
package config

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptimizelyConfig is a snapshot of the experiments and features in the project config
type OptimizelyConfig struct {
	Revision       string                          `json:"revision"`
	ExperimentsMap map[string]OptimizelyExperiment `json:"experimentsMap"`
	FeaturesMap    map[string]OptimizelyFeature    `json:"featuresMap"`
	datafile       string
}

// GetDatafile returns a string representation of the environment's datafile
func (c OptimizelyConfig) GetDatafile() string {
	return c.datafile
}

// OptimizelyExperiment has experiment info
type OptimizelyExperiment struct {
	ID            string                         `json:"id"`
	Key           string                         `json:"key"`
	VariationsMap map[string]OptimizelyVariation `json:"variationsMap"`
}

// OptimizelyFeature has feature info
type OptimizelyFeature struct {
	ID             string                          `json:"id"`
	Key            string                          `json:"key"`
	ExperimentsMap map[string]OptimizelyExperiment `json:"experimentsMap"`
	VariablesMap   map[string]OptimizelyVariable   `json:"variablesMap"`
}

// OptimizelyVariation has variation info
type OptimizelyVariation struct {
	ID             string                        `json:"id"`
	Key            string                        `json:"key"`
	FeatureEnabled bool                          `json:"featureEnabled"`
	VariablesMap   map[string]OptimizelyVariable `json:"variablesMap"`
}

// OptimizelyVariable has variable info
type OptimizelyVariable struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func getVariableByIDMap(features []entities.Feature) (variableByIDMap map[string]entities.Variable) {
	variableByIDMap = map[string]entities.Variable{}
	for _, feature := range features {
		for _, variable := range feature.VariableMap {
			variableByIDMap[variable.ID] = variable
		}
	}
	return variableByIDMap
}

func getExperimentVariablesMap(features []entities.Feature) (experimentVariableMap map[string]map[string]OptimizelyVariable) {
	experimentVariableMap = map[string]map[string]OptimizelyVariable{}
	for _, feature := range features {

		var optimizelyVariableMap = map[string]OptimizelyVariable{}
		for _, variable := range feature.VariableMap {
			optimizelyVariableMap[variable.Key] = OptimizelyVariable{Key: variable.Key, ID: variable.ID, Value: variable.DefaultValue, Type: string(variable.Type)}

		}
		for _, experiment := range feature.FeatureExperiments {
			experimentVariableMap[experiment.Key] = optimizelyVariableMap
		}
	}
	return experimentVariableMap
}

func getExperimentMap(features []entities.Feature, experiments []entities.Experiment, variableByIDMap map[string]entities.Variable) (optlyExperimentMap map[string]OptimizelyExperiment) {

	optlyExperimentMap = map[string]OptimizelyExperiment{}
	experimentVariablesMap := getExperimentVariablesMap(features)

	for _, experiment := range experiments {
		var optlyVariationsMap = map[string]OptimizelyVariation{}
		for _, variation := range experiment.Variations {
			var optlyVariablesMap = map[string]OptimizelyVariable{}

			if variableMap, ok := experimentVariablesMap[experiment.Key]; ok {
				for index, element := range variableMap { // copy by value
					optlyVariablesMap[index] = element
				}
			}

			for _, variable := range variation.Variables {
				if experiment.IsFeatureExperiment && variation.FeatureEnabled {
					if convertedVariable, ok := variableByIDMap[variable.ID]; ok {
						optlyVariable := OptimizelyVariable{Key: convertedVariable.Key, ID: convertedVariable.ID,
							Type: string(convertedVariable.Type), Value: variable.Value}
						optlyVariablesMap[convertedVariable.Key] = optlyVariable
					}
				}
			}
			optVariation := OptimizelyVariation{ID: variation.ID, Key: variation.Key, VariablesMap: optlyVariablesMap, FeatureEnabled: variation.FeatureEnabled}
			optlyVariationsMap[variation.Key] = optVariation
		}
		optlyExperiment := OptimizelyExperiment{ID: experiment.ID, Key: experiment.Key, VariationsMap: optlyVariationsMap}
		optlyExperimentMap[experiment.Key] = optlyExperiment
	}
	return optlyExperimentMap
}

func getFeatureMap(features []entities.Feature, experimentsMap map[string]OptimizelyExperiment) (optlyFeatureMap map[string]OptimizelyFeature) {

	optlyFeatureMap = map[string]OptimizelyFeature{}

	for _, feature := range features {

		var optlyFeatureVariablesMap = map[string]OptimizelyVariable{}
		for _, featureVarible := range feature.VariableMap {
			optlyVariable := OptimizelyVariable{Key: featureVarible.Key, ID: featureVarible.ID,
				Type: string(featureVarible.Type), Value: featureVarible.DefaultValue}
			optlyFeatureVariablesMap[featureVarible.Key] = optlyVariable
		}

		var optlyExperimentMap = map[string]OptimizelyExperiment{}
		for _, experiment := range feature.FeatureExperiments {
			optlyExperimentMap[experiment.Key] = experimentsMap[experiment.Key]
		}

		optlyFeature := OptimizelyFeature{ID: feature.ID, Key: feature.Key, ExperimentsMap: optlyExperimentMap, VariablesMap: optlyFeatureVariablesMap}
		optlyFeatureMap[feature.Key] = optlyFeature

	}
	return optlyFeatureMap
}

// NewOptimizelyConfig constructs OptimizelyConfig object
func NewOptimizelyConfig(projConfig ProjectConfig) *OptimizelyConfig {

	if projConfig == nil {
		return nil
	}
	featuresList := projConfig.GetFeatureList()
	experimentsList := projConfig.GetExperimentList()
	revision := projConfig.GetRevision()

	optimizelyConfig := &OptimizelyConfig{}

	variableByIDMap := getVariableByIDMap(featuresList)

	optimizelyConfig.ExperimentsMap = getExperimentMap(featuresList, experimentsList, variableByIDMap)
	optimizelyConfig.FeaturesMap = getFeatureMap(featuresList, optimizelyConfig.ExperimentsMap)
	optimizelyConfig.Revision = revision
	optimizelyConfig.datafile = projConfig.GetDatafile()

	return optimizelyConfig
}
