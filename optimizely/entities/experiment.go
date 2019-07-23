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

// Variation represents a variation in the experiment
type Variation struct {
	ID             string
	Key            string
	FeatureEnabled bool
}

// Experiment represents an experiment
type Experiment struct {
	AudienceIds           []string
	ID                    string
	LayerID               string
	Key                   string
	Variations            map[string]Variation
	TrafficAllocation     []Range
	GroupID               string
	AudienceConditionTree *TreeNode
}

// Range represents bucketing range that the specify entityID falls into
type Range struct {
	EntityID   string
	EndOfRange int
}
