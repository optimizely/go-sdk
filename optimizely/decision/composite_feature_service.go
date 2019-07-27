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

package decision

import (
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// CompositeFeatureService is the default out-of-the-box feature decision service
type CompositeFeatureService struct {
	experimentDecisionService ExperimentDecisionService
	rolloutDecisionService    FeatureDecisionService
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService() *CompositeFeatureService {
	return &CompositeFeatureService{
		rolloutDecisionService: NewRolloutService(),
	}
}

// GetDecision returns a decision for the given feature and user context
func (f CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	feature := decisionContext.Feature

	// Check if user is bucketed in feature experiment
	if f.experimentDecisionService != nil {
		// @TODO: add in a feature decision service that takes into account multiple experiments (via group mutex)
		experiment := feature.FeatureExperiments[0]
		experimentDecisionContext := ExperimentDecisionContext{
			Experiment:    &experiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}

		experimentDecision, err := f.experimentDecisionService.GetDecision(experimentDecisionContext, userContext)
		featureDecision := FeatureDecision{
			Experiment: experiment,
			Decision:   experimentDecision.Decision,
			Variation:  experimentDecision.Variation,
		}
		return featureDecision, err
	}

	featureDecisionContext := FeatureDecisionContext{
		Feature:       feature,
		ProjectConfig: decisionContext.ProjectConfig,
	}
	featureDecision, err := f.rolloutDecisionService.GetDecision(featureDecisionContext, userContext)
	featureDecision.Source = Rollout
	return featureDecision, err
}
