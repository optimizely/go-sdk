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

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// DecisionService interface is used to make a decision for a given feature or experiment
type DecisionService interface {
	GetFeatureDecision(FeatureDecisionContext, entities.UserContext) (FeatureDecision, error)
}

// ExperimentDecisionService can make a decision about an experiment
type ExperimentDecisionService interface {
	GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error)
}

// FeatureDecisionService can make a decision about a Feature Flag (can be feature test or rollout)
type FeatureDecisionService interface {
	GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error)
}
