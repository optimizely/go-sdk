/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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
	"fmt"
	"net/http"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/cache"
	"github.com/optimizely/go-sdk/v2/pkg/cmab"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/bucketer"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	pkgReasons "github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// ExperimentCmabService makes a decision using CMAB
type ExperimentCmabService struct {
	audienceTreeEvaluator evaluator.TreeEvaluator
	bucketer              bucketer.ExperimentBucketer
	cmabService           cmab.Service
	logger                logging.OptimizelyLogProducer
}

// cmabClientAdapter adapts the DefaultCmabClient to the CmabClient interface
type cmabClientAdapter struct {
	client *cmab.DefaultCmabClient
}

// FetchDecision implements the CmabClient interface by calling the DefaultCmabClient with a background context
func (a *cmabClientAdapter) FetchDecision(ruleID, userID string, attributes map[string]interface{}, cmabUUID string) (string, error) {
	// Use background context for the adapted call
	return a.client.FetchDecision(context.Background(), ruleID, userID, attributes, cmabUUID)
}

// NewExperimentCmabService creates a new instance of ExperimentCmabService with all dependencies initialized
func NewExperimentCmabService(sdkKey string) *ExperimentCmabService {
	// Initialize CMAB cache
	cmabCache := cache.NewLRUCache(100, 0)

	// Create retry config for CMAB client
	retryConfig := &cmab.RetryConfig{
		MaxRetries:        cmab.DefaultMaxRetries,
		InitialBackoff:    cmab.DefaultInitialBackoff,
		MaxBackoff:        cmab.DefaultMaxBackoff,
		BackoffMultiplier: cmab.DefaultBackoffMultiplier,
	}

	// Create CMAB client options
	cmabClientOptions := cmab.ClientOptions{
		HTTPClient:  &http.Client{Timeout: 10 * time.Second},
		RetryConfig: retryConfig,
		Logger:      logging.GetLogger(sdkKey, "DefaultCmabClient"),
	}

	// Create CMAB client with adapter to match interface
	defaultCmabClient := cmab.NewDefaultCmabClient(cmabClientOptions)
	cmabClientAdapter := &cmabClientAdapter{client: defaultCmabClient}

	// Create CMAB service options
	cmabServiceOptions := cmab.ServiceOptions{
		CmabCache:  cmabCache,
		CmabClient: cmabClientAdapter,
		Logger:     logging.GetLogger(sdkKey, "DefaultCmabService"),
	}

	// Create CMAB service
	cmabService := cmab.NewDefaultCmabService(cmabServiceOptions)

	// Create logger for this service
	logger := logging.GetLogger(sdkKey, "ExperimentCmabService")

	return &ExperimentCmabService{
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(logger),
		bucketer:              *bucketer.NewMurmurhashExperimentBucketer(logger, bucketer.DefaultHashSeed),
		cmabService:           cmabService,
		logger:                logger,
	}
}

// GetDecision returns a decision for the given experiment and user context
func (s *ExperimentCmabService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision ExperimentDecision, decisionReasons decide.DecisionReasons, err error) {
	decisionReasons = decide.NewDecisionReasons(options)
	experiment := decisionContext.Experiment
	projectConfig := decisionContext.ProjectConfig

	// Check if experiment is nil
	if experiment == nil {
		if options != nil && options.IncludeReasons {
			decisionReasons.AddInfo("experiment is nil")
		}
		return decision, decisionReasons, nil
	}

	if !isCmab(*experiment) {
		return decision, decisionReasons, nil
	}

	// Check if CMAB service is available
	if s.cmabService == nil {
		message := "CMAB service is not available"
		decisionReasons.AddInfo(message)
		return decision, decisionReasons, fmt.Errorf(message)
	}

	// Audience evaluation using common function
	inAudience, audienceReasons := evaluator.CheckIfUserInAudience(experiment, userContext, projectConfig, s.audienceTreeEvaluator, options, s.logger)
	decisionReasons.Append(audienceReasons)

	if !inAudience {
		logMessage := decisionReasons.AddInfo("User %s not in audience for CMAB experiment %s", userContext.ID, experiment.Key)
		s.logger.Debug(logMessage)
		decision.Reason = pkgReasons.FailedAudienceTargeting
		return decision, decisionReasons, nil
	}

	// Traffic allocation check with CMAB-specific traffic allocation
	var group entities.Group
	if experiment.GroupID != "" {
		group, _ = projectConfig.GetGroupByID(experiment.GroupID)
	}

	bucketingID, err := userContext.GetBucketingID()
	if err != nil {
		errorMessage := decisionReasons.AddInfo("Error computing bucketing ID for CMAB experiment %s: %s", experiment.Key, err.Error())
		s.logger.Debug(errorMessage)
	}

	if bucketingID != userContext.ID {
		s.logger.Debug(fmt.Sprintf("Using bucketing ID: %s for user %s in CMAB experiment", bucketingID, userContext.ID))
	}

	// Update traffic allocation for CMAB experiments (100% allocation)
	updatedExperiment := *experiment
	updatedExperiment.TrafficAllocation = []entities.Range{
		{
			EntityID:   "",    // Empty entity ID
			EndOfRange: 10000, // 100% traffic allocation
		},
	}

	// Check if user is in experiment traffic allocation
	variation, reason, _ := s.bucketer.Bucket(bucketingID, updatedExperiment, group)
	if variation == nil {
		logMessage := decisionReasons.AddInfo("User %s not in CMAB experiment %s due to traffic allocation", userContext.ID, experiment.Key)
		s.logger.Debug(logMessage)
		decision.Reason = reason
		return decision, decisionReasons, nil
	}

	// User passed audience and traffic allocation - now use CMAB service
	// Get CMAB decision
	cmabDecision, err := s.cmabService.GetDecision(projectConfig, userContext, experiment.ID, options)
	if err != nil {
		message := fmt.Sprintf("Failed to get CMAB decision: %v", err)
		decisionReasons.AddInfo(message)
		return decision, decisionReasons, fmt.Errorf("failed to get CMAB decision: %w", err)
	}

	// Find variation by ID
	for _, variation := range experiment.Variations {
		if variation.ID != cmabDecision.VariationID {
			continue
		}

		// Create a copy of the variation to avoid memory aliasing
		variationCopy := variation
		decision.Variation = &variationCopy
		decision.Reason = reasons.CmabVariationAssigned
		message := fmt.Sprintf("User bucketed into variation %s by CMAB service", variation.Key)
		decisionReasons.AddInfo(message)
		return decision, decisionReasons, nil
	}

	// If we get here, the variation ID returned by CMAB service was not found
	message := fmt.Sprintf("variation with ID %s not found in experiment %s", cmabDecision.VariationID, experiment.ID)
	decisionReasons.AddInfo(message)
	return decision, decisionReasons, fmt.Errorf("variation with ID %s not found in experiment %s", cmabDecision.VariationID, experiment.ID)
}

// isCmab is a helper method to check if an experiment is a CMAB experiment
func isCmab(experiment entities.Experiment) bool {
	return experiment.Cmab != nil
}
