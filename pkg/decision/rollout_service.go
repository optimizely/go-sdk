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
	"strconv"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	pkgReasons "github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// RolloutService makes a feature decision for a given feature rollout
type RolloutService struct {
	audienceTreeEvaluator     evaluator.TreeEvaluator
	experimentBucketerService ExperimentService
	logger                    logging.OptimizelyLogProducer
}

// NewRolloutService returns a new instance of the Rollout service
func NewRolloutService(sdkKey string) *RolloutService {
	logger := logging.GetLogger(sdkKey, "RolloutService")
	return &RolloutService{
		logger:                    logger,
		audienceTreeEvaluator:     evaluator.NewMixedTreeEvaluator(logger),
		experimentBucketerService: NewExperimentBucketerService(logging.GetLogger(sdkKey, "ExperimentBucketerService")),
	}
}

// GetDecision returns a decision for the given feature and user context
func (r RolloutService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons, error) {
	featureDecision := FeatureDecision{
		Source: Rollout,
	}
	feature := decisionContext.Feature
	rollout := feature.Rollout
	reasons := decide.NewDecisionReasons(options)

	evaluateConditionTree := func(experiment *entities.Experiment, loggingKey string) bool {
		condTreeParams := entities.NewTreeParameters(&userContext, decisionContext.ProjectConfig.GetAudienceMap())
		r.logger.Debug(fmt.Sprintf(logging.EvaluatingAudiencesForRollout.String(), loggingKey))
		evalResult, _, decisionReasons := r.audienceTreeEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams, options)
		reasons.Append(decisionReasons)
		if !evalResult {
			featureDecision.Reason = pkgReasons.FailedRolloutTargeting
		}
		return evalResult
	}

	getExperimentDecisionContext := func(experiment *entities.Experiment) ExperimentDecisionContext {
		return ExperimentDecisionContext{
			Experiment:    experiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}
	}

	if rollout.ID == "" {
		featureDecision.Reason = pkgReasons.NoRolloutForFeature
		reasons.AddInfo(`Rollout with ID "%v" is not in the datafile.`, rollout.ID)
		return featureDecision, reasons, nil
	}

	numberOfExperiments := len(rollout.Experiments)
	if numberOfExperiments == 0 {
		featureDecision.Reason = pkgReasons.RolloutHasNoExperiments
		return featureDecision, reasons, nil
	}

	checkForForcedDecision := func(exp *entities.Experiment) *FeatureDecision {
		forcedDecision, _reasons := r.getForcedDecision(decisionContext, *exp, options)
		reasons.Append(_reasons)
		if forcedDecision != nil {
			experimentDecision := &ExperimentDecision{
				Variation: forcedDecision,
				Decision:  Decision{Reason: pkgReasons.ForcedDecisionFound},
			}
			decision := r.getFeatureDecision(&featureDecision, userContext, *feature, exp, experimentDecision)
			return &decision
		}
		return nil
	}

	// Evaluate targeted delivery rules (all except last "Everyone Else" rule)
	for index := 0; index < numberOfExperiments-1; index++ {
		loggingKey := strconv.Itoa(index + 1)
		experiment := &rollout.Experiments[index]

		decision, decisionReasons, shouldReturn := r.evaluateRolloutRule(
			experiment, loggingKey, &featureDecision, feature,
			userContext, decisionContext, options,
			checkForForcedDecision, evaluateConditionTree, getExperimentDecisionContext,
		)
		reasons.Append(decisionReasons)

		if shouldReturn {
			return decision, reasons, nil
		}
	}

	// Evaluate "Everyone Else" rule (last rule)
	decision, decisionReasons, _ := r.evaluateEveryoneElseRule(
		&rollout.Experiments[numberOfExperiments-1], &featureDecision, feature,
		userContext, decisionContext, options,
		checkForForcedDecision, evaluateConditionTree, getExperimentDecisionContext,
	)
	reasons.Append(decisionReasons)
	return decision, reasons, nil
}

// creating this sub method to avoid cyco-complexity warning
func (r RolloutService) getFeatureDecision(featureDecision *FeatureDecision, userContext entities.UserContext, feature entities.Feature, experiment *entities.Experiment, decision *ExperimentDecision) FeatureDecision {
	// translate the experiment reason into a more rollouts-appropriate reason
	switch decision.Reason {
	case pkgReasons.NotBucketedIntoVariation:
		featureDecision.Decision = Decision{Reason: pkgReasons.FailedRolloutBucketing}
	case pkgReasons.BucketedIntoVariation:
		featureDecision.Decision = Decision{Reason: pkgReasons.BucketedIntoRollout}
	default:
		featureDecision.Decision = decision.Decision
	}

	featureDecision.Variation = decision.Variation
	if featureDecision.Variation != nil {
		featureDecision.Experiment = *experiment
	}
	r.logger.Debug(fmt.Sprintf(`Decision made for user %q for feature rollout with key %q: %s.`, userContext.ID, feature.Key, featureDecision.Reason))
	return *featureDecision
}

func (r RolloutService) getForcedDecision(decisionContext FeatureDecisionContext, experiment entities.Experiment, options *decide.Options) (variation *entities.Variation, reasons decide.DecisionReasons) {
	reasons = decide.NewDecisionReasons(options)
	if decisionContext.ForcedDecisionService != nil {
		forcedDecision, _reasons, err := decisionContext.ForcedDecisionService.FindValidatedForcedDecision(decisionContext.ProjectConfig, OptimizelyDecisionContext{FlagKey: decisionContext.Feature.Key, RuleKey: experiment.Key}, options)
		reasons.Append(_reasons)
		if err == nil {
			return forcedDecision, reasons
		}
	}
	return nil, reasons
}

// evaluateHoldout evaluates a single local holdout for a user and returns a decision
func (r RolloutService) evaluateHoldout(holdout *entities.Holdout, userContext entities.UserContext, decisionContext FeatureDecisionContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons) {
	reasons := decide.NewDecisionReasons(options)

	// Create holdout service for evaluation
	holdoutService := NewHoldoutService("")

	r.logger.Debug(fmt.Sprintf("Evaluating local holdout %s for delivery rule", holdout.Key))

	// Check if holdout is running
	if holdout.Status != entities.HoldoutStatusRunning {
		reason := reasons.AddInfo("Local holdout %s is not running.", holdout.Key)
		r.logger.Info(reason)
		return FeatureDecision{}, reasons
	}

	// Check audience conditions
	inAudience := holdoutService.CheckIfUserInHoldoutAudience(holdout, userContext, decisionContext.ProjectConfig, options)
	reasons.Append(inAudience.reasons)

	if !inAudience.result {
		reason := reasons.AddInfo("User %s does not meet conditions for local holdout %s.", userContext.ID, holdout.Key)
		r.logger.Info(reason)
		return FeatureDecision{}, reasons
	}

	reason := reasons.AddInfo("User %s meets conditions for local holdout %s.", userContext.ID, holdout.Key)
	r.logger.Info(reason)

	// Get bucketing ID
	bucketingID, err := userContext.GetBucketingID()
	if err != nil {
		errorMessage := reasons.AddInfo("Error computing bucketing ID for local holdout %q: %q", holdout.Key, err.Error())
		r.logger.Debug(errorMessage)
	}

	if bucketingID != userContext.ID {
		r.logger.Debug(fmt.Sprintf("Using bucketing ID: %q for user %q", bucketingID, userContext.ID))
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
		r.logger.Info(reason)

		featureDecision := FeatureDecision{
			Experiment: experimentForBucketing,
			Variation:  variation,
			Source:     Holdout,
		}
		return featureDecision, reasons
	}

	reason = reasons.AddInfo("User %s is in no local holdout variation.", userContext.ID)
	r.logger.Info(reason)
	return FeatureDecision{}, reasons
}

// evaluateRolloutRule evaluates a single rollout rule (forced decision, local holdouts, audience, bucketing)
func (r RolloutService) evaluateRolloutRule(
	experiment *entities.Experiment,
	loggingKey string,
	featureDecision *FeatureDecision,
	feature *entities.Feature,
	userContext entities.UserContext,
	decisionContext FeatureDecisionContext,
	options *decide.Options,
	checkForForcedDecision func(*entities.Experiment) *FeatureDecision,
	evaluateConditionTree func(*entities.Experiment, string) bool,
	getExperimentDecisionContext func(*entities.Experiment) ExperimentDecisionContext,
) (FeatureDecision, decide.DecisionReasons, bool) {
	reasons := decide.NewDecisionReasons(options)

	// Check for forced decision
	if forcedDecision := checkForForcedDecision(experiment); forcedDecision != nil {
		return *forcedDecision, reasons, true
	}

	// Check local holdouts targeting this delivery rule
	localHoldouts := decisionContext.ProjectConfig.GetHoldoutsForRule(experiment.ID)
	for i := range localHoldouts {
		holdout := &localHoldouts[i]
		holdoutDecision, holdoutReasons := r.evaluateHoldout(holdout, userContext, decisionContext, options)
		reasons.Append(holdoutReasons)
		if holdoutDecision.Variation != nil {
			// User is in local holdout - return immediately
			return holdoutDecision, reasons, true
		}
	}

	// Evaluate audience conditions
	evaluationResult := experiment.AudienceConditionTree == nil || evaluateConditionTree(experiment, loggingKey)
	r.logger.Debug(fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), loggingKey, evaluationResult))
	if !evaluationResult {
		logMessage := reasons.AddInfo(logging.UserNotInRollout.String(), userContext.ID, loggingKey)
		r.logger.Debug(logMessage)
		// Continue to next rule
		return FeatureDecision{}, reasons, false
	}

	// Bucket user into variation
	experimentDecisionContext := getExperimentDecisionContext(experiment)
	decision, decisionReasons, _ := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext, options)
	reasons.Append(decisionReasons)

	if decision.Variation == nil {
		// No variation, move to fall back rule
		return FeatureDecision{}, reasons, false
	}

	// User bucketed successfully
	finalFeatureDecision := r.getFeatureDecision(featureDecision, userContext, *feature, experiment, &decision)
	return finalFeatureDecision, reasons, true
}

// evaluateEveryoneElseRule evaluates the "Everyone Else" rule (last rule in rollout)
func (r RolloutService) evaluateEveryoneElseRule(
	experiment *entities.Experiment,
	featureDecision *FeatureDecision,
	feature *entities.Feature,
	userContext entities.UserContext,
	decisionContext FeatureDecisionContext,
	options *decide.Options,
	checkForForcedDecision func(*entities.Experiment) *FeatureDecision,
	evaluateConditionTree func(*entities.Experiment, string) bool,
	getExperimentDecisionContext func(*entities.Experiment) ExperimentDecisionContext,
) (FeatureDecision, decide.DecisionReasons, error) {
	reasons := decide.NewDecisionReasons(options)

	// Check for forced decision
	if forcedDecision := checkForForcedDecision(experiment); forcedDecision != nil {
		return *forcedDecision, reasons, nil
	}

	// Check local holdouts targeting the "Everyone Else" rule
	localHoldouts := decisionContext.ProjectConfig.GetHoldoutsForRule(experiment.ID)
	for i := range localHoldouts {
		holdout := &localHoldouts[i]
		holdoutDecision, holdoutReasons := r.evaluateHoldout(holdout, userContext, decisionContext, options)
		reasons.Append(holdoutReasons)
		if holdoutDecision.Variation != nil {
			// User is in local holdout - return immediately
			return holdoutDecision, reasons, nil
		}
	}

	// Evaluate audience conditions
	evaluationResult := experiment.AudienceConditionTree == nil || evaluateConditionTree(experiment, "Everyone Else")
	r.logger.Debug(fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "Everyone Else", evaluationResult))

	if evaluationResult {
		experimentDecisionContext := getExperimentDecisionContext(experiment)
		decision, decisionReasons, err := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if err == nil {
			logMessage := reasons.AddInfo(logging.UserInEveryoneElse.String(), userContext.ID)
			r.logger.Debug(logMessage)
		}
		finalFeatureDecision := r.getFeatureDecision(featureDecision, userContext, *feature, experiment, &decision)
		return finalFeatureDecision, reasons, nil
	}

	return *featureDecision, reasons, nil
}
