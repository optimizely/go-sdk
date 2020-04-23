package decision

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

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
