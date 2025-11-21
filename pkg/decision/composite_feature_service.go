/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// CompositeFeatureService is the default out-of-the-box feature decision service
type CompositeFeatureService struct {
	featureServices []FeatureService
	logger          logging.OptimizelyLogProducer
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService(sdkKey string, compositeExperimentService ExperimentService) *CompositeFeatureService {
	return &CompositeFeatureService{
		logger: logging.GetLogger(sdkKey, "CompositeFeatureService"),
		featureServices: []FeatureService{
			NewHoldoutService(sdkKey),
			NewFeatureExperimentService(logging.GetLogger(sdkKey, "FeatureExperimentService"), compositeExperimentService),
			NewRolloutService(sdkKey),
		},
	}
}

// GetDecision returns a decision for the given feature and user context
func (f CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons, error) {
	var featureDecision = FeatureDecision{}
	reasons := decide.NewDecisionReasons(options)
	var err error
	for _, featureDecisionService := range f.featureServices {
		var decisionReasons decide.DecisionReasons
		featureDecision, decisionReasons, err = featureDecisionService.GetDecision(decisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if err != nil {
			f.logger.Debug(err.Error())
			reasons.AddError(err.Error())
			// Return the error to let the caller handle it properly
			return FeatureDecision{}, reasons, err
		}

		if featureDecision.Variation != nil {
			return featureDecision, reasons, err
		}
	}
	return featureDecision, reasons, err
}
