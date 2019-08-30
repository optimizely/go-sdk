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
	"errors"
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// CompositeExperimentService bridges together the various experiment decision services that ship by default with the SDK
type CompositeExperimentService struct {
	experimentDecisionServices []ExperimentService
}

// NewCompositeExperimentService creates a new instance of the CompositeExperimentService
func NewCompositeExperimentService() *CompositeExperimentService {
	// These decision services are applied in order:
	// 1. Targeting
	// 2. Bucketing
	// @TODO(mng): Prepend forced variation and whitelisting services
	experimentDecisionServices := []ExperimentService{
		NewExperimentTargetingService(),
		NewExperimentBucketerService(),
	}
	return &CompositeExperimentService{
		experimentDecisionServices: experimentDecisionServices,
	}
}

// GetDecision returns a decision for the given experiment and user context
func (s CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {

	if decisionContext.Experiment.Status != entities.Running {
		errorMessage := fmt.Sprintf("Experiment %s is not running", decisionContext.Experiment.Key)
		err := errors.New(errorMessage)
		return ExperimentDecision{}, err
	}

	for _, experimentService := range s.experimentDecisionServices {
		decision, err := experimentService.GetDecision(decisionContext, userContext)
		if decision.DecisionMade {
			return decision, err
		}
	}

	// zero-value for DecisionMade is false
	return ExperimentDecision{}, nil
}
