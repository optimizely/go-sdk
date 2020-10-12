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

// Package decision //
package decision

import (
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// ExperimentDecisionContext contains the information needed to be able to make a decision for a given experiment
type ExperimentDecisionContext struct {
	Experiment    *entities.Experiment
	ProjectConfig config.ProjectConfig
}

// FeatureDecisionContext contains the information needed to be able to make a decision for a given feature
type FeatureDecisionContext struct {
	Feature       *entities.Feature
	ProjectConfig config.ProjectConfig
	Variable      entities.Variable
}

// UnsafeFeatureDecisionInfo represents response for GetDetailedFeatureDecisionUnsafe api
type UnsafeFeatureDecisionInfo struct {
	Enabled       bool
	VariableMap   map[string]interface{}
	ExperimentKey string
	VariationKey  string
}

// Source is where the decision came from
type Source = string

const (
	// Rollout - the decision came from a rollout
	Rollout Source = "rollout"
	// FeatureTest - the decision came from a feature test
	FeatureTest Source = "feature-test"
)

// Decision contains base information about a decision
type Decision struct {
	Reason reasons.Reason
}

// FeatureDecision contains the decision information about a feature
type FeatureDecision struct {
	Decision
	Source     Source
	Experiment entities.Experiment
	Variation  *entities.Variation
}

// ExperimentDecision contains the decision information about an experiment
type ExperimentDecision struct {
	Decision
	Variation *entities.Variation
}

// UserDecisionKey is used to access the saved decisions in a user profile
type UserDecisionKey struct {
	ExperimentID string
	Field        string
}

// NewUserDecisionKey returns a new UserDecisionKey with the given experiment ID
func NewUserDecisionKey(experimentID string) UserDecisionKey {
	return UserDecisionKey{
		ExperimentID: experimentID,
		Field:        "variation_id",
	}
}

// UserProfile represents a saved user profile
type UserProfile struct {
	ID                  string
	ExperimentBucketMap map[UserDecisionKey]string
}
