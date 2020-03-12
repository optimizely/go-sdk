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
	"context"
	"fmt"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// CompositeFeatureService is the default out-of-the-box feature decision service
type CompositeFeatureService struct {
	featureServices []FeatureService
	context context.Context
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService(ctx context.Context, compositeExperimentService ExperimentService) *CompositeFeatureService {
	return &CompositeFeatureService{
		featureServices: []FeatureService{
			NewFeatureExperimentService(logging.GetLogger(ctx, "FeatureExperimentService" ), compositeExperimentService),
			NewRolloutService(ctx),
		},
		context:ctx,
	}
}

// GetDecision returns a decision for the given feature and user context
func (f CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	var featureDecision = FeatureDecision{}
	var err error
	for _, featureDecisionService := range f.featureServices {
		featureDecision, err = featureDecisionService.GetDecision(decisionContext, userContext)
		if err != nil {
			logging.GetLogger(f.context, "CompositeFeatureService").Debug(fmt.Sprintf("%v", err))
		}

		if featureDecision.Variation != nil && err == nil {
			return featureDecision, err
		}
	}
	return featureDecision, err
}
