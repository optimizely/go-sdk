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

	"github.com/optimizely/go-sdk/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
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
	return &RolloutService{
		logger:                    logging.GetLogger(sdkKey, "RolloutService"),
		audienceTreeEvaluator:     evaluator.NewMixedTreeEvaluator(),
		experimentBucketerService: NewExperimentBucketerService(logging.GetLogger(sdkKey, "ExperimentBucketerService")),
	}
}

// GetDecision returns a decision for the given feature and user context
func (r RolloutService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {

	featureDecision := FeatureDecision{
		Source: Rollout,
	}
	feature := decisionContext.Feature
	rollout := feature.Rollout

	evaluateConditionTree := func(experiment *entities.Experiment) bool {
		condTreeParams := entities.NewTreeParameters(&userContext, decisionContext.ProjectConfig.GetAudienceMap())
		evalResult, _ := r.audienceTreeEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams)
		if !evalResult {
			featureDecision.Reason = reasons.FailedRolloutTargeting
			r.logger.Debug(fmt.Sprintf(`User "%s" failed targeting for feature rollout with key "%s".`, userContext.ID, feature.Key))
		}
		return evalResult
	}

	getFeatureDecision := func(experiment *entities.Experiment, decision *ExperimentDecision) (FeatureDecision, error) {
		// translate the experiment reason into a more rollouts-appropriate reason
		switch decision.Reason {
		case reasons.NotBucketedIntoVariation:
			featureDecision.Decision = Decision{Reason: reasons.FailedRolloutBucketing}
		case reasons.BucketedIntoVariation:
			featureDecision.Decision = Decision{Reason: reasons.BucketedIntoRollout}
		default:
			featureDecision.Decision = decision.Decision
		}

		featureDecision.Experiment = *experiment
		featureDecision.Variation = decision.Variation
		r.logger.Debug(fmt.Sprintf(`Decision made for user "%s" for feature rollout with key "%s": %s.`, userContext.ID, feature.Key, featureDecision.Reason))
		return featureDecision, nil
	}

	getExperimentDecisionContext := func(experiment *entities.Experiment) ExperimentDecisionContext {
		return ExperimentDecisionContext{
			Experiment:    experiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}
	}

	if rollout.ID == "" {
		featureDecision.Reason = reasons.NoRolloutForFeature
		return featureDecision, nil
	}

	numberOfExperiments := len(rollout.Experiments)
	if numberOfExperiments == 0 {
		featureDecision.Reason = reasons.RolloutHasNoExperiments
		return featureDecision, nil
	}

	for index := 0; index < numberOfExperiments-1; index++ {
		experiment := &rollout.Experiments[index]
		experimentDecisionContext := getExperimentDecisionContext(experiment)
		// Move to next evaluation if condition tree is available and evaluation fails
		if experiment.AudienceConditionTree != nil && !evaluateConditionTree(experiment) {
			// Evaluate this user for the next rule
			continue
		}
		decision, _ := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext)
		if decision.Variation == nil {
			// Evaluate fall back rule / last rule now
			break
		}
		return getFeatureDecision(experiment, &decision)
	}

	// fall back rule / last rule
	experiment := &rollout.Experiments[numberOfExperiments-1]
	experimentDecisionContext := getExperimentDecisionContext(experiment)
	// Move to bucketing if conditionTree is unavailable or evaluation passes
	if experiment.AudienceConditionTree == nil || evaluateConditionTree(experiment) {
		decision, _ := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext)
		return getFeatureDecision(experiment, &decision)
	}

	return featureDecision, nil
}
