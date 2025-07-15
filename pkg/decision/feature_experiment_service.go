/****************************************************************************
 * Copyright 2019-2021, Optimizely, Inc. and contributors                   *
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

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// FeatureExperimentService helps evaluate feature test associated with the feature
type FeatureExperimentService struct {
	compositeExperimentService ExperimentService
	logger                     logging.OptimizelyLogProducer
}

// NewFeatureExperimentService returns a new instance of the FeatureExperimentService
func NewFeatureExperimentService(logger logging.OptimizelyLogProducer, compositeExperimentService ExperimentService) *FeatureExperimentService {
	return &FeatureExperimentService{
		logger:                     logger,
		compositeExperimentService: compositeExperimentService,
	}
}

// GetDecision returns a decision for the given feature test and user context
func (f FeatureExperimentService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons, error) {
	feature := decisionContext.Feature
	reasons := decide.NewDecisionReasons(options)
	// @TODO this can be improved by getting group ID first and determining experiment and then bucketing in experiment
	for _, featureExperiment := range feature.FeatureExperiments {

		// Checking for forced decision
		if decisionContext.ForcedDecisionService != nil {
			forcedDecision, _reasons, err := decisionContext.ForcedDecisionService.FindValidatedForcedDecision(decisionContext.ProjectConfig, OptimizelyDecisionContext{FlagKey: feature.Key, RuleKey: featureExperiment.Key}, options)
			reasons.Append(_reasons)
			if err == nil {
				featureDecision := FeatureDecision{
					Experiment: featureExperiment,
					Variation:  forcedDecision,
					Source:     FeatureTest,
				}
				return featureDecision, reasons, nil
			}
		}

		experiment := featureExperiment
		experimentDecisionContext := ExperimentDecisionContext{
			Experiment:    &experiment,
			ProjectConfig: decisionContext.ProjectConfig,
			UserProfile:   decisionContext.UserProfile,
		}

		experimentDecision, decisionReasons, err := f.compositeExperimentService.GetDecision(experimentDecisionContext, userContext, options)
		reasons.Append(decisionReasons)
		f.logger.Debug(fmt.Sprintf(
			`Decision made for feature test with key %q for user %q with the following reason: %q.`,
			feature.Key,
			userContext.ID,
			experimentDecision.Reason,
		))

		// Handle CMAB experiment errors - they should terminate the decision process
		if err != nil && experiment.Cmab != nil {
			// For CMAB experiments, errors should prevent fallback to other experiments
			// The error is already in reasons from decisionReasons, so return nil error
			return FeatureDecision{}, reasons, nil
		}

		// Variation not nil means we got a decision and should return it
		if experimentDecision.Variation != nil {
			featureDecision := FeatureDecision{
				Experiment: experiment,
				Decision:   experimentDecision.Decision,
				Variation:  experimentDecision.Variation,
				Source:     FeatureTest,
				CmabUUID:   experimentDecision.CmabUUID,
			}

			return featureDecision, reasons, err
		}
	}

	return FeatureDecision{}, reasons, nil
}
