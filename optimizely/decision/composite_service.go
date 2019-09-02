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
	// experimentDecisionServices []ExperimentDecisionService
	featureDecisionServices []FeatureService
	notificationCenter      notification.Center
}

// NewCompositeService returns a new instance of the DefeaultDecisionEngine
func NewCompositeService(notificationCenter notification.Center) *CompositeService {
	featureDecisionService := NewCompositeFeatureService()
	return &CompositeService{
		featureDecisionServices: []FeatureService{featureDecisionService},
		notificationCenter:      notificationCenter,
	}
}

// GetFeatureDecision returns a decision for the given feature key
func (s CompositeService) GetFeatureDecision(featureDecisionContext FeatureDecisionContext, userContext entities.UserContext) (FeatureDecision, error) {
	var featureDecision FeatureDecision
	var err error
	// loop through the different features decision services until we get a decision
	for _, decisionService := range s.featureDecisionServices {
		featureDecision, err = decisionService.GetDecision(featureDecisionContext, userContext)
		if err != nil {
			// @TODO: log error
			func() {}() // cheat linters
		}

		if featureDecision.Variation != nil {
			break
		}
	}

	// @TODO: add errors
	if s.notificationCenter != nil {
		featureMap := map[string]interface{}{
			"feature_key":     featureDecisionContext.Feature.Key,
			"feature_enabled": false,
			"source":          featureDecision.Source,
		}
		if featureDecision.Variation != nil {
			featureMap["feature_enabled"] = featureDecision.Variation.FeatureEnabled
		}

		decisionInfo := map[string]interface{}{
			"feature": featureMap,
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

// OnDecision registers a handler for Decision notifications
func (s CompositeService) OnDecision(callback func(notification.DecisionNotification)) {
	handler := func(payload interface{}) {
		if decisionNotification, ok := payload.(notification.DecisionNotification); ok {
			callback(decisionNotification)
		} else {
			csLogger.Warning(fmt.Sprintf("Unable to convert notification payload %v into DecisionNotification", payload))
		}
	}
	if _, err := s.notificationCenter.AddHandler(notification.Decision, handler); err != nil {
		csLogger.Warning("Problem with adding notification handler")
	}
}
