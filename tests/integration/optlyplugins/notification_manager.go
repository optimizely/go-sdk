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

package optlyplugins

import (
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// NotificationManager manager class for notification listeners
type NotificationManager struct {
	DecisionService *decision.CompositeService
	Client          *client.OptimizelyClient
	listenersCalled []interface{}
}

// SubscribeNotifications subscribes to the provided notification listeners
func (n *NotificationManager) SubscribeNotifications(listeners map[string]int) {

	addNotificationCallback := func(notificationType string) {
		switch notificationType {
		case models.KeyDecision:
			n.DecisionService.OnDecision(n.decisionCallback)
			break
		case models.KeyTrack:
			n.Client.OnTrack(n.trackCallback)
			break
		}
	}

	for key, count := range listeners {
		for i := 0; i < count; i++ {
			addNotificationCallback(key)
		}
	}
}

func (n *NotificationManager) decisionCallback(notification notification.DecisionNotification) {

	model := models.DecisionListener{}
	model.Type = notification.Type
	model.UserID = notification.UserContext.ID
	if notification.UserContext.Attributes == nil {
		model.Attributes = make(map[string]interface{})
	} else {
		model.Attributes = notification.UserContext.Attributes
	}

	decisionInfoDict := getDecisionInfoForNotification(notification)
	model.DecisionInfo = decisionInfoDict
	n.listenersCalled = append(n.listenersCalled, model)
}

func (n *NotificationManager) trackCallback(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
	listener := models.TrackListener{
		EventKey:   eventKey,
		UserID:     userContext.ID,
		Attributes: userContext.Attributes,
		EventTags:  eventTags,
	}
	n.listenersCalled = append(n.listenersCalled, listener)
}

func getDecisionInfoForNotification(decisionNotification notification.DecisionNotification) map[string]interface{} {
	decisionInfoDict := make(map[string]interface{})

	updateSourceInfo := func(source string) {
		decisionInfoDict["source_info"] = make(map[string]interface{})
		if source == string(decision.FeatureTest) {
			featureInfoDict := decisionNotification.DecisionInfo["feature"].(map[string]interface{})
			if sourceInfo, ok := featureInfoDict["sourceInfo"].(interface{}); ok {
				sourceInfoDict := sourceInfo.((map[string]string))
				if experimentKey, ok := sourceInfoDict["experimentKey"]; ok {
					if variationKey, ok := sourceInfoDict["variationKey"]; ok {
						dict := make(map[string]interface{})
						dict["experiment_key"] = experimentKey
						dict["variation_key"] = variationKey
						decisionInfoDict["source_info"] = dict
					}
				}
			}
		}
	}

	switch notificationType := decisionNotification.Type; notificationType {
	case notification.ABTest, notification.FeatureTest:
		decisionInfoDict["experiment_key"] = decisionNotification.DecisionInfo["experimentKey"]
		decisionInfoDict["variation_key"] = decisionNotification.DecisionInfo["variationKey"]
		break
	case notification.Feature:
		featureInfoDict := decisionNotification.DecisionInfo["feature"].(map[string]interface{})
		source := ""
		if decisionSource, ok := featureInfoDict["source"].(decision.Source); ok {
			source = string(decisionSource)
		} else {
			source = featureInfoDict["source"].(string)
		}
		decisionInfoDict["source"] = source
		decisionInfoDict["feature_enabled"] = featureInfoDict["featureEnabled"]
		decisionInfoDict["feature_key"] = featureInfoDict["featureKey"]
		updateSourceInfo(source)
	case notification.FeatureVariable:
		featureInfoDict := decisionNotification.DecisionInfo["feature"].(map[string]interface{})
		source := ""
		if decisionSource, ok := featureInfoDict["source"].(decision.Source); ok {
			source = string(decisionSource)
		} else {
			source = featureInfoDict["source"].(string)
		}
		decisionInfoDict["source"] = source
		decisionInfoDict["variable_key"] = featureInfoDict["variableKey"]
		if variableType, ok := featureInfoDict["variableType"].(entities.VariableType); ok {
			decisionInfoDict["variable_type"] = string(variableType)
		} else {
			decisionInfoDict["variable_type"] = featureInfoDict["variableType"].(string)
		}
		decisionInfoDict["variable_value"] = featureInfoDict["variableValue"]
		decisionInfoDict["feature_enabled"] = featureInfoDict["featureEnabled"]
		decisionInfoDict["feature_key"] = featureInfoDict["featureKey"]
		updateSourceInfo(source)
	default:
	}
	return decisionInfoDict
}

// GetListenersCalled - Returns listeners called
func (n *NotificationManager) GetListenersCalled() []interface{} {
	listenerCalled := n.listenersCalled
	// Since for every scenario, a new sdk instance is created, emptying listenersCalled is required for scenario's
	// where multiple requests are executed but no session is to be maintained among them.
	// @TODO: Make it optional once event-batching(sessioned) tests are implemented.
	n.listenersCalled = nil
	return listenerCalled
}
