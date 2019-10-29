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

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OverrideKey is the type of keys in the Overrides map of ExperimentOverrideService
type OverrideKey struct {
	experiment, user string
}

// ExperimentOverrideService makes a decision using a given map of (experiment key, user id) to variation keys
// Implements the ExperimentService interface
type ExperimentOverrideService struct {
	Overrides map[OverrideKey]string
}

// NewExperimentOverrideService returns a pointer to an initialized ExperimentOverrideService
func NewExperimentOverrideService(overrides map[OverrideKey]string) *ExperimentOverrideService {
	return &ExperimentOverrideService{
		Overrides: overrides,
	}
}

// GetDecision returns a decision with a variation when a variation assignment is found in the configured overrides for the given user and experiment
func (s ExperimentOverrideService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	decision := ExperimentDecision{}

	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	variationKey, ok := s.Overrides[OverrideKey{experiment: decisionContext.Experiment.Key, user: userContext.ID}]
	if !ok {
		decision.Reason = reasons.NoOverrideVariationForUser
		return decision, nil
	}

	variation, ok := decisionContext.Experiment.Variations[variationKey]
	if !ok {
		decision.Reason = reasons.InvalidOverrideVariationForUser
		return decision, nil
	}

	decision.Variation = &variation
	decision.Reason = reasons.OverrideVariationFound
	return decision, nil
}
