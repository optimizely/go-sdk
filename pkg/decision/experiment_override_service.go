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

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

var eosLogger = logging.GetLogger("ExperimentOverrideService")

// ExperimentOverrideKey represents the user ID and experiment associated with an override variation
type ExperimentOverrideKey struct {
	ExperimentKey, UserID string
}

// OverrideStore provides read access to overrides
type OverrideStore interface {
	// Returns a variation associated with overrideKey
	GetVariation(overrideKey ExperimentOverrideKey) (string, bool)
}

// MapOverridesStore is a map-based implementation of OverrideStore
type mapOverridesStore struct {
	overridesMap map[ExperimentOverrideKey]string
}

// GetVariation returns the override associated with the given key in the map
func (m *mapOverridesStore) GetVariation(overrideKey ExperimentOverrideKey) (string, bool) {
	variationKey, ok := m.overridesMap[overrideKey]
	return variationKey, ok
}

// ExperimentOverrideService makes a decision using an OverridesStore
// Implements the ExperimentService interface
type experimentOverrideService struct {
	Overrides OverrideStore
}

// NewExperimentOverrideService returns a pointer to an initialized ExperimentOverrideService
func newExperimentOverrideService(overrides OverrideStore) *experimentOverrideService {
	return &experimentOverrideService{
		Overrides: overrides,
	}
}

// GetDecision returns a decision with a variation when a variation assignment is found in the configured overrides for the given user and experiment
func (s experimentOverrideService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	decision := ExperimentDecision{}

	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	variationKey, ok := s.Overrides.GetVariation(ExperimentOverrideKey{ExperimentKey: decisionContext.Experiment.Key, UserID: userContext.ID})
	if !ok {
		decision.Reason = reasons.NoOverrideVariationForUser
		return decision, nil
	}

	for _, variation := range decisionContext.Experiment.Variations {
		if variation.Key == variationKey {
			decision.Variation = &variation
			decision.Reason = reasons.OverrideVariationFound
			eosLogger.Info(fmt.Sprintf("Override variation %v found for user %v", variationKey, userContext.ID))
			return decision, nil
		}
	}

	decision.Reason = reasons.InvalidOverrideVariationForUser
	return decision, nil
}
