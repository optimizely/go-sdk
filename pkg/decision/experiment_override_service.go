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
	"strings"
	"sync"

	"github.com/optimizely/go-sdk/pkg"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

var eosLogger = logging.GetLogger("ExperimentOverrideService")

// ExperimentOverrideKey represents the user ID and experiment associated with an override variation
type ExperimentOverrideKey struct {
	ExperimentKey, UserID string
}

// ExperimentOverrideStore provides read access to overrides
type ExperimentOverrideStore interface {
	// Returns a variation associated with overrideKey
	GetVariation(overrideKey ExperimentOverrideKey) (string, bool)
}

// MEOptionFunc is used to pass custom config options into the MapExperimentOverridesStore.
type MEOptionFunc func(*MapExperimentOverridesStore)

// MapExperimentOverridesStore is a map-based implementation of ExperimentOverridesStore that is safe to use concurrently
type MapExperimentOverridesStore struct {
	overridesMap map[ExperimentOverrideKey]string
	config       pkg.ProjectConfig
	mutex        sync.RWMutex
}

// WithConfig sets the config on the MapExperimentOverridesStore
func WithConfig(config pkg.ProjectConfig) MEOptionFunc {
	return func(f *MapExperimentOverridesStore) {
		f.config = config
	}
}

// NewMapExperimentOverridesStore returns a new MapExperimentOverridesStore
func NewMapExperimentOverridesStore(options ...MEOptionFunc) *MapExperimentOverridesStore {
	overrideStore := &MapExperimentOverridesStore{
		overridesMap: make(map[ExperimentOverrideKey]string),
	}

	for _, opts := range options {
		opts(overrideStore)
	}
	return overrideStore
}

// GetVariation returns the override variation key associated with the given user+experiment key
func (m *MapExperimentOverridesStore) GetVariation(overrideKey ExperimentOverrideKey) (string, bool) {
	m.mutex.RLock()
	variationKey, ok := m.overridesMap[overrideKey]
	m.mutex.RUnlock()
	return variationKey, ok
}

// SetVariation sets the given variation key as an override for the given user+experiment key
func (m *MapExperimentOverridesStore) SetVariation(overrideKey ExperimentOverrideKey, variationKey string) bool {

	assignVariationKey := func() {
		m.mutex.Lock()
		m.overridesMap[overrideKey] = variationKey
		m.mutex.Unlock()
	}
	// Assign directly if no config provided
	if m.config == nil {
		assignVariationKey()
		return true
	}
	// Check if experiment and variation exist
	if experiment, err := m.config.GetExperimentByKey(overrideKey.ExperimentKey); err == nil {
		if strings.TrimSpace(variationKey) == "" {
			return false
		}
		if _, ok := experiment.VariationsKeyMap[variationKey]; ok {
			assignVariationKey()
			return true
		}
	}
	return false
}

// RemoveVariation removes the override variation key associated with the argument user+experiment key.
// If there is no override variation key set, this method has no effect.
func (m *MapExperimentOverridesStore) RemoveVariation(overrideKey ExperimentOverrideKey) {
	m.mutex.Lock()
	delete(m.overridesMap, overrideKey)
	m.mutex.Unlock()
}

// ExperimentOverrideService makes a decision using an ExperimentOverridesStore
// Implements the ExperimentService interface
type ExperimentOverrideService struct {
	Overrides ExperimentOverrideStore
}

// NewExperimentOverrideService returns a pointer to an initialized ExperimentOverrideService
func NewExperimentOverrideService(overrides ExperimentOverrideStore) *ExperimentOverrideService {
	return &ExperimentOverrideService{
		Overrides: overrides,
	}
}

// GetDecision returns a decision with a variation when the store returns a variation assignment for the given user and experiment
func (s ExperimentOverrideService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	decision := ExperimentDecision{}

	if decisionContext.Experiment == nil {
		return decision, errors.New("decisionContext Experiment is nil")
	}

	variationKey, ok := s.Overrides.GetVariation(ExperimentOverrideKey{ExperimentKey: decisionContext.Experiment.Key, UserID: userContext.ID})
	if !ok {
		decision.Reason = reasons.NoOverrideVariationAssignment
		return decision, nil
	}

	if variation, ok := decisionContext.Experiment.VariationsKeyMap[variationKey]; ok {
		decision.Variation = &variation
		decision.Reason = reasons.OverrideVariationAssignmentFound
		eosLogger.Debug(fmt.Sprintf("Override variation %v found for user %v", variationKey, userContext.ID))
		return decision, nil
	}

	decision.Reason = reasons.InvalidOverrideVariationAssignment
	return decision, nil
}
