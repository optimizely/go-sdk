/****************************************************************************
 * Copyright 2019-2026, Optimizely, Inc. and contributors                   *
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
	holdoutService  *HoldoutService
	featureServices []FeatureService
	// rolloutService is the same instance held in featureServices; it is referenced directly
	// (rather than by slice index) for the ExcludeTargetedDeliveries path, so the logic does not
	// depend on the ordering of featureServices.
	rolloutService FeatureService
	logger         logging.OptimizelyLogProducer
}

// NewCompositeFeatureService returns a new instance of the CompositeFeatureService
func NewCompositeFeatureService(sdkKey string, compositeExperimentService ExperimentService) *CompositeFeatureService {
	holdoutService := NewHoldoutService(sdkKey)
	rolloutService := NewRolloutService(sdkKey)
	return &CompositeFeatureService{
		holdoutService: holdoutService,
		logger:         logging.GetLogger(sdkKey, "CompositeFeatureService"),
		featureServices: []FeatureService{
			NewFeatureExperimentService(logging.GetLogger(sdkKey, "FeatureExperimentService"), compositeExperimentService, holdoutService),
			rolloutService,
		},
		rolloutService: rolloutService,
	}
}

// GetDecision returns a decision for the given feature and user context
func (f CompositeFeatureService) GetDecision(decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options) (FeatureDecision, decide.DecisionReasons, error) {
	reasons := decide.NewDecisionReasons(options)

	// Evaluate global holdouts first, before any rule-level services
	if f.holdoutService != nil {
		holdoutDecision, holdoutReasons, _ := f.holdoutService.GetGlobalDecision(decisionContext, userContext, options)
		reasons.Append(holdoutReasons)
		if holdoutDecision.Variation != nil {
			if holdoutDecision.Holdout != nil && holdoutDecision.Holdout.ExcludeTargetedDeliveries {
				return f.getDecisionWithExcludedTD(holdoutDecision, decisionContext, userContext, options, reasons)
			}
			return holdoutDecision, reasons, nil
		}
	}

	var featureDecision FeatureDecision
	var err error
	for _, featureDecisionService := range f.featureServices {
		var decisionReasons decide.DecisionReasons
		featureDecision, decisionReasons, err = featureDecisionService.GetDecision(decisionContext, userContext, options)
		reasons.Append(decisionReasons)
		if err != nil {
			f.logger.Debug(err.Error())
			reasons.AddError(err.Error())
			return FeatureDecision{}, reasons, err
		}

		if featureDecision.Variation != nil {
			return featureDecision, reasons, err
		}
	}
	return featureDecision, reasons, err
}

// getDecisionWithExcludedTD handles the exclude_targeted_deliveries holdout logic.
// When a holdout has ExcludeTargetedDeliveries set, AB/MAB/CMAB experiments are
// blocked (holdout returned) but targeted delivery rules are allowed through.
func (f CompositeFeatureService) getDecisionWithExcludedTD(holdoutDecision FeatureDecision, decisionContext FeatureDecisionContext, userContext entities.UserContext, options *decide.Options, reasons decide.DecisionReasons) (FeatureDecision, decide.DecisionReasons, error) {
	holdoutExp := holdoutDecision.Experiment
	holdoutVar := holdoutDecision.Variation

	reasons.AddInfo("Holdout \"%s\" has excludeTargetedDeliveries enabled, continuing to rollout evaluation.", holdoutDecision.Holdout.Key)

	// Skip experiment evaluation entirely (A/B/MAB/CMAB are blocked by holdout).
	// Evaluate rollout service only (targeted deliveries are excluded from holdout blocking).
	if f.rolloutService != nil {
		rolloutDecision, rolloutReasons, err := f.rolloutService.GetDecision(decisionContext, userContext, options)
		reasons.Append(rolloutReasons)
		if err != nil {
			f.logger.Debug(err.Error())
			reasons.AddError(err.Error())
			return FeatureDecision{}, reasons, err
		}
		if rolloutDecision.Variation != nil {
			rolloutDecision.HoldoutExperiment = &holdoutExp
			rolloutDecision.HoldoutVariation = holdoutVar
			return rolloutDecision, reasons, nil
		}
	}

	emptyDecision := FeatureDecision{
		// Match the rollout service's no-match behavior (see RolloutService.GetDecision),
		// so a served impression under sendFlagDecisions carries ruleType "rollout" rather
		// than a blank value.
		Source:            Rollout,
		HoldoutExperiment: &holdoutExp,
		HoldoutVariation:  holdoutVar,
	}
	return emptyDecision, reasons, nil
}
