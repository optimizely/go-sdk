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

package decision

import "github.com/optimizely/go-sdk/optimizely/entities"

// ExperimentDecisionContext contains the information needed to be able to make a decision for a given experiment
type ExperimentDecisionContext struct {
	AudienceMap map[string]entities.Audience
	Experiment  entities.Experiment
	Group       entities.Group
}

// FeatureDecisionContext contains the information needed to be able to make a decision for a given feature
type FeatureDecisionContext struct {
	Feature entities.Feature
	Group   entities.Group
}

// Decision contains base information about a decision
type Decision struct {
	DecisionMade bool
}

// FeatureDecision contains the decision information about a feature
type FeatureDecision struct {
	Decision
	FeatureEnabled bool
}

// ExperimentDecision contains the decision information about an experiment
type ExperimentDecision struct {
	Decision
	Variation entities.Variation
}
