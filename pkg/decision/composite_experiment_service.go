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
	"context"
	"net/http"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cache"

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

	// Always create CMAB service - no conditional check
	cmabCache := cache.NewLRUCache(100, 0)

	// Create retry config for CMAB client
	retryConfig := &RetryConfig{
		MaxRetries:        DefaultMaxRetries,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiplier,
	}

	// Create CMAB client options
	cmabClientOptions := CmabClientOptions{
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		RetryConfig: retryConfig,
		Logger:      logging.GetLogger(sdkKey, "DefaultCmabClient"),
	}

	// Create CMAB client with adapter to match interface
	defaultCmabClient := NewDefaultCmabClient(cmabClientOptions)
	cmabClientAdapter := &cmabClientAdapter{client: defaultCmabClient}

	// Create CMAB service options
	cmabServiceOptions := CmabServiceOptions{
		CmabCache:  cmabCache,
		CmabClient: cmabClientAdapter,
		Logger:     logging.GetLogger(sdkKey, "DefaultCmabService"),
	}

	// Create CMAB service
	cmabService := NewDefaultCmabService(cmabServiceOptions)
	experimentCmabService := NewExperimentCmabService(cmabService, logging.GetLogger(sdkKey, "ExperimentCmabService"))
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

// cmabClientAdapter adapts the DefaultCmabClient to the CmabClient interface
type cmabClientAdapter struct {
	client *DefaultCmabClient
}

// FetchDecision implements the CmabClient interface by calling the DefaultCmabClient with a background context
func (a *cmabClientAdapter) FetchDecision(ruleID, userID string, attributes map[string]interface{}, cmabUUID string) (string, error) {
	// Use background context for the adapted call
	return a.client.FetchDecision(context.Background(), ruleID, userID, attributes, cmabUUID)
}

// GetDecision returns a decision for the given experiment and user context
func (s CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision ExperimentDecision, reasons decide.DecisionReasons, err error) {
	// Run through the various decision services until we get a decision
	reasons = decide.NewDecisionReasons(options)
	for _, experimentService := range s.experimentServices {
		var decisionReasons decide.DecisionReasons
		decision, decisionReasons, err = experimentService.GetDecision(decisionContext, userContext, options)
		reasons.Append(decisionReasons)

		// If there's an error, it should only come from CMAB service
		// We immediately return it without trying other services
		if err != nil {
			s.logger.Error("Error getting decision", err)
			return decision, reasons, err
		}

		if decision.Variation != nil {
			return decision, reasons, nil
		}
	}
	return decision, reasons, nil
}
