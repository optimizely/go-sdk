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

package entities

// Audience represents an Audience object from the Optimizely datafile
type Audience struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"condition"`
}

// Experiment represents an Experiment object from the Optimizely datafile
type Experiment struct {
	// @TODO(mng): include audienceConditions
	ID                string              `json:"id"`
	Key               string              `json:"key"`
	LayerID           string              `json:"layerId"`
	Status            string              `json:"status"`
	Variations        []Variation         `json:"variations"`
	TrafficAllocation []trafficAllocation `json:"trafficAllocation"`
	AudienceIds       []string            `json:"audienceIds"`
	ForcedVariations  map[string]string   `json:"forcedVariations"`
}

// FeatureFlag represents a FeatureFlag object from the Optimizely datafile
type FeatureFlag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

// trafficAllocation represents a traffic allocation range from the Optimizely datafile
type trafficAllocation struct {
	EntityID   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

// Variation represents an experiment variation from the Optimizely datafile
type Variation struct {
	ID string `json:"id"`
	// @TODO(mng): include variables
	Key            string `json:"key"`
	FeatureEnabled bool   `json:"featureEnabled"`
}

// Datafile represents the datafile we get from Optimizely
type Datafile struct {
	AccountID    string        `json:"accountId"`
	AnonymizeIP  bool          `json:"anonymizeIP"`
	Audiences    []Audience    `json:"audiences"`
	BotFiltering bool          `json:"botFiltering"`
	Experiments  []Experiment  `json:"experiments"`
	FeatureFlags []FeatureFlag `json:"featureFlags"`
	ProjectID    string        `json:"projectId"`
	Revision     string        `json:"revision"`
	Variables    []string      `json:"variables"`
	Version      string        `json:"version"`
}
