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
	"github.com/optimizely/go-sdk/v2/pkg/decision/bucketer"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator"
	pkgReasons "github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// ExperimentBucketerService makes a decision using the experiment bucketer
type ExperimentBucketerService struct {
	audienceTreeEvaluator evaluator.TreeEvaluator
	bucketer              bucketer.ExperimentBucketer
	logger                logging.OptimizelyLogProducer
}

// NewExperimentBucketerService returns a new instance of the ExperimentBucketerService
func NewExperimentBucketerService(logger logging.OptimizelyLogProducer) *ExperimentBucketerService {
	// @TODO(mng): add experiment override service
	return &ExperimentBucketerService{
		logger:                logger,
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(logger),
		bucketer:              *bucketer.NewMurmurhashExperimentBucketer(logger, bucketer.DefaultHashSeed),
	}
}

// GetDecision returns the decision with the variation the user is bucketed into
func (s ExperimentBucketerService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext, options *decide.Options) (ExperimentDecision, decide.DecisionReasons, error) {
	experimentDecision := ExperimentDecision{}
	experiment := decisionContext.Experiment
	reasons := decide.NewDecisionReasons(options)

	// Determine if user can be part of the experiment
	if experiment.AudienceConditionTree != nil {
		condTreeParams := entities.NewTreeParameters(&userContext, decisionContext.ProjectConfig.GetAudienceMap())
		s.logger.Debug(fmt.Sprintf(logging.EvaluatingAudiencesForExperiment.String(), experiment.Key))
		evalResult, _, decisionReasons := s.audienceTreeEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams, options)
		reasons.Append(decisionReasons)
		logMessage := reasons.AddInfo(logging.ExperimentAudiencesEvaluatedTo.String(), experiment.Key, evalResult)
		s.logger.Debug(logMessage)
		if !evalResult {
			logMessage := reasons.AddInfo(logging.UserNotInExperiment.String(), userContext.ID, experiment.Key)
			s.logger.Debug(logMessage)
			experimentDecision.Reason = pkgReasons.FailedAudienceTargeting
			return experimentDecision, reasons, nil
		}
	} else {
		logMessage := reasons.AddInfo(logging.ExperimentAudiencesEvaluatedTo.String(), experiment.Key, true)
		s.logger.Debug(logMessage)
	}

	var group entities.Group
	if experiment.GroupID != "" {
		// @TODO: figure out what to do if group is not found
		group, _ = decisionContext.ProjectConfig.GetGroupByID(experiment.GroupID)
	}
	// bucket user into a variation
	bucketingID, err := userContext.GetBucketingID()
	if err != nil {
		errorMessage := reasons.AddInfo(`Error computing bucketing ID for experiment %q: %q`, experiment.Key, err.Error())
		s.logger.Debug(errorMessage)
	}

	if bucketingID != userContext.ID {
		s.logger.Debug(fmt.Sprintf(`Using bucketing ID: %q for user %q`, bucketingID, userContext.ID))
	}
	// @TODO: handle error from bucketer
	variation, reason, _ := s.bucketer.Bucket(bucketingID, *experiment, group)
	experimentDecision.Reason = reason
	experimentDecision.Variation = variation
	return experimentDecision, reasons, nil
}
