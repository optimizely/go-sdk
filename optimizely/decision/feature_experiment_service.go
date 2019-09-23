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
)

// FeatureExperimentService helps evaluate feature test associated with the feature
type FeatureExperimentService struct {
	compositeExperimentService ExperimentService
}

// NewFeatureExperimentService returns a new instance of the FeatureExperimentService
func NewFeatureExperimentService() *FeatureExperimentService {
	return &FeatureExperimentService{
		compositeExperimentService: NewCompositeExperimentService(),
	}
}

// GetDecision returns a decision for the given feature test and user context
func (f FeatureExperimentService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	feature := decisionContext.Feature
	// @TODO this can be improved by getting group ID first and determining experiment and then bucketing in experiment
	for _, experiment := range feature.FeatureExperiments {
		featureExperiment := experiment
		experimentDecisionContext := ExperimentDecisionContext{
			Experiment:    &featureExperiment,
			ProjectConfig: decisionContext.ProjectConfig,
		}

		experimentDecision, err := f.compositeExperimentService.GetDecision(experimentDecisionContext, userContext)
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

	return FeatureDecision{}, nil
}
