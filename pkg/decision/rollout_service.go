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
	"strconv"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/decision/evaluator"
	pkgReasons "github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
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

	getFeatureDecision := func(experiment *entities.Experiment, decision *ExperimentDecision) FeatureDecision {
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
		r.logger.Debug(fmt.Sprintf(`Decision made for user "%s" for feature rollout with key "%s": %s.`, userContext.ID, feature.Key, featureDecision.Reason))
		return featureDecision
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

	for index := 0; index < numberOfExperiments-1; index++ {
		loggingKey := strconv.Itoa(index + 1)
		experiment := &rollout.Experiments[index]

		// Checking for forced decision
		if decisionContext.ForcedDecisionService != nil {
			forcedDecision, _reasons, err := decisionContext.ForcedDecisionService.FindValidatedForcedDecision(decisionContext.ProjectConfig, decisionContext.Feature.Key, experiment.Key, options)
			reasons.Append(_reasons)
			if err == nil {
				return getFeatureDecision(experiment, &ExperimentDecision{
					Variation: forcedDecision,
				}), reasons, nil
			}
		}

		experimentDecisionContext := getExperimentDecisionContext(experiment)
		// Move to next evaluation if condition tree is available and evaluation fails

		evaluationResult := experiment.AudienceConditionTree == nil || evaluateConditionTree(experiment, loggingKey)
		r.logger.Debug(fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), loggingKey, evaluationResult))
		if !evaluationResult {
			logMessage := reasons.AddInfo(logging.UserNotInRollout.String(), userContext.ID, loggingKey)
			r.logger.Debug(logMessage)
			// Evaluate this user for the next rule
			continue
		}

		decision, decisionReasons, _ := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if decision.Variation == nil {
			// Evaluate fall back rule / last rule now
			break
		}
		finalFeatureDecision := getFeatureDecision(experiment, &decision)
		return finalFeatureDecision, reasons, nil
	}

	// fall back rule / last rule
	experiment := &rollout.Experiments[numberOfExperiments-1]
	experimentDecisionContext := getExperimentDecisionContext(experiment)
	// Move to bucketing if conditionTree is unavailable or evaluation passes
	evaluationResult := experiment.AudienceConditionTree == nil || evaluateConditionTree(experiment, "Everyone Else")
	r.logger.Debug(fmt.Sprintf(logging.RolloutAudiencesEvaluatedTo.String(), "Everyone Else", evaluationResult))

	if evaluationResult {
		decision, decisionReasons, err := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if err == nil {
			logMessage := reasons.AddInfo(logging.UserInEveryoneElse.String(), userContext.ID)
			r.logger.Debug(logMessage)
		}
		finalFeatureDecision := getFeatureDecision(experiment, &decision)
		return finalFeatureDecision, reasons, nil
	}

	return featureDecision, reasons, nil
}
