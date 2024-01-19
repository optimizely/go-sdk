/****************************************************************************
 * Copyright 2019-2021, Optimizely, Inc. and contributors                   *
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/optimizely/go-sdk/v2/pkg/config/datafileprojectconfig/mappers"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

// OptimizelyConfig is a snapshot of the experiments and features in the project config
type OptimizelyConfig struct {
	EnvironmentKey string `json:"environmentKey"`
	SdkKey         string `json:"sdkKey"`
	Revision       string `json:"revision"`

	// This experimentsMap is for experiments of legacy projects only.
	// For flag projects, experiment keys are not guaranteed to be unique
	// across multiple flags, so this map may not include all experiments
	// when keys conflict.
	ExperimentsMap map[string]OptimizelyExperiment `json:"experimentsMap"`

	FeaturesMap map[string]OptimizelyFeature `json:"featuresMap"`
	Attributes  []OptimizelyAttribute        `json:"attributes"`
	Audiences   []OptimizelyAudience         `json:"audiences"`
	Events      []OptimizelyEvent            `json:"events"`
	datafile    string
}

// GetDatafile returns a string representation of the environment's datafile
func (c OptimizelyConfig) GetDatafile() string {
	return c.datafile
}

// OptimizelyExperiment has experiment info
type OptimizelyExperiment struct {
	ID            string                         `json:"id"`
	Key           string                         `json:"key"`
	Audiences     string                         `json:"audiences"`
	VariationsMap map[string]OptimizelyVariation `json:"variationsMap"`
}

// OptimizelyAttribute has attribute info
type OptimizelyAttribute struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// OptimizelyAudience has audience info
type OptimizelyAudience struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Conditions string `json:"conditions"`
}

// OptimizelyEvent has event info
type OptimizelyEvent struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}

// OptimizelyFeature has feature info
type OptimizelyFeature struct {
	ID              string                        `json:"id"`
	Key             string                        `json:"key"`
	ExperimentRules []OptimizelyExperiment        `json:"experimentRules"`
	DeliveryRules   []OptimizelyExperiment        `json:"deliveryRules"`
	VariablesMap    map[string]OptimizelyVariable `json:"variablesMap"`

	// Deprecated: Use experimentRules and deliveryRules
	ExperimentsMap map[string]OptimizelyExperiment `json:"experimentsMap"`
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

func getAudiences(audiencesList []entities.Audience) []OptimizelyAudience {
	audiences := []OptimizelyAudience{}
	for _, audience := range audiencesList {
		if audience.ID != "$opt_dummy_audience" {
			optlyAudience := OptimizelyAudience{
				ID:         audience.ID,
				Name:       audience.Name,
				Conditions: "",
			}
			switch item := audience.Conditions.(type) {
			case string:
				optlyAudience.Conditions = item
			case interface{}:
				jsonConditionsString, err := json.Marshal(item)
				if err == nil {
					optlyAudience.Conditions = string(jsonConditionsString)
				}
			}
			audiences = append(audiences, optlyAudience)
		}
	}
	return audiences
}

func getSerializedAudiences(conditions interface{}, audiencesByID map[string]entities.Audience) string {
	operators := map[string]bool{}
	defaultOperators := []mappers.OperatorType{mappers.And, mappers.Or, mappers.Not}
	for _, operator := range defaultOperators {
		operators[string(operator)] = true
	}
	var serializedAudience string

	if conditions != nil {
		var cond string
		if conditionsList, ok := conditions.([]interface{}); ok {
			for _, condition := range conditionsList {
				var subAudience string
				// Checks if item is list of conditions means it is sub audience
				switch item := condition.(type) {
				case []interface{}:
					subAudience = getSerializedAudiences(item, audiencesByID)
					subAudience = fmt.Sprintf("(%s)", subAudience)
				case string:
					if operators[item] {
						cond = strings.ToUpper(item)
					} else {
						// Checks if item is audience id
						var audienceName = item
						if audience, ok := audiencesByID[item]; ok {
							audienceName = audience.Name
						}
						// if audience condition is "NOT" then add "NOT" at start.
						// Otherwise check if there is already audience id in serializedAudience
						// then append condition between serializedAudience and item
						if serializedAudience != "" || cond == (strings.ToUpper(string(mappers.Not))) {
							if cond == "" {
								cond = strings.ToUpper(string(mappers.Or))
							}
							if serializedAudience == "" {
								serializedAudience = fmt.Sprintf(`%s %q`, cond, audiencesByID[item].Name)
							} else {
								serializedAudience += fmt.Sprintf(` %s %q`, cond, audienceName)
							}
						} else {
							serializedAudience = fmt.Sprintf(`%q`, audienceName)
						}
					}
				default:
				}
				// Had to create a different method to reduce cyclomatic complexity
				evaluateSubAudience(&subAudience, &serializedAudience, &cond)
			}
		}
	}
	return serializedAudience
}

func evaluateSubAudience(subAudience, serializedAudience, cond *string) {
	// Checks if sub audience is empty or not
	if *subAudience != "" {
		if *serializedAudience != "" || *cond == (strings.ToUpper(string(mappers.Not))) {
			if *cond == "" {
				*cond = (strings.ToUpper(string(mappers.Or)))
			}
			if *serializedAudience == "" {
				*serializedAudience = fmt.Sprintf(`%s %s`, *cond, *subAudience)
			} else {
				*serializedAudience += fmt.Sprintf(` %s %s`, *cond, *subAudience)
			}
		} else {
			*serializedAudience += *subAudience
		}
	}
}

func getExperimentAudiences(experiment entities.Experiment, audiencesByID map[string]entities.Audience) string {
	return getSerializedAudiences(experiment.AudienceConditions, audiencesByID)
}

func mergeFeatureVariables(feature entities.Feature, variableIDMap map[string]entities.Variable, featureVariableUsages map[string]entities.VariationVariable, isFeatureEnabled bool) map[string]OptimizelyVariable {
	variablesMap := map[string]OptimizelyVariable{}
	for _, featureVariable := range feature.VariableMap {
		variablesMap[featureVariable.Key] = OptimizelyVariable{
			ID:    featureVariable.ID,
			Key:   featureVariable.Key,
			Type:  string(featureVariable.Type),
			Value: featureVariable.DefaultValue,
		}
	}
	if len(featureVariableUsages) > 0 {
		for _, featureVariableUsage := range featureVariableUsages {
			var defaultVariable = variableIDMap[featureVariableUsage.ID]
			var value = defaultVariable.DefaultValue
			if isFeatureEnabled {
				value = featureVariableUsage.Value
			}
			variablesMap[defaultVariable.Key] = OptimizelyVariable{
				ID:    featureVariableUsage.ID,
				Key:   defaultVariable.Key,
				Type:  string(defaultVariable.Type),
				Value: value,
			}
		}
	}
	return variablesMap
}

func getVariationsMap(feature entities.Feature, variations map[string]entities.Variation, variableIDMap map[string]entities.Variable) map[string]OptimizelyVariation {
	variationsMap := map[string]OptimizelyVariation{}
	for _, variation := range variations {
		variablesMap := mergeFeatureVariables(feature, variableIDMap, variation.Variables, variation.FeatureEnabled)
		variationsMap[variation.Key] = OptimizelyVariation{
			ID:             variation.ID,
			Key:            variation.Key,
			FeatureEnabled: variation.FeatureEnabled,
			VariablesMap:   variablesMap,
		}
	}
	return variationsMap
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

func getDeliveryRules(variableByIDMap map[string]entities.Variable, audiencesByID map[string]entities.Audience, feature entities.Feature, experiments []entities.Experiment) []OptimizelyExperiment {
	optimizelyExpriments := []OptimizelyExperiment{}
	for _, experiment := range experiments {
		optimizelyExpriments = append(optimizelyExpriments, OptimizelyExperiment{
			ID:            experiment.ID,
			Key:           experiment.Key,
			Audiences:     getExperimentAudiences(experiment, audiencesByID),
			VariationsMap: getVariationsMap(feature, experiment.Variations, variableByIDMap),
		})
	}
	return optimizelyExpriments
}

func getRolloutExperimentsIdsMap(rolloutIDMap map[string]entities.Rollout) map[string]bool {
	var rolloutExperimentIdsMap = map[string]bool{}
	for _, rollout := range rolloutIDMap {
		for _, experiment := range rollout.Experiments {
			rolloutExperimentIdsMap[experiment.ID] = true
		}
	}
	return rolloutExperimentIdsMap
}

func getExperimentFeatureMap(features []entities.Feature) map[string][]string {
	experimentFeatureMap := map[string][]string{}
	for _, feat := range features {
		for _, exp := range feat.FeatureExperiments {
			if featureIds, ok := experimentFeatureMap[exp.ID]; ok {
				featureIds = append(featureIds, feat.ID)
				experimentFeatureMap[exp.ID] = featureIds
			} else {
				experimentFeatureMap[exp.ID] = []string{feat.ID}
			}
		}
	}
	return experimentFeatureMap
}

func getMappedExperiments(audiencesByID map[string]entities.Audience, experiments []entities.Experiment, features []entities.Feature, featuresMap map[string]entities.Feature, rolloutMap map[string]entities.Rollout) map[string]OptimizelyExperiment {
	mappedExperiments := map[string]OptimizelyExperiment{}
	variableIDMap := getVariableByIDMap(features)
	rolloutExperimentIdsMap := getRolloutExperimentsIdsMap(rolloutMap)
	experimentFeaturesMap := getExperimentFeatureMap(features)

	for _, experiment := range experiments {
		if rolloutExperimentIdsMap[experiment.ID] {
			continue
		}
		featureIds := experimentFeaturesMap[experiment.ID]
		featureID := ""
		if len(featureIds) > 0 {
			featureID = featureIds[0]
		}
		variationsMap := getVariationsMap(featuresMap[featureID], experiment.Variations, variableIDMap)
		mappedExperiments[experiment.ID] = OptimizelyExperiment{
			ID:            experiment.ID,
			Key:           experiment.Key,
			Audiences:     getExperimentAudiences(experiment, audiencesByID),
			VariationsMap: variationsMap,
		}
	}
	return mappedExperiments
}

func getExperimentsKeyMap(mappedExperiments map[string]OptimizelyExperiment) map[string]OptimizelyExperiment {
	experimentKeysMap := map[string]OptimizelyExperiment{}
	for _, exp := range mappedExperiments {
		experimentKeysMap[exp.Key] = exp
	}
	return experimentKeysMap
}

func getFeaturesMap(audiencesByID map[string]entities.Audience, mappedExperiments map[string]OptimizelyExperiment, features []entities.Feature, rolloutIDMap map[string]entities.Rollout, variableByIDMap map[string]entities.Variable) map[string]OptimizelyFeature {
	featuresMap := map[string]OptimizelyFeature{}
	for _, featureFlag := range features {
		featureExperimentMap := map[string]OptimizelyExperiment{}
		experimentRules := []OptimizelyExperiment{}
		for _, expID := range featureFlag.ExperimentIDs {
			if exp, ok := mappedExperiments[expID]; ok {
				featureExperimentMap[exp.Key] = exp
				experimentRules = append(experimentRules, exp)
			}
		}
		optimizelyFeatureVariablesMap := map[string]OptimizelyVariable{}
		for _, variable := range featureFlag.VariableMap {
			optimizelyFeatureVariablesMap[variable.Key] = OptimizelyVariable{
				ID:    variable.ID,
				Key:   variable.Key,
				Type:  string(variable.Type),
				Value: variable.DefaultValue,
			}
		}
		deliveryRules := []OptimizelyExperiment{}
		if rollout, ok := rolloutIDMap[featureFlag.Rollout.ID]; ok {
			deliveryRules = getDeliveryRules(variableByIDMap, audiencesByID, featureFlag, rollout.Experiments)
		}
		featuresMap[featureFlag.Key] = OptimizelyFeature{
			ID:              featureFlag.ID,
			Key:             featureFlag.Key,
			ExperimentRules: experimentRules,
			DeliveryRules:   deliveryRules,
			ExperimentsMap:  featureExperimentMap,
			VariablesMap:    optimizelyFeatureVariablesMap,
		}
	}
	return featuresMap
}

// NewOptimizelyConfig constructs OptimizelyConfig object
func NewOptimizelyConfig(projConfig ProjectConfig) *OptimizelyConfig {

	if projConfig == nil {
		return nil
	}
	optimizelyConfig := &OptimizelyConfig{}
	optimizelyConfig.SdkKey = projConfig.GetSdkKey()
	optimizelyConfig.EnvironmentKey = projConfig.GetEnvironmentKey()
	optimizelyAttributes := []OptimizelyAttribute{}
	for _, attribute := range projConfig.GetAttributes() {
		optimizelyAttributes = append(optimizelyAttributes, OptimizelyAttribute(attribute))
	}
	optimizelyConfig.Attributes = optimizelyAttributes
	optimizelyConfig.Audiences = getAudiences(projConfig.GetAudienceList())

	optlyEvents := []OptimizelyEvent{}
	for _, event := range projConfig.GetEvents() {
		optlyEvents = append(optlyEvents, OptimizelyEvent{ID: event.ID, Key: event.Key, ExperimentIds: event.ExperimentIds})
	}
	optimizelyConfig.Events = optlyEvents
	optimizelyConfig.Revision = projConfig.GetRevision()

	featuresList := projConfig.GetFeatureList()
	featuresIDMap := map[string]entities.Feature{}
	for _, feature := range featuresList {
		featuresIDMap[feature.ID] = feature
	}

	rolloutIDMap := map[string]entities.Rollout{}
	for _, rollout := range projConfig.GetRolloutList() {
		rolloutIDMap[rollout.ID] = rollout
	}

	mappedExperiments := getMappedExperiments(projConfig.GetAudienceMap(), projConfig.GetExperimentList(), projConfig.GetFeatureList(), featuresIDMap, rolloutIDMap)
	optimizelyConfig.ExperimentsMap = getExperimentsKeyMap(mappedExperiments)

	variableByIDMap := getVariableByIDMap(featuresList)
	optimizelyConfig.FeaturesMap = getFeaturesMap(projConfig.GetAudienceMap(), mappedExperiments, featuresList, rolloutIDMap, variableByIDMap)

	optimizelyConfig.datafile = projConfig.GetDatafile()

	return optimizelyConfig
}
