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
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/decision/reasons"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ExperimentWhitelistService makes a decision using an experiment's whitelist (a map of user id to variation keys)
type ExperimentWhitelistService struct{}

// NewExperimentWhitelistService returns a new instance of ExperimentWhitelistService
func NewExperimentWhitelistService() *ExperimentWhitelistService {
	return &ExperimentWhitelistService{}
}

// GetDecision returns a decision with a variation when a variation assignment is found in the experiment whitelist for the given user and experiment
func (s ExperimentWhitelistService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	decision := ExperimentDecision{}

	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	experiment, err := decisionContext.ProjectConfig.GetExperimentByKey(decisionContext.Experiment.Key)
	if err != nil {
		return decision, fmt.Errorf("error looking up experiment in decision context: %v", err)
	}

	variationKey, ok := experiment.UserIDToVariationKeyMap[userContext.ID]
	if !ok {
		decision.Reason = reasons.NoWhitelistVariationAssignment
		return decision, err
	}

	variation, ok := experiment.Variations[variationKey]
	if !ok {
		decision.Reason = reasons.InvalidWhitelistVariationAssignment
		return decision, err
	}

	decision.Reason = reasons.WhitelistVariationAssignmentFound
	decision.Variation = &variation
	return decision, err

}
