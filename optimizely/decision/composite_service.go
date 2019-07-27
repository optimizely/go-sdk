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

// CompositeService is the entrypoint into the decision service. It provides out of the box decision making for Features and Experiments.
type CompositeService struct {
	experimentDecisionServices []ExperimentDecisionService
	featureDecisionServices    []FeatureDecisionService
}

// NewCompositeService returns a new instance of the DefeaultDecisionEngine
func NewCompositeService() *CompositeService {
	featureDecisionService := NewCompositeFeatureService()
	return &CompositeService{
		featureDecisionServices: []FeatureDecisionService{featureDecisionService},
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (s CompositeService) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	var featureDecision FeatureDecision

	// loop through the different features decision services until we get a decision
	for _, decisionService := range s.featureDecisionServices {
		featureDecision, err := decisionService.GetDecision(featureDecisionContext, userContext)
		if err != nil {
			// @TODO: log error
		}

		if featureDecision.DecisionMade {
			return featureDecision, err
		}
	}

	// @TODO: add errors
	return featureDecision, nil
}
