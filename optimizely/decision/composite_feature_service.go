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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

var cfLogger = logging.GetLogger("CompositeFeatureService")

// CompositeFeatureService is the default out-of-the-box feature decision service
type CompositeFeatureService struct {
	featureExperimentDecisionService ExperimentService
	rolloutDecisionService    FeatureService
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService() *CompositeFeatureService {
	return &CompositeFeatureService{
		featureExperimentDecisionService: NewFeatureExperimentService(),
		rolloutDecisionService:    NewRolloutService(),
	}
}

// GetDecision returns a decision for the given feature and user context
func (f CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	feature := decisionContext.Feature

	// Check if user is bucketed in feature experiment
	if f.featureExperimentDecisionService != nil && len(feature.FeatureExperiments) > 0 {
		// @TODO: add in a feature decision service that takes into account multiple experiments (via group mutex)
		experiment := feature.FeatureExperiments[0]
		experimentDecisionContext := ExperimentDecisionContext{
			Experiment:    &experiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}

		experimentDecision, err := f.featureExperimentDecisionService.GetDecision(experimentDecisionContext, userContext)
		// Variation not nil means we got a decision and should return it
		if experimentDecision.Variation != nil {
			featureDecision := FeatureDecision{
				Experiment: experiment,
				Decision:   experimentDecision.Decision,
				Variation:  experimentDecision.Variation,
				Source:     FeatureTest,
			}

			cfLogger.Debug(fmt.Sprintf(
				`Decision made for feature test with key "%s" for user "%s" with the following reason: "%s".`,
				feature.Key,
				userContext.ID,
				featureDecision.Reason,
			))
			return featureDecision, err
		}
	}

	featureDecisionContext := FeatureDecisionContext{
		Feature:       feature,
		ProjectConfig: decisionContext.ProjectConfig,
	}
	featureDecision, err := f.rolloutDecisionService.GetDecision(featureDecisionContext, userContext)
	featureDecision.Source = Rollout
	cfLogger.Debug(fmt.Sprintf(
		`Decision made for feature rollout with key "%s" for user "%s" with the following reason: "%s".`,
		feature.Key,
		userContext.ID,
		featureDecision.Reason,
	))

	return featureDecision, err
}
