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
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// TestCompositeService represents a CompositeService with custom implementations
type TestCompositeService struct {
	decision.CompositeService
	listenersCalled []models.DecisionListener
}

// AddListeners - Adds Notification Listeners
func (c *TestCompositeService) AddListeners(listeners map[string]int) {

	if len(listeners) < 1 {
		return
	}
	for listenerType, count := range listeners {
		for i := 1; i <= count; i++ {
			switch listenerType {
			case "Decision":
				c.OnDecision(c.decisionNotificationCallback)
				break
			default:
				break
			}
		}
	}
}

// GetListenersCalled - Returns listeners called
func (c *TestCompositeService) GetListenersCalled() []models.DecisionListener {
	listenerCalled := c.listenersCalled
	// Since for every scenario, a new sdk instance is created, emptying listenersCalled is required for scenario's
	// where multiple requests are executed but no session is to be maintained among them.
	// @TODO: Make it optional once event-batching(sessioned) tests are implemented.
	c.listenersCalled = nil
	return listenerCalled
}

func (c *TestCompositeService) decisionNotificationCallback(notification notification.DecisionNotification) {

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
	c.listenersCalled = append(c.listenersCalled, model)
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
