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

// Feature represents a feature flag
type Feature struct {
	ID                 string
	Key                string
	FeatureExperiments []Experiment
	Rollout            Rollout
}

// FeatureVariable represents a variable
type FeatureVariable struct {
	Key   string
	Type  string
	Value string
}

// Rollout represents a feature rollout
type Rollout struct {
	ID          string
	Experiments []Experiment
}
