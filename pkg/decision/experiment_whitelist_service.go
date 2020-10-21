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
	"errors"

	"github.com/optimizely/go-sdk/pkg/decide"
	pkgReasons "github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// ExperimentWhitelistService makes a decision using an experiment's whitelist (a map of user id to variation keys)
// Implements the ExperimentService interface
type ExperimentWhitelistService struct{}

// NewExperimentWhitelistService returns a new instance of ExperimentWhitelistService
func NewExperimentWhitelistService() *ExperimentWhitelistService {
	return &ExperimentWhitelistService{}
}

// GetDecision returns a decision with a variation when a variation assignment is found in the experiment whitelist for the given user and experiment
func (s ExperimentWhitelistService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options []decide.Options, reasons decide.DecisionReasons) (ExperimentDecision, error) {
	decision := ExperimentDecision{}

	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	variationKey, ok := decisionContext.Experiment.Whitelist[userContext.ID]
	if !ok {
		decision.Reason = pkgReasons.NoWhitelistVariationAssignment
		return decision, nil
	}

	if id, ok := decisionContext.Experiment.VariationKeyToIDMap[variationKey]; ok {
		if variation, ok := decisionContext.Experiment.Variations[id]; ok {
			decision.Reason = pkgReasons.WhitelistVariationAssignmentFound
			decision.Variation = &variation
			reasons.AddInfof(`User "%s" is forced in variation "%s".`, userContext.ID, variationKey)
			return decision, nil
		}
		reasons.AddInfof(`Variation \"%s\" is not in the datafile. Not activating user \"%s\".`, variationKey, userContext.ID)
	}

	decision.Reason = pkgReasons.InvalidWhitelistVariationAssignment
	return decision, nil
}
