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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

var ceLogger = logging.GetLogger("CompositeExperimentService")

// CompositeExperimentService bridges together the various experiment decision services that ship by default with the SDK
type CompositeExperimentService struct {
	experimentServices []ExperimentService
}

// NewCompositeExperimentService creates a new instance of the CompositeExperimentService
func NewCompositeExperimentService() *CompositeExperimentService {
	// These decision services are applied in order:
	// 1. Whitelist
	// 2. Bucketing
	// @TODO(mng): Prepend forced variation
	return &CompositeExperimentService{
		experimentServices: []ExperimentService{
			NewExperimentWhitelistService(),
			NewExperimentBucketerService(),
		},
	}
}

// GetDecision returns a decision for the given experiment and user context
func (s CompositeExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {

	experimentDecision := ExperimentDecision{}

	// Run through the various decision services until we get a decision
	for _, experimentService := range s.experimentServices {
		decision, err := experimentService.GetDecision(decisionContext, userContext)
		if err != nil {
			ceLogger.Debug(fmt.Sprintf("%v", err))
		}
		if decision.Variation != nil && err == nil {
			return decision, err
		}
	}

	return experimentDecision, errors.New("no decision was made")
}
