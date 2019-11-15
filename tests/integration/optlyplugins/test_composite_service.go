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
	return c.listenersCalled
}

func (c *TestCompositeService) decisionNotificationCallback(notify notification.DecisionNotification) {

	model := models.DecisionListener{}
	model.Type = notify.Type
	model.UserID = notify.UserContext.ID
	if notify.UserContext.Attributes == nil {
		model.Attributes = make(map[string]interface{})
	} else {
		model.Attributes = notify.UserContext.Attributes
	}

	decisionInfoDict := getDecisionInfoForNotification(notify)
	model.DecisionInfo = decisionInfoDict
	c.listenersCalled = append(c.listenersCalled, model)
}

func getDecisionInfoForNotification(notify notification.DecisionNotification) map[string]interface{} {
	decisionInfoDict := make(map[string]interface{})

	updateSourceInfo := func(source string) {
		decisionInfoDict["source_info"] = make(map[string]interface{})
		if source == string(decision.FeatureTest) {
			if sourceInfo, ok := notify.DecisionInfo["sourceInfo"].(map[string]string); ok {
				if experimentKey, ok := sourceInfo["experimentKey"]; ok {
					if variationKey, ok := sourceInfo["variationKey"]; ok {
						dict := make(map[string]interface{})
						dict["experiment_key"] = experimentKey
						dict["variation_key"] = variationKey
						decisionInfoDict["source_info"] = dict
					}
				}
			}
		}
	}

	switch notificationType := notify.Type; notificationType {
	case notification.ABTest, notification.FeatureTest:
		decisionInfoDict["experiment_key"] = notify.DecisionInfo["experimentKey"]
		decisionInfoDict["variation_key"] = notify.DecisionInfo["variationKey"]
		break
	case notification.Feature:
		source := ""
		if decisionSource, ok := notify.DecisionInfo["source"].(decision.Source); ok {
			source = string(decisionSource)
		} else {
			source = decisionInfoDict["source"].(string)
		}
		decisionInfoDict["source"] = source
		updateSourceInfo(source)
	case notification.FeatureVariable:
		source := ""
		if decisionSource, ok := notify.DecisionInfo["source"].(decision.Source); ok {
			source = string(decisionSource)
		} else {
			source = decisionInfoDict["source"].(string)
		}
		decisionInfoDict["source"] = source
		decisionInfoDict["variable_key"] = notify.DecisionInfo["variableKey"]
		decisionInfoDict["variable_type"] = notify.DecisionInfo["variableType"]
		decisionInfoDict["variable_value"] = notify.DecisionInfo["variableValue"]
		updateSourceInfo(source)
	default:
	}
	return decisionInfoDict
}
