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

// Package entities //
package entities

// Variation represents a variation in the experiment
type Variation struct {
	ID             string
	Variables      map[string]VariationVariable
	Key            string
	FeatureEnabled bool
}

// Experiment represents an experiment
type Experiment struct {
	AudienceIds           []string
	ID                    string
	LayerID               string
	Key                   string
	Variations            map[string]Variation // keyed by variation ID
	VariationKeyToIDMap   map[string]string
	TrafficAllocation     []Range
	GroupID               string
	AudienceConditionTree *TreeNode
	Whitelist             map[string]string
	IsFeatureExperiment   bool
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
