package client

import (
	"fmt"
	"runtime/debug"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
)

func (o *OptimizelyClient) Decide(key string, user *entities.UserContext, options *entities.OptimizelyDecideOptions) (optlyDecision *decision.OptimizelyDecision) {
	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf("Decide call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, r)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	if user == nil {
		return decision.ErrorDecision(key, user, reasons.UserNotSet)
	}
	projectConfig, e := o.getProjectConfig()
	if e != nil {
		return decision.ErrorDecision(key, user, reasons.SdkNotReady)
	}
	var isFeatureKey, isExperimentKey bool

	_, featErr := projectConfig.GetFeatureByKey(key)
	_, expErr := projectConfig.GetExperimentByKey(key)

	isFeatureKey = featErr == nil
	isExperimentKey = expErr == nil

	if options.ForExperiment {
		isFeatureKey = false
		isExperimentKey = true
	}

	if isExperimentKey && !isFeatureKey {
		return o.experimentDecide(projectConfig, key, user, options)
	}

	return o.featureDecide(projectConfig, key, user, options)
}

func (o *OptimizelyClient) DecideAll(keys []string, user *entities.UserContext, options *entities.OptimizelyDecideOptions) (optlyDecisions map[string]*decision.OptimizelyDecision) {
	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf("Decide call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, r)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	if user == nil {
		o.logger.Error(string(reasons.UserNotSet), nil)
		return optlyDecisions
	}

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		o.logger.Error(string(reasons.SdkNotReady), nil)
		return optlyDecisions
	}

	optlyDecisions = map[string]*decision.OptimizelyDecision{}
	for _, key := range keys {
		var isFeatureKey, isExperimentKey bool

		_, featErr := projectConfig.GetFeatureByKey(key)
		_, expErr := projectConfig.GetExperimentByKey(key)

		isFeatureKey = featErr == nil
		isExperimentKey = expErr == nil

		if options.ForExperiment {
			isFeatureKey = false
			isExperimentKey = true
		}

		if isExperimentKey && !isFeatureKey {
			optlyDecisions[key] = o.experimentDecide(projectConfig, key, user, options)
		}

		optlyDecisions[key] = o.featureDecide(projectConfig, key, user, options)
		decision := o.featureDecide(projectConfig, key, user, options)

		if !options.EnabledOnly || (decision.Enabled) {
			optlyDecisions[key] = decision
		}
	}

	return optlyDecisions
}

func (o *OptimizelyClient) featureDecide(projectConfig config.ProjectConfig, featureKey string, userContext *entities.UserContext, options *entities.OptimizelyDecideOptions) (optlyDecision *decision.OptimizelyDecision) {
	_, featErr := projectConfig.GetFeatureByKey(featureKey)
	if featErr != nil {
		return decision.ErrorDecision(featureKey, userContext, reasons.Reason(fmt.Sprintf(string(reasons.FeatureKeyInvalid), featureKey)))
	}
	var variationKey string
	var enabled, tracked bool

	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", *userContext)
	if err != nil {
		return decision.ErrorDecision(featureKey, userContext, reasons.FeatureDecisionError)
	}

	if featureDecision.Variation == nil {
		enabled = false
		variationKey = ""
	} else {
		enabled = featureDecision.Variation.FeatureEnabled
		variationKey = featureDecision.Variation.Key
	}

	variableMap := map[string]interface{}{}
	feature := decisionContext.Feature

	if feature == nil {
		return decision.ErrorDecision(featureKey, userContext, reasons.FeatureKeyInvalid)
	}

	decideReasons := []reasons.Reason{}
	for _, v := range feature.VariableMap {
		val := v.DefaultValue

		if enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				val = variable.Value
			}
		}

		typedValue, typedError := o.getTypedValue(val, v.Type)
		if typedError != nil {
			decideReasons = append(decideReasons, reasons.Reason(fmt.Sprintf(string(reasons.VariableValueInvalid), v.Key)))
		}
		variableMap[v.Key] = typedValue
	}
	optlyJSON := optimizelyjson.NewOptimizelyJSONfromMap(variableMap)
	if optlyJSON == nil {
		decideReasons = append(decideReasons, reasons.InvalidJSONVariable)
	}

	if !options.DisableTracking && featureDecision.Variation != nil {
		impressionEvent := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment, *featureDecision.Variation, *userContext)
		o.EventProcessor.ProcessEvent(impressionEvent)
		tracked = true
	}

	if o.notificationCenter != nil {
		decisionNotification := decision.FeatureNotificationWithVariablesForDecide(featureKey, tracked, &featureDecision, userContext,
			map[string]interface{}{"variableValues": variableMap})
		decisionNotification.Type = notification.FlagDecide

		if err = o.notificationCenter.Send(notification.Decision, *decisionNotification); err != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}

	return &decision.OptimizelyDecision{VariationKey: variationKey,
		Enabled: enabled, Variables: optlyJSON,
		Key: feature.Key, User: userContext, Reasons: decideReasons}
}

func (o *OptimizelyClient) experimentDecide(projectConfig config.ProjectConfig, experimentKey string, userContext *entities.UserContext, options *entities.OptimizelyDecideOptions) (optlyDecision *decision.OptimizelyDecision) {
	experiment, featErr := projectConfig.GetExperimentByKey(experimentKey)
	if featErr != nil {
		return decision.ErrorDecision(experimentKey, userContext, reasons.Reason(fmt.Sprintf(string(reasons.ExperimentKeyInvalid), experimentKey)))
	}
	var variationKey string
	var tracked bool

	decisionContext, experimentDecision, err := o.getExperimentDecision(experimentKey, *userContext)
	if err != nil {
		return decision.ErrorDecision(experimentKey, userContext, reasons.ExperimentDecisionError)
	}

	if !options.DisableTracking {
		impressionEvent := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, experiment, *experimentDecision.Variation, *userContext)
		o.EventProcessor.ProcessEvent(impressionEvent)
		tracked = true
	}

	if o.notificationCenter != nil && experimentDecision.Variation != nil {
		decisionInfo := map[string]interface{}{
			"experimentKey": decisionContext.Experiment.Key,
			"variationKey":  experimentDecision.Variation.Key,
			"tracked":       tracked,
		}

		decisionNotification := notification.DecisionNotification{
			DecisionInfo: decisionInfo,
			Type:         notification.ABTest,
			UserContext:  *userContext,
		}

		if err = o.notificationCenter.Send(notification.Decision, decisionNotification); err != nil {
			o.logger.Warning("Error sending sending notification")
		}
	}

	return &decision.OptimizelyDecision{VariationKey: variationKey,
		Enabled: false, Variables: nil,
		Key: experiment.Key, User: userContext, Reasons: []reasons.Reason{}}
}
