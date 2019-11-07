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

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
)

var csLogger = logging.GetLogger("CompositeDecisionService")

// CompositeService is the entrypoint into the decision service. It provides out of the box decision making for Features and Experiments.
type CompositeService struct {
	compositeExperimentService ExperimentService
	compositeFeatureService    FeatureService
	notificationCenter         notification.Center
}

// NewCompositeService returns a new instance of the CompositeService with the defaults
func NewCompositeService(sdkKey string) *CompositeService {
	// @TODO: add factory method with option funcs to accept custom feature and experiment services
	compositeExperimentService := NewCompositeExperimentService()
	compositeFeatureDecisionService := NewCompositeFeatureService(compositeExperimentService)
	return &CompositeService{
		compositeExperimentService: compositeExperimentService,
		compositeFeatureService:    compositeFeatureDecisionService,
		notificationCenter:         registry.GetNotificationCenter(sdkKey),
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (s CompositeService) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	featureDecision, err := s.compositeFeatureService.GetDecision(featureDecisionContext, userContext)

	// @TODO: add errors
	if s.notificationCenter != nil {
		sourceInfo := map[string]string{}

		if featureDecision.Source == FeatureTest {
			sourceInfo["experimentKey"] = featureDecisionContext.Feature.Key
			sourceInfo["variationKey"] = featureDecision.Variation.Key
		}

		featureInfo := map[string]interface{}{
			"featureKey":     featureDecisionContext.Feature.Key,
			"featureEnabled": false,
			//"variableKey":    featureDecisionContext.Feature.variableKey,
			//"variableValue":  typeCastedValue,
			//"variableType":   variableType,
			"source":     featureDecision.Source,
			"sourceInfo": sourceInfo,
		}
		if featureDecision.Variation != nil {
			featureInfo["featureEnabled"] = featureDecision.Variation.FeatureEnabled
		}

		decisionInfo := map[string]interface{}{
			"feature": featureInfo,
		}

		decisionNotification := notification.DecisionNotification{
			DecisionInfo: decisionInfo,
			Type:         notification.Feature,
			UserContext:  userContext,
		}
		if err = s.notificationCenter.Send(notification.Decision, decisionNotification); err != nil {
			csLogger.Warning("Problem with sending notification")
		}
	}
	return featureDecision, err
}

// GetExperimentDecision returns a decision for the given experiment key
func (s CompositeService) GetExperimentDecision(experimentDecisionContext ExperimentDecisionContext, userContext entities.UserContext) (experimentDecision ExperimentDecision, err error) {
	if experimentDecision, err = s.compositeExperimentService.GetDecision(experimentDecisionContext, userContext); err != nil {
		return experimentDecision, err
	}

	if s.notificationCenter != nil && experimentDecision.Variation != nil {
		decisionInfo := map[string]interface{}{
			"experimentKey": experimentDecisionContext.Experiment.Key,
			"variationKey":  experimentDecision.Variation.Key,
		}

		decisionNotification := notification.DecisionNotification{
			DecisionInfo: decisionInfo,
			Type:         notification.ABTest,
			UserContext:  userContext,
		}

		if err = s.notificationCenter.Send(notification.Decision, decisionNotification); err != nil {
			csLogger.Warning("Error sending sending notification")
		}
	}

	return experimentDecision, err
}

// OnDecision registers a handler for Decision notifications
func (s CompositeService) OnDecision(callback func(notification.DecisionNotification)) (int, error) {
	handler := func(payload interface{}) {
		if decisionNotification, ok := payload.(notification.DecisionNotification); ok {
			callback(decisionNotification)
		} else {
			csLogger.Warning(fmt.Sprintf("Unable to convert notification payload %v into DecisionNotification", payload))
		}
	}
	id, err := s.notificationCenter.AddHandler(notification.Decision, handler)
	if err != nil {
		csLogger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnDecision removes handler for Decision notification with given id
func (s CompositeService) RemoveOnDecision(id int) error {
	if err := s.notificationCenter.RemoveHandler(id, notification.Decision); err != nil {
		csLogger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}
