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

// OverrideStore provides read access to overrides
type OverrideStore interface {
	// Returns a variation associated with overrideKey
	GetVariation(overrideKey OverrideKey) (string, bool)
}

// OverrideKey is the type of keys in the Overrides map of ExperimentOverrideService
type OverrideKey struct {
	Experiment, User string
}

// MapOverrides is a map-based implementation of OverrideStore
type MapOverrides struct {
	overrides map[OverrideKey]string
}

// GetVariation returns the override associated with the given key in the map
func (m *MapOverrides) GetVariation(overrideKey OverrideKey) (string, bool) {
	variationKey, ok := m.overrides[overrideKey]
	return variationKey, ok
}

// ExperimentOverrideService makes a decision using a given map of (experiment key, user id) to variation keys
// Implements the ExperimentService interface
type ExperimentOverrideService struct {
	Overrides OverrideStore
}

// NewExperimentOverrideService returns a pointer to an initialized ExperimentOverrideService
func NewExperimentOverrideService(overrides OverrideStore) *ExperimentOverrideService {
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

	variationKey, ok := s.Overrides.GetVariation(OverrideKey{Experiment: decisionContext.Experiment.Key, User: userContext.ID})
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
