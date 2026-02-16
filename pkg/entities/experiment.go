/****************************************************************************
 * Copyright 2019,2021-2025 Optimizely, Inc. and contributors                   *
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

// Package entities //
package entities

// Variation represents a variation in the experiment
type Variation struct {
	ID             string
	Variables      map[string]VariationVariable
	Key            string
	FeatureEnabled bool
}

// Cmab represents the Contextual Multi-Armed Bandit configuration for an experiment
type Cmab struct {
	AttributeIds      []string `json:"attributes"`
	TrafficAllocation int      `json:"trafficAllocation"`
}

// SupportedExperimentTypes defines the set of experiment types recognized by the SDK.
// Experiments with a type not in this list will be skipped during flag decisions.
// If an experiment has no type (empty string), it is still evaluated.
var SupportedExperimentTypes = map[string]bool{
	"a/b":              true,
	"mab":              true,
	"cmab":             true,
	"feature_rollouts": true,
}

// Experiment represents an experiment
type Experiment struct {
	AudienceIds           []string
	AudienceConditions    interface{}
	ID                    string
	LayerID               string
	Key                   string
	Type                  string // experiment type (e.g., "a/b", "mab", "cmab", "feature_rollouts")
	Variations            map[string]Variation // keyed by variation ID
	VariationKeyToIDMap   map[string]string
	TrafficAllocation     []Range
	GroupID               string
	AudienceConditionTree *TreeNode
	Whitelist             map[string]string
	IsFeatureExperiment   bool
	Cmab                  *Cmab
}

// Range represents bucketing range that the specify entityID falls into
type Range struct {
	EntityID   string
	EndOfRange int
}

// VariationVariable represents a Variable object from the Variation
type VariationVariable struct {
	ID    string
	Value string
}

// HoldoutStatus represents the status of a holdout
type HoldoutStatus string

const (
	// HoldoutStatusRunning - the holdout status is running
	HoldoutStatusRunning HoldoutStatus = "Running"
)

// Holdout represents a holdout that can be applied to feature flags
type Holdout struct {
	ID                    string
	Key                   string
	Status                HoldoutStatus
	AudienceIds           []string
	AudienceConditions    interface{}
	Variations            map[string]Variation // keyed by variation ID
	TrafficAllocation     []Range
	AudienceConditionTree *TreeNode
}
