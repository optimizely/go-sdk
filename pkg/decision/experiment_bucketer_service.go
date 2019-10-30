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

	"github.com/optimizely/go-sdk/pkg/decision/bucketer"
	"github.com/optimizely/go-sdk/pkg/decision/evaluator"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// These variables are package-scoped, meaning that they can be accessed within the same package so we need unique names.
var bLogger = logging.GetLogger("ExperimentBucketerService")

// ExperimentBucketerService makes a decision using the experiment bucketer
type ExperimentBucketerService struct {
	audienceTreeEvaluator evaluator.TreeEvaluator
	bucketer              bucketer.ExperimentBucketer
	userProfileService    UserProfileService
}

// NewExperimentBucketerService returns a new instance of the ExperimentBucketerService
func NewExperimentBucketerService() *ExperimentBucketerService {
	// @TODO(mng): add experiment override service
	return &ExperimentBucketerService{
		audienceTreeEvaluator: evaluator.NewMixedTreeEvaluator(),
		bucketer:              *bucketer.NewMurmurhashBucketer(bucketer.DefaultHashSeed),
	}
}

// ExpBucketerOptionFunc is used to extend the ExperimentBucketerService with additional config options
type ExpBucketerOptionFunc func(*ExperimentBucketerService)

// WithUserProfileService sets a user profile service on the experiment bucketer service
func WithUserProfileService(userProfileService UserProfileService) ExpBucketerOptionFunc {
	return func(s *ExperimentBucketerService) {
		s.userProfileService = userProfileService
	}
}

// GetDecision returns the decision with the variation the user is bucketed into
func (s ExperimentBucketerService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	experimentDecision := ExperimentDecision{}
	experiment := decisionContext.Experiment

	// Check for previous bucketing assignments
	savedDecision, userProfile := s.getSavedDecision(decisionContext, userContext)
	if savedDecision.Variation != nil {
		return savedDecision, nil
	}

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

	// save bucketing assignment, if applicable
	s.saveDecision(userProfile, *experiment, experimentDecision)
	return experimentDecision, nil
}

func (s ExperimentBucketerService) getSavedDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, UserProfile) {
	experimentDecision := ExperimentDecision{}

	if s.userProfileService == nil {
		return experimentDecision, UserProfile{}
	}

	userProfile, err := s.userProfileService.Lookup(userContext.ID)
	if err != nil {
		errMessage := fmt.Sprintf(`Error looking up user from user profile service: %s`, err)
		bLogger.Warning(errMessage)
		return experimentDecision, UserProfile{}
	}

	// look up experiment decision from user profile
	if savedExperimentDecision, ok := userProfile.ExperimentBucketMap[decisionContext.Experiment.ID]; ok {
		variationID := savedExperimentDecision["variation_id"]
		if variation, ok := decisionContext.Experiment.Variations[variationID]; ok {
			experimentDecision.Variation = &variation
			bLogger.Debug(fmt.Sprintf(`User "%s" was previously bucketed into variation "%s" of experiment "%s".`, userContext.ID, variation.Key, decisionContext.Experiment.Key))
		} else {
			bLogger.Warning(fmt.Sprintf(`User "%s" was previously bucketed into variation with ID "%s" for experiment "%s", but no matching variation was found.`, userContext.ID, variationID, decisionContext.Experiment.Key))
		}
	}

	return experimentDecision, userProfile
}

func (s ExperimentBucketerService) saveDecision(userProfile UserProfile, experiment entities.Experiment, decision ExperimentDecision) {
	if s.userProfileService != nil {
		if savedDecision, ok := userProfile.ExperimentBucketMap[experiment.ID]; ok {
			savedDecision["variation_id"] = decision.Variation.ID
		} else {
			userProfile.ExperimentBucketMap[experiment.ID] = map[string]string{
				"variation_id": decision.Variation.ID,
			}
		}

		err := s.userProfileService.Save(userProfile)
		if err != nil {
			bLogger.Error(`Unable to save decision to user profile service`, err)
		}
	}
}
