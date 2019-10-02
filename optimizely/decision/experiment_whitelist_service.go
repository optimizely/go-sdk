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

// ExperimentWhitelistService makes a decision using a whitelist (a set of experiment + variation assignments for a set of users)
// whitelist should be a map of user ID, to a map of Experiment key to Variation key
type ExperimentWhitelistService struct {
	whitelist map[string]map[string]string
}

// NewExperimentWhitelistService returns a new instance of ExperimentWhitelistService
func NewExperimentWhitelistService(whitelist map[string]map[string]string) *ExperimentWhitelistService {
	return &ExperimentWhitelistService{
		whitelist: whitelist,
	}
}

// GetDecision returns a decision with a variation when an entry is found for a given user ID and experiment key
func (s ExperimentWhitelistService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (decision ExperimentDecision, err error) {
	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	experiment, err := decisionContext.ProjectConfig.GetExperimentByKey(decisionContext.Experiment.Key)
	if err != nil {
		return decision, fmt.Errorf("error looking up experiment in decision context: %v", err)
	}

	userEntry, ok := s.whitelist[userContext.ID]
	if !ok {
		decision.Reason = reasons.NoWhitelistVariationAssignment
		return decision, nil
	}

	variationKey, ok := userEntry[decisionContext.Experiment.Key]
	if !ok {
		decision.Reason = reasons.NoWhitelistVariationAssignment
		return decision, nil
	}

	variation, ok := experiment.Variations[variationKey]
	if !ok {
		decision.Reason = reasons.InvalidWhitelistVariationAssignment
		return decision, nil
	}

	decision.Reason = reasons.WhitelistVariationAssignmentFound
	decision.Variation = &variation
	return decision, nil
}
