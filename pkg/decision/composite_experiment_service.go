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

// WithCmabService adds a CMAB service
func WithCmabService(cmabService CmabService) CESOptionFunc {
	return func(f *CompositeExperimentService) {
		f.cmabService = cmabService
	}
}

// CompositeExperimentService bridges together the various experiment decision services that ship by default with the SDK
type CompositeExperimentService struct {
	experimentServices []ExperimentService
	overrideStore      ExperimentOverrideStore
	userProfileService UserProfileService
	cmabService        CmabService
	logger             logging.OptimizelyLogProducer
}

// NewCompositeExperimentService creates a new instance of the CompositeExperimentService
func NewCompositeExperimentService(sdkKey string, options ...CESOptionFunc) *CompositeExperimentService {
	// These decision services are applied in order:
	// 1. Overrides (if supplied)
	// 2. Whitelist
	// 3. CMAB (if experiment is a CMAB experiment)
	// 4. Bucketing (with User profile integration if supplied)
	compositeExperimentService := &CompositeExperimentService{logger: logging.GetLogger(sdkKey, "CompositeExperimentService")}
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

	// Add CMAB service if available
	if compositeExperimentService.cmabService != nil {
		cmabService := NewExperimentCmabService(compositeExperimentService.cmabService, logging.GetLogger(sdkKey, "ExperimentCmabService"))
		experimentServices = append(experimentServices, cmabService)
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
func (s CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision ExperimentDecision, reasons decide.DecisionReasons, err error) {
	// Run through the various decision services until we get a decision
	reasons = decide.NewDecisionReasons(options)
	for _, experimentService := range s.experimentServices {
		var decisionReasons decide.DecisionReasons
		decision, decisionReasons, err = experimentService.GetDecision(decisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if err != nil {
			s.logger.Debug(err.Error())
		}
		if decision.Variation != nil && err == nil {
			return decision, reasons, err
		}
	}

	return decision, reasons, err
}
