package listener

import (
	"github.com/optimizely/go-sdk/checking/integration/optimizely/datamodels"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/notification"
)

// AddListener - Adds Notification Listeners
func AddListener(decisionService decision.Service, params datamodels.RequestParams) (getListenersCalled func() []map[string]interface{}) {
	var listenersCalled []map[string]interface{}
	getListenersCalled = func() []map[string]interface{} {
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
					var listenerResponseMap = make(map[string]interface{})
					listenerResponseMap["type"] = notification.Type
					listenerResponseMap["user_id"] = notification.UserContext.ID

					if notification.UserContext.Attributes == nil {
						listenerResponseMap["attributes"] = make(map[string]interface{})
					} else {
						listenerResponseMap["attributes"] = notification.UserContext.Attributes
					}

					decisionInfoDict := make(map[string]interface{})

					switch notificationType := notification.Type; notificationType {
					case "feature":
						decisionInfoDict = notification.DecisionInfo["feature"].(map[string]interface{})
						decisionInfoDict["source_info"] = make(map[string]interface{})
						if source, ok := notification.DecisionInfo["source"]; ok && source == "feature-test" {
							if sourceInfo, ok := notification.DecisionInfo["source_info"].(map[string]interface{}); ok {
								if experimentKey, ok := sourceInfo["experiment_key"].(string); ok {
									if variationKey, ok := sourceInfo["variation_key"].(string); ok {
										decisionInfoDict["source_info"] =
											map[string]string{
												"experiment_key": experimentKey,
												"variation_key":  variationKey,
											}
									}
								}
							}
						}

					default:
					}

					listenerResponseMap["decision_info"] = decisionInfoDict
					listenersCalled = append(listenersCalled, listenerResponseMap)
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
