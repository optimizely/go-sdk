/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// CESOptionFunc is used to assign optional configuration options
type CESOptionFunc func(*CompositeExperimentService)

// WithUserProfileService adds a user profile service
func WithUserProfileService(userProfileService UserProfileService) CESOptionFunc {
	return func(f *CompositeExperimentService) {
		f.userProfileService = userProfileService
	}
}

// WithOverrideStore adds an experiment override store
func WithOverrideStore(overrideStore ExperimentOverrideStore) CESOptionFunc {
	return func(f *CompositeExperimentService) {
		f.overrideStore = overrideStore
	}
}

// CompositeExperimentService bridges together the various experiment decision services that ship by default with the SDK
type CompositeExperimentService struct {
	experimentServices []ExperimentService
	overrideStore      ExperimentOverrideStore
	userProfileService UserProfileService
	logger             logging.OptimizelyLogProducer
}

// NewCompositeExperimentService creates a new composite experiment service with the given SDK key.
// It initializes a service that combines multiple decision services in a specific order:
// 1. Overrides (if supplied)
// 2. Whitelist
// 3. CMAB (Contextual Multi-Armed Bandit)
// 4. Bucketing (with User profile integration if supplied)
// Additional options can be provided via CESOptionFunc parameters.
func NewCompositeExperimentService(sdkKey string, options ...CESOptionFunc) *CompositeExperimentService {
	// These decision services are applied in order:
	// 1. Overrides (if supplied)
	// 2. Whitelist
	// 3. CMAB (always created)
	// 4. Bucketing (with User profile integration if supplied)
	compositeExperimentService := &CompositeExperimentService{logger: logging.GetLogger(sdkKey, "CompositeExperimentService")}

	for _, opt := range options {
		opt(compositeExperimentService)
	}

	experimentServices := []ExperimentService{
		NewExperimentWhitelistService(), // No logger argument
	}

	if compositeExperimentService.overrideStore != nil {
		overrideService := NewExperimentOverrideService(compositeExperimentService.overrideStore, logging.GetLogger(sdkKey, "ExperimentOverrideService"))
		experimentServices = append([]ExperimentService{overrideService}, experimentServices...)
	}

	// Create CMAB service with all initialization handled internally
	experimentCmabService := NewExperimentCmabService(sdkKey)
	experimentServices = append(experimentServices, experimentCmabService)

	experimentBucketerService := NewExperimentBucketerService(logging.GetLogger(sdkKey, "ExperimentBucketerService"))
	if compositeExperimentService.userProfileService != nil {
		persistingExperimentService := NewPersistingExperimentService(compositeExperimentService.userProfileService, experimentBucketerService, logging.GetLogger(sdkKey, "PersistingExperimentService"))
		experimentServices = append(experimentServices, persistingExperimentService)
	} else {
		experimentServices = append(experimentServices, experimentBucketerService)
	}

	compositeExperimentService.experimentServices = experimentServices
	return compositeExperimentService
}

// GetDecision attempts to get an experiment decision by trying each configured experiment service
// in order until one returns a valid decision. If a service returns an error, it continues to the next service.
// Returns the first valid decision found, accumulated decision reasons, and any error from the last failed service.
func (s *CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (ExperimentDecision, decide.DecisionReasons, error) {
	var experDecision ExperimentDecision
	var err error
	reasons := decide.NewDecisionReasons(options)

	for _, experimentService := range s.experimentServices {
		var serviceReasons decide.DecisionReasons

		decision, serviceReasons, serviceErr := experimentService.GetDecision(decisionContext, userContext, options)
		reasons.Append(serviceReasons)

		// If there's an error, log it and continue to next service
		if serviceErr != nil {
			// Optionally store the last error for potential logging
			err = serviceErr
			continue
		}

		// If we got a valid decision (has a variation), return it
		if decision.Variation != nil {
			return decision, reasons, nil
		}

		// No error and no decision - continue to next service
	}

	// No service could make a decision
	return experDecision, reasons, err // Returns last error (or nil if no errors)
}
