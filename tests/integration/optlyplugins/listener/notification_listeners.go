package listener

import (
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/notification"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// AddListener - Adds Notification Listeners
func AddListener(decisionService decision.Service, params *models.RequestParams) (getListenersCalled func() []models.DecisionListener) {
	var listenersCalled []models.DecisionListener
	getListenersCalled = func() []models.DecisionListener {
		return listenersCalled
	}
	if len(params.Listeners) < 1 {
		return
	}

	for listenerType, count := range params.Listeners {
		for i := 1; i <= count; i++ {
			switch listenerType {
			case "Decision":
				callback := func(notification notification.DecisionNotification) {

					model := models.DecisionListener{}
					model.Type = notification.Type
					model.UserID = notification.UserContext.ID
					if notification.UserContext.Attributes == nil {
						model.Attributes = make(map[string]interface{})
					} else {
						model.Attributes = notification.UserContext.Attributes
					}

					decisionInfoDict := make(map[string]interface{})
					switch notificationType := notification.Type; notificationType {
					case "feature":
						decisionInfoDict = notification.DecisionInfo["feature"].(map[string]interface{})
						source := string(decisionInfoDict["source"].(decision.Source))
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

					model.DecisionInfo = decisionInfoDict
					listenersCalled = append(listenersCalled, model)
				}
				decisionService.OnDecision(callback)
				break
			default:
				break
			}
		}

	}
	return getListenersCalled
}
