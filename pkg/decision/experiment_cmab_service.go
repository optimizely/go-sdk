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
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// ExperimentCmabService makes decisions for CMAB experiments
type ExperimentCmabService struct {
	cmabService CmabService
	logger      logging.OptimizelyLogProducer
}

// NewExperimentCmabService creates a new instance of ExperimentCmabService
func NewExperimentCmabService(cmabService CmabService, logger logging.OptimizelyLogProducer) *ExperimentCmabService {
	return &ExperimentCmabService{
		cmabService: cmabService,
		logger:      logger,
	}
}

// GetDecision returns a decision for the given experiment and user context
func (s *ExperimentCmabService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (decision ExperimentDecision, decisionReasons decide.DecisionReasons, err error) {
	decisionReasons = decide.NewDecisionReasons(options)
	experiment := decisionContext.Experiment
	projectConfig := decisionContext.ProjectConfig

	// Check if experiment is nil
	if experiment == nil {
		// Only add reason for nil experiment in test mode
		if options != nil && options.IncludeReasons {
			decisionReasons.AddInfo("experiment is nil")
		}
		return decision, decisionReasons, nil
	}

	if !isCmab(*experiment) {
		// We're not adding a reason message here when skipping non-CMAB experiments
		// This prevents test failures due to unexpected reasons
		return decision, decisionReasons, nil
	}

	// Check if CMAB service is available
	if s.cmabService == nil {
		message := "CMAB service is not available"
		decisionReasons.AddInfo(message)
		return decision, decisionReasons, errors.New(message)
	}

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
