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
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/bucketer"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// HoldoutService evaluates holdout groups for feature flags
type HoldoutService struct {
	audienceTreeEvaluator evaluator.TreeEvaluator
	bucketer              bucketer.ExperimentBucketer
	logger                logging.OptimizelyLogProducer
}

// NewHoldoutService returns a new instance of the HoldoutService
func NewHoldoutService(sdkKey string) *HoldoutService {
	logger := logging.GetLogger(sdkKey, "HoldoutService")
	return &HoldoutService{
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(logger),
		bucketer:              bucketer.NewMurmurhashExperimentBucketer(logger, bucketer.DefaultHashSeed),
		logger:                logger,
	}
}

// GetDecision returns a decision for holdouts associated with the feature
func (h HoldoutService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons, error) {
	feature := decisionContext.Feature
	reasons := decide.NewDecisionReasons(options)

	holdouts := decisionContext.ProjectConfig.GetHoldoutsForFlag(feature.Key)

	for i := range holdouts {
		holdout := &holdouts[i]
		h.logger.Debug(fmt.Sprintf("Evaluating holdout %s for feature %s", holdout.Key, feature.Key))

		// Check if holdout is running
		if holdout.Status != entities.HoldoutStatusRunning {
			reason := reasons.AddInfo("Holdout %s is not running.", holdout.Key)
			h.logger.Info(reason)
			continue
		}

		// Check audience conditions
		inAudience := h.checkIfUserInHoldoutAudience(holdout, userContext, decisionContext.ProjectConfig, options)
		reasons.Append(inAudience.reasons)

		if !inAudience.result {
			reason := reasons.AddInfo("User %s does not meet conditions for holdout %s.", userContext.ID, holdout.Key)
			h.logger.Info(reason)
			continue
		}

		reason := reasons.AddInfo("User %s meets conditions for holdout %s.", userContext.ID, holdout.Key)
		h.logger.Info(reason)

		// Get bucketing ID
		bucketingID, err := userContext.GetBucketingID()
		if err != nil {
			errorMessage := reasons.AddInfo("Error computing bucketing ID for holdout %q: %q", holdout.Key, err.Error())
			h.logger.Debug(errorMessage)
		}

		if bucketingID != userContext.ID {
			h.logger.Debug(fmt.Sprintf("Using bucketing ID: %q for user %q", bucketingID, userContext.ID))
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
		variation, _, _ := h.bucketer.Bucket(bucketingID, experimentForBucketing, entities.Group{})

		if variation != nil {
			reason := reasons.AddInfo("User %s is in variation %s of holdout %s.", userContext.ID, variation.Key, holdout.Key)
			h.logger.Info(reason)

			featureDecision := FeatureDecision{
				Experiment: experimentForBucketing,
				Variation:  variation,
				Source:     Holdout,
			}
			return featureDecision, reasons, nil
		}

		reason = reasons.AddInfo("User %s is in no holdout variation.", userContext.ID)
		h.logger.Info(reason)
	}

	return FeatureDecision{}, reasons, nil
}

// checkIfUserInHoldoutAudience evaluates if user meets holdout audience conditions
func (h HoldoutService) checkIfUserInHoldoutAudience(holdout *entities.Holdout, userContext entities.UserContext, projectConfig config.ProjectConfig, options *decide.Options) decisionResult {
	decisionReasons := decide.NewDecisionReasons(options)

	if holdout == nil {
		logMessage := decisionReasons.AddInfo("Holdout is nil, defaulting to false")
		h.logger.Debug(logMessage)
		return decisionResult{result: false, reasons: decisionReasons}
	}

	if holdout.AudienceConditionTree != nil {
		condTreeParams := entities.NewTreeParameters(&userContext, projectConfig.GetAudienceMap())
		h.logger.Debug(fmt.Sprintf("Evaluating audiences for holdout %q.", holdout.Key))

		evalResult, _, audienceReasons := h.audienceTreeEvaluator.Evaluate(holdout.AudienceConditionTree, condTreeParams, options)
		decisionReasons.Append(audienceReasons)

		logMessage := decisionReasons.AddInfo("Audiences for holdout %s collectively evaluated to %v.", holdout.Key, evalResult)
		h.logger.Debug(logMessage)

		return decisionResult{result: evalResult, reasons: decisionReasons}
	}

	logMessage := decisionReasons.AddInfo("Audiences for holdout %s collectively evaluated to true.", holdout.Key)
	h.logger.Debug(logMessage)
	return decisionResult{result: true, reasons: decisionReasons}
}

// decisionResult is a helper struct to return both result and reasons
type decisionResult struct {
	result  bool
	reasons decide.DecisionReasons
}
