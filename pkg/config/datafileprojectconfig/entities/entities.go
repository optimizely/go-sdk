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

// Package entities has entity definitions
package entities

import "github.com/optimizely/go-sdk/pkg/entities"

// Audience represents an Audience object from the Optimizely datafile
type Audience struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"conditions"`
}

// Attribute represents an Attribute object from the Optimizely datafile
type Attribute struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// Experiment represents an Experiment object from the Optimizely datafile
type Experiment struct {
	ID                 string              `json:"id"`
	Key                string              `json:"key"`
	LayerID            string              `json:"layerId"`
	Status             string              `json:"status"`
	Variations         []Variation         `json:"variations"`
	TrafficAllocation  []TrafficAllocation `json:"trafficAllocation"`
	AudienceIds        []string            `json:"audienceIds"`
	ForcedVariations   map[string]string   `json:"forcedVariations"`
	AudienceConditions interface{}         `json:"audienceConditions"`
}

// Group represents an Group object from the Optimizely datafile
type Group struct {
	ID                string              `json:"id"`
	Policy            string              `json:"policy"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
	Experiments       []Experiment        `json:"experiments"`
}

// FeatureFlag represents a FeatureFlag object from the Optimizely datafile
type FeatureFlag struct {
	ID            string     `json:"id"`
	RolloutID     string     `json:"rolloutId"`
	Key           string     `json:"key"`
	ExperimentIDs []string   `json:"experimentIds"`
	Variables     []Variable `json:"variables"`
}

// Variable represents a Variable object from the Optimizely datafile
type Variable struct {
	DefaultValue string                `json:"defaultValue"`
	ID           string                `json:"id"`
	Key          string                `json:"key"`
	Type         entities.VariableType `json:"type"`
	SubType      entities.VariableType `json:"subType"`
}

// TrafficAllocation represents a traffic allocation range from the Optimizely datafile
type TrafficAllocation struct {
	EntityID   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

// Variation represents an experiment variation from the Optimizely datafile
type Variation struct {
	ID             string              `json:"id"`
	Variables      []VariationVariable `json:"variables"`
	Key            string              `json:"key"`
	FeatureEnabled bool                `json:"featureEnabled"`
}

// VariationVariable represents a Variable object from the Variation
type VariationVariable struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// Event represents an event from the Optimizely datafile
type Event struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}

// Rollout represents a rollout from the Optimizely datafile
type Rollout struct {
	ID          string       `json:"id"`
	Experiments []Experiment `json:"experiments"`
}

// Datafile represents the datafile we get from Optimizely
type Datafile struct {
	Attributes        []Attribute   `json:"attributes"`
	Audiences         []Audience    `json:"audiences"`
	Experiments       []Experiment  `json:"experiments"`
	Groups            []Group       `json:"groups"`
	FeatureFlags      []FeatureFlag `json:"featureFlags"`
	Events            []Event       `json:"events"`
	Rollouts          []Rollout     `json:"rollouts"`
	TypedAudiences    []Audience    `json:"typedAudiences"`
	Variables         []string      `json:"variables"`
	AccountID         string        `json:"accountId"`
	ProjectID         string        `json:"projectId"`
	Revision          string        `json:"revision"`
	Version           string        `json:"version"`
	AnonymizeIP       bool          `json:"anonymizeIP"`
	BotFiltering      bool          `json:"botFiltering"`
	SendFlagDecisions bool          `json:"sendFlagDecisions"`
}
