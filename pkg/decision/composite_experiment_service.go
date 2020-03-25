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
	"fmt"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
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

// NewCompositeExperimentService creates a new instance of the CompositeExperimentService
func NewCompositeExperimentService(sdkKey string, options ...CESOptionFunc) *CompositeExperimentService {
	// These decision services are applied in order:
	// 1. Overrides (if supplied)
	// 2. Whitelist
	// 3. Bucketing (with User profile integration if supplied)
	compositeExperimentService := &CompositeExperimentService{logger:logging.GetLogger(sdkKey, "CompositeExperimentService")}
	for _, opt := range options {
		opt(compositeExperimentService)
	}
	experimentServices := []ExperimentService{
		NewExperimentWhitelistService(),
	}

	// Prepend overrides if supplied
	if compositeExperimentService.overrideStore != nil {
		overrideService := NewExperimentOverrideService(compositeExperimentService.overrideStore, logging.GetLogger(sdkKey, "ExperimentOverrideService"))
		experimentServices = append([]ExperimentService{overrideService}, experimentServices...)
	}

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

// GetDecision returns a decision for the given experiment and user context
func (s CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (decision ExperimentDecision, err error) {

	// Run through the various decision services until we get a decision
	for _, experimentService := range s.experimentServices {
		decision, err = experimentService.GetDecision(decisionContext, userContext)
		if err != nil {
			s.logger.Debug(fmt.Sprintf("%v", err))
		}
		if decision.Variation != nil && err == nil {
			return decision, err
		}
	}

	return decision, err
}
