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
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/notification"
)

var csLogger = logging.GetLogger("CompositeDecisionService")

// CompositeService is the entrypoint into the decision service. It provides out of the box decision making for Features and Experiments.
type CompositeService struct {
	compositeExperimentService ExperimentService
	compositeFeatureService    FeatureService
	sdkKey                     string
}

// NewCompositeService returns a new instance of the CompositeService with the defaults
func NewCompositeService(sdkKey string) *CompositeService {
	// @TODO: add factory method with option funcs to accept custom feature and experiment services
	compositeFeatureDecisionService := NewCompositeFeatureService()
	compositeExperimentService := NewCompositeExperimentService()
	return &CompositeService{
		compositeExperimentService: compositeExperimentService,
		compositeFeatureService:    compositeFeatureDecisionService,

		sdkKey: sdkKey,
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (s CompositeService) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	featureDecision, err := s.compositeFeatureService.GetDecision(featureDecisionContext, userContext)

	// @TODO: add errors
	featureInfo := map[string]interface{}{
		"feature_key":     featureDecisionContext.Feature.Key,
		"feature_enabled": false,
		"source":          featureDecision.Source,
	}
	if featureDecision.Variation != nil {
		featureInfo["feature_enabled"] = featureDecision.Variation.FeatureEnabled
	}

	decisionInfo := map[string]interface{}{
		"feature": featureInfo,
	}

	decisionNotification := notification.DecisionNotification{
		DecisionInfo: decisionInfo,
		Type:         notification.Feature,
		UserContext:  userContext,
	}
	if err = notification.Send(s.sdkKey, notification.Decision, decisionNotification); err != nil {
		csLogger.Warning("Problem with sending notification")
	}

	return featureDecision, err
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
	id, err := notification.RegisterHandler(s.sdkKey, notification.Decision, handler)
	if err != nil {
		csLogger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnDecision removes handler for Decision notification with given id
func (s CompositeService) RemoveOnDecision(id int) error {
	if err := notification.RemoveHandler(s.sdkKey, notification.Decision, id); err != nil {
		csLogger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}
