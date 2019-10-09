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

func getDecisionInfoForNotification(notification notification.DecisionNotification) map[string]interface{} {
	decisionInfoDict := make(map[string]interface{})
	switch notificationType := notification.Type; notificationType {
	case "feature":
		decisionInfoDict = notification.DecisionInfo["feature"].(map[string]interface{})
		source := ""
		if decisionSource, ok := decisionInfoDict["source"].(decision.Source); ok {
			source = string(decisionSource)
		} else {
			source = decisionInfoDict["source"].(string)
		}
		decisionInfoDict["source"] = source

		if source == "feature-test" {
			if sourceInfo, ok := notification.DecisionInfo["source_info"].(map[string]interface{}); ok {
				if experimentKey, ok := sourceInfo["experiment_key"].(string); ok {
					if variationKey, ok := sourceInfo["variation_key"].(string); ok {
						dict := make(map[string]interface{})
						dict["experiment_key"] = experimentKey
						dict["variation_key"] = variationKey
						decisionInfoDict["source_info"] = dict
					}
				}
			}
		}

	default:
	}
	return decisionInfoDict
}
