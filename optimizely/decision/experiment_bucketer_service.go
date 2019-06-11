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

// ExperimentBucketerService is the default out-of-the-box experiment decision service
type ExperimentBucketerService struct {
	overrides []ExperimentDecisionService
}

// NewExperimentBucketerService returns a new instance of the ExperimentBucketerService
func NewExperimentBucketerService() *ExperimentBucketerService {
	// @TODO(mng): add experiment override service
	return &ExperimentBucketerService{
		overrides: []ExperimentDecisionService{},
	}
}

// GetDecision returns a decision for the given experiment and user context
func (service *ExperimentBucketerService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	// @TODO(mng): use audience evaluator + bucketer to determine the variation to return
	return ExperimentDecision{
		Variation: &entities.Variation{FeatureEnabled: true},
	}, nil
}
