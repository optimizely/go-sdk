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
	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/optimizely/go-sdk/pkg/entities"
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
	if rollout.ID == "" {
		featureDecision.Reason = reasons.NoRolloutForFeature
		return featureDecision, nil
	}

	numberOfExperiments := len(rollout.Experiments)
	if numberOfExperiments == 0 {
		featureDecision.Reason = reasons.RolloutHasNoExperiments
		return featureDecision, nil
	}

	index := 0
	for index < numberOfExperiments {
		experiment := rollout.Experiments[index]
		experimentDecisionContext := ExperimentDecisionContext{
			Experiment:    &experiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}
		if experiment.AudienceConditionTree != nil {
			condTreeParams := entities.NewTreeParameters(&userContext, decisionContext.ProjectConfig.GetAudienceMap())
			evalResult, _ := r.audienceTreeEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams)
			if !evalResult { // Evaluate this user for the next rule
				featureDecision.Reason = reasons.FailedRolloutTargeting
				r.logger.Debug(fmt.Sprintf(`User "%s" failed targeting for feature rollout with key "%s".`, userContext.ID, feature.Key))
				index++
				continue
			}
		}
		decision, _ := r.experimentBucketerService.GetDecision(experimentDecisionContext, userContext)
		if decision.Variation == nil && index < numberOfExperiments-1 {
			// Evaluate fall back rule / last rule now
			index = numberOfExperiments - 1
			continue
		}
		// translate the experiment reason into a more rollouts-appropriate reason
		switch decision.Reason {
		case reasons.NotBucketedIntoVariation:
			featureDecision.Decision = Decision{Reason: reasons.FailedRolloutBucketing}
		case reasons.BucketedIntoVariation:
			featureDecision.Decision = Decision{Reason: reasons.BucketedIntoRollout}
		default:
			featureDecision.Decision = decision.Decision
		}

		featureDecision.Experiment = experiment
		featureDecision.Variation = decision.Variation
		r.logger.Debug(fmt.Sprintf(`Decision made for user "%s" for feature rollout with key "%s": %s.`, userContext.ID, feature.Key, featureDecision.Reason))
		break
	}
	return featureDecision, nil
}
