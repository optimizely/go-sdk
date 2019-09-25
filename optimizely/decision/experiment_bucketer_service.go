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

	"github.com/optimizely/go-sdk/optimizely/decision/bucketer"
	"github.com/optimizely/go-sdk/optimizely/decision/evaluator"
	"github.com/optimizely/go-sdk/optimizely/decision/reasons"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

// These variables are package-scoped, meaning that they can be accessed within the same package so we need unique names.
var bLogger = logging.GetLogger("ExperimentBucketerService")

// ExperimentBucketerService makes a decision using the experiment bucketer
type ExperimentBucketerService struct {
	audienceTreeEvaluator evaluator.TreeEvaluator
	bucketer              bucketer.ExperimentBucketer
}

// NewExperimentBucketerService returns a new instance of the ExperimentBucketerService
func NewExperimentBucketerService() *ExperimentBucketerService {
	// @TODO(mng): add experiment override service
	return &ExperimentBucketerService{
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(),
		bucketer:              *bucketer.NewMurmurhashBucketer(bucketer.DefaultHashSeed),
	}
}

// GetDecision returns the decision with the variation the user is bucketed into
func (s ExperimentBucketerService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	experimentDecision := ExperimentDecision{}
	experiment := decisionContext.Experiment

	// Determine if user can be part of the experiment
	if experiment.AudienceConditionTree != nil {
		condTreeParams := entities.NewTreeParameters(&userContext, decisionContext.ProjectConfig.GetAudienceMap())
		evalResult := s.audienceTreeEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams)
		if !evalResult {
			experimentDecision.Reason = reasons.FailedAudienceTargeting
			return experimentDecision, nil
		}
	}

	var group entities.Group
	if experiment.GroupID != "" {
		// @TODO: figure out what to do if group is not found
		group, _ = decisionContext.ProjectConfig.GetGroupByID(experiment.GroupID)
	}
	// bucket user into a variation
	bucketingID, err := userContext.GetBucketingID()
	if err != nil {
		bLogger.Debug(fmt.Sprintf(`Error computing bucketing ID for experiment "%s": "%s"`, experiment.Key, err.Error()))
	}

	bLogger.Debug(fmt.Sprintf(`Using bucketing ID: "%s"`, bucketingID))
	// @TODO: handle error from bucketer
	variation, reason, _ := s.bucketer.Bucket(bucketingID, *experiment, group)
	experimentDecision.Reason = reason
	experimentDecision.Variation = variation
	return experimentDecision, nil
}
