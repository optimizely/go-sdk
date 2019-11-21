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

// Package decision //
package decision

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// Service interface is used to make a decision for a given feature or experiment
type Service interface {
	GetFeatureDecision(FeatureDecisionContext, entities.UserContext) (FeatureDecision, error)
	GetExperimentDecision(ExperimentDecisionContext, entities.UserContext) (ExperimentDecision, error)
	OnDecision(func(notification.DecisionNotification)) (int, error)
	RemoveOnDecision(id int) error
}

// ExperimentService can make a decision about an experiment
type ExperimentService interface {
	GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error)
}

// FeatureService can make a decision about a Feature Flag (can be feature test or rollout)
type FeatureService interface {
	GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error)
}

// UserProfileService is used to save and retrieve past bucketing decisions for users
type UserProfileService interface {
	Lookup(string) UserProfile
	Save(UserProfile)
}
