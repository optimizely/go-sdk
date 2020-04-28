/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// FeatureNotificationWithVariables constructs feature notification with variables
func FeatureNotificationWithVariables(featureKey string, featureDecision *FeatureDecision, userContext *entities.UserContext,
	variables map[string]interface{}) *notification.DecisionNotification {

	decisionNotification := FeatureNotification(featureKey, featureDecision, userContext)

	if featureInfo, ok := decisionNotification.DecisionInfo["feature"].(map[string]interface{}); ok {
		for key, val := range variables {
			featureInfo[key] = val
		}
	}
	return decisionNotification
}

// FeatureNotification constructs default feature notification
func FeatureNotification(featureKey string, featureDecision *FeatureDecision, userContext *entities.UserContext) *notification.DecisionNotification {
	sourceInfo := map[string]string{}

	if featureDecision.Source == FeatureTest {
		sourceInfo["experimentKey"] = featureDecision.Experiment.Key
		sourceInfo["variationKey"] = featureDecision.Variation.Key
	}

	featureInfo := map[string]interface{}{
		"featureKey":     featureKey,
		"featureEnabled": false,
		"source":         featureDecision.Source,
		"sourceInfo":     sourceInfo,
	}
	if featureDecision.Variation != nil {
		featureInfo["featureEnabled"] = featureDecision.Variation.FeatureEnabled
	}

	notificationType := notification.Feature

	decisionInfo := map[string]interface{}{
		"feature": featureInfo,
	}

	decisionNotification := &notification.DecisionNotification{
		DecisionInfo: decisionInfo,
		Type:         notificationType,
		UserContext:  *userContext,
	}
	return decisionNotification
}
