/****************************************************************************
 * Copyright 2019-2026, Optimizely, Inc. and contributors                   *
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

		// Check local holdouts targeting this specific rule
		localHoldouts := decisionContext.ProjectConfig.GetHoldoutsForRule(featureExperiment.ID)
		for i := range localHoldouts {
			holdout := &localHoldouts[i]
			holdoutDecision, holdoutReasons := f.evaluateHoldout(holdout, userContext, decisionContext, options, feature.Key)
			reasons.Append(holdoutReasons)
			if holdoutDecision.Variation != nil {
				// User is in local holdout - return immediately, skip rule evaluation
				return holdoutDecision, reasons, nil
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
			// For CMAB experiments, errors should prevent fallback to other experiments AND rollouts
			// Return the error so CompositeFeatureService can detect it
			return FeatureDecision{}, reasons, err
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

// evaluateHoldout evaluates a single holdout for a user and returns a decision
func (f FeatureExperimentService) evaluateHoldout(holdout *entities.Holdout, userContext entities.UserContext, decisionContext FeatureDecisionContext, options *decide.Options, flagKey string) (FeatureDecision, decide.DecisionReasons) {
	reasons := decide.NewDecisionReasons(options)

	// Import the holdout service evaluation logic to avoid code duplication
	// For now, create a minimal holdout service for this evaluation
	holdoutService := NewHoldoutService("")

	// We need to evaluate just this one holdout, so we'll call the holdout service's internal logic
	// by checking audience and bucketing inline
	f.logger.Debug(fmt.Sprintf("Evaluating local holdout %s for rule %s", holdout.Key, flagKey))

	// Check if holdout is running
	if holdout.Status != entities.HoldoutStatusRunning {
		reason := reasons.AddInfo("Local holdout %s is not running.", holdout.Key)
		f.logger.Info(reason)
		return FeatureDecision{}, reasons
	}

	// Check audience conditions using the holdout service's method
	inAudience := holdoutService.CheckIfUserInHoldoutAudience(holdout, userContext, decisionContext.ProjectConfig, options)
	reasons.Append(inAudience.reasons)

	if !inAudience.result {
		reason := reasons.AddInfo("User %s does not meet conditions for local holdout %s.", userContext.ID, holdout.Key)
		f.logger.Info(reason)
		return FeatureDecision{}, reasons
	}

	reason := reasons.AddInfo("User %s meets conditions for local holdout %s.", userContext.ID, holdout.Key)
	f.logger.Info(reason)

	// Get bucketing ID
	bucketingID, err := userContext.GetBucketingID()
	if err != nil {
		errorMessage := reasons.AddInfo("Error computing bucketing ID for local holdout %q: %q", holdout.Key, err.Error())
		f.logger.Debug(errorMessage)
	}

	if bucketingID != userContext.ID {
		f.logger.Debug(fmt.Sprintf("Using bucketing ID: %q for user %q", bucketingID, userContext.ID))
	}

	// Convert holdout to experiment structure for bucketing
	experimentForBucketing := entities.Experiment{
		ID:                    holdout.ID,
		Key:                   holdout.Key,
		Variations:            holdout.Variations,
		TrafficAllocation:     holdout.TrafficAllocation,
		AudienceIds:           holdout.AudienceIds,
		AudienceConditions:    holdout.AudienceConditions,
		AudienceConditionTree: holdout.AudienceConditionTree,
	}

	// Bucket user into holdout variation
	variation, _, _ := holdoutService.bucketer.Bucket(bucketingID, experimentForBucketing, entities.Group{})

	if variation != nil {
		reason = reasons.AddInfo("User %s is in variation %s of local holdout %s.", userContext.ID, variation.Key, holdout.Key)
		f.logger.Info(reason)

		featureDecision := FeatureDecision{
			Experiment: experimentForBucketing,
			Variation:  variation,
			Source:     Holdout,
		}
		return featureDecision, reasons
	}

	reason = reasons.AddInfo("User %s is in no local holdout variation.", userContext.ID)
	f.logger.Info(reason)
	return FeatureDecision{}, reasons
}
