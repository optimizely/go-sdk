/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

// Package client has client definitions
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
	"github.com/optimizely/go-sdk/pkg/utils"

	"github.com/hashicorp/go-multierror"
)

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	ConfigManager      config.ProjectConfigManager
	DecisionService    decision.Service
	EventProcessor     event.Processor
	notificationCenter notification.Center
	execGroup          *utils.ExecGroup
	logger             logging.OptimizelyLogProducer
}

// Activate returns the key of the variation the user is bucketed into and queues up an impression event to be sent to
// the Optimizely log endpoint for results processing.
func (o *OptimizelyClient) Activate(experimentKey string, userContext entities.UserContext) (result string, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("Activate call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	decisionContext, experimentDecision, err := o.getExperimentDecision(experimentKey, userContext)
	if err != nil {
		o.logger.Error("received an error while computing experiment decision", err)
		return result, err
	}

	if experimentDecision.Variation != nil && decisionContext.Experiment != nil {
		// send an impression event
		result = experimentDecision.Variation.Key
		if ue, ok := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, *decisionContext.Experiment,
			experimentDecision.Variation, userContext, "", experimentKey, "experiment", true); ok {
			o.EventProcessor.ProcessEvent(ue)
		}
	}

	return result, err
}

// IsFeatureEnabled returns true if the feature is enabled for the given user. If the user is part of a feature test
// then an impression event will be queued up to be sent to the Optimizely log endpoint for results processing.
func (o *OptimizelyClient) IsFeatureEnabled(featureKey string, userContext entities.UserContext) (result bool, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("IsFeatureEnabled call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", userContext)
	if err != nil {
		o.logger.Error("received an error while computing feature decision", err)
		return result, err
	}

	if featureDecision.Variation == nil {
		result = false
	} else {
		result = featureDecision.Variation.FeatureEnabled
	}

	if result {
		o.logger.Info(fmt.Sprintf(`Feature "%s" is enabled for user "%s".`, featureKey, userContext.ID))
	} else {
		o.logger.Info(fmt.Sprintf(`Feature "%s" is not enabled for user "%s".`, featureKey, userContext.ID))
	}

	if o.notificationCenter != nil {
		decisionNotification := decision.FeatureNotification(featureKey, &featureDecision, &userContext)

		if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}

	if ue, ok := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment,
		featureDecision.Variation, userContext, featureKey, featureDecision.Experiment.Key, featureDecision.Source, result); ok && featureDecision.Source != "" {
		o.EventProcessor.ProcessEvent(ue)
	}

	return result, err
}

// GetEnabledFeatures returns an array containing the keys of all features in the project that are enabled for the given
// user. For features tests, impression events will be queued up to be sent to the Optimizely log endpoint for results processing.
func (o *OptimizelyClient) GetEnabledFeatures(userContext entities.UserContext) (enabledFeatures []string, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("GetEnabledFeatures call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	projectConfig, err := o.getProjectConfig()
	if err != nil {
		o.logger.Error("Error retrieving ProjectConfig", err)
		return enabledFeatures, err
	}

	featureList := projectConfig.GetFeatureList()
	for _, feature := range featureList {
		if isEnabled, _ := o.IsFeatureEnabled(feature.Key, userContext); isEnabled {
			enabledFeatures = append(enabledFeatures, feature.Key)
		}
	}
	return enabledFeatures, err
}

// GetFeatureVariableBoolean returns the feature variable value of type bool associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey, variableKey string, userContext entities.UserContext) (convertedValue bool, err error) {

	stringValue, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)
	defer func() {
		if o.notificationCenter != nil {
			variableMap := map[string]interface{}{
				"variableKey":   variableKey,
				"variableType":  variableType,
				"variableValue": convertedValue,
			}
			if err != nil {
				variableMap["variableValue"] = stringValue
			}
			decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
			decisionNotification.Type = notification.FeatureVariable

			if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
				o.logger.Warning("Problem with sending notification")
			}
		}
	}()

	if err != nil {
		return convertedValue, err
	}
	convertedValue, err = strconv.ParseBool(stringValue)
	if err != nil || variableType != entities.Boolean {
		return false, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}
	return convertedValue, err
}

// GetFeatureVariableDouble returns the feature variable value of type double associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey, variableKey string, userContext entities.UserContext) (convertedValue float64, err error) {

	stringValue, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)
	defer func() {
		if o.notificationCenter != nil {
			variableMap := map[string]interface{}{
				"variableKey":   variableKey,
				"variableType":  variableType,
				"variableValue": convertedValue,
			}
			if err != nil {
				variableMap["variableValue"] = stringValue
			}
			decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
			decisionNotification.Type = notification.FeatureVariable

			if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
				o.logger.Warning("Problem with sending notification")
			}
		}
	}()
	if err != nil {
		return convertedValue, err
	}
	convertedValue, err = strconv.ParseFloat(stringValue, 64)
	if err != nil || variableType != entities.Double {
		return 0, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}

	return convertedValue, err
}

// GetFeatureVariableInteger returns the feature variable value of type int associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey, variableKey string, userContext entities.UserContext) (convertedValue int, err error) {

	stringValue, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)
	defer func() {
		if o.notificationCenter != nil {
			variableMap := map[string]interface{}{
				"variableKey":   variableKey,
				"variableType":  variableType,
				"variableValue": convertedValue,
			}
			if err != nil {
				variableMap["variableValue"] = stringValue
			}
			decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
			decisionNotification.Type = notification.FeatureVariable

			if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
				o.logger.Warning("Problem with sending notification")
			}
		}
	}()
	if err != nil {
		return convertedValue, err
	}
	convertedValue, err = strconv.Atoi(stringValue)
	if err != nil || variableType != entities.Integer {
		return 0, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}

	return convertedValue, err
}

// GetFeatureVariableString returns the feature variable value of type string associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableString(featureKey, variableKey string, userContext entities.UserContext) (stringValue string, err error) {

	stringValue, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)

	defer func() {
		if o.notificationCenter != nil {
			variableMap := map[string]interface{}{
				"variableKey":   variableKey,
				"variableType":  variableType,
				"variableValue": stringValue,
			}

			decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
			decisionNotification.Type = notification.FeatureVariable

			if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
				o.logger.Warning("Problem with sending notification")
			}
		}
	}()
	if err != nil {
		return "", err
	}
	if variableType != entities.String {
		return "", fmt.Errorf("variable value for key %s is wrong type", variableKey)
	}

	return stringValue, err
}

// GetFeatureVariableJSON returns the feature variable value of type json associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableJSON(featureKey, variableKey string, userContext entities.UserContext) (optlyJSON *optimizelyjson.OptimizelyJSON, err error) {

	stringVal, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)
	defer func() {
		if o.notificationCenter != nil {
			var variableValue interface{}
			if optlyJSON != nil {
				variableValue = optlyJSON.ToMap()
			} else {
				variableValue = stringVal
			}
			variableMap := map[string]interface{}{
				"variableKey":   variableKey,
				"variableType":  variableType,
				"variableValue": variableValue,
			}
			decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
			decisionNotification.Type = notification.FeatureVariable

			if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
				o.logger.Warning("Problem with sending notification")
			}
		}
	}()
	if err != nil {
		return optlyJSON, err
	}

	optlyJSON, err = optimizelyjson.NewOptimizelyJSONfromString(stringVal)
	if err != nil || variableType != entities.JSON {
		optlyJSON, err = nil, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}

	return optlyJSON, err
}

// getFeatureVariable is a helper function, returns feature variable as a string along with it's associated type and feature decision
func (o *OptimizelyClient) getFeatureVariable(featureKey, variableKey string, userContext entities.UserContext) (string, entities.VariableType, *decision.FeatureDecision, error) {

	featureDecisionContext, featureDecision, err := o.getFeatureDecision(featureKey, variableKey, userContext)
	if err != nil {
		return "", "", &featureDecision, err
	}

	variable := featureDecisionContext.Variable

	if featureDecision.Variation != nil {
		if v, ok := featureDecision.Variation.Variables[variable.ID]; ok && featureDecision.Variation.FeatureEnabled {
			return v.Value, variable.Type, &featureDecision, nil
		}
	}

	return variable.DefaultValue, variable.Type, &featureDecision, nil
}

// GetFeatureVariable returns feature variable as a string along with it's associated type.
func (o *OptimizelyClient) GetFeatureVariable(featureKey, variableKey string, userContext entities.UserContext) (string, entities.VariableType, error) {

	stringValue, variableType, featureDecision, err := o.getFeatureVariable(featureKey, variableKey, userContext)

	func() {
		var convertedValue interface{}
		var e error

		convertedValue = stringValue
		switch variableType {
		case entities.Integer:
			convertedValue, e = strconv.Atoi(stringValue)
		case entities.Double:
			convertedValue, e = strconv.ParseFloat(stringValue, 64)
		case entities.Boolean:
			convertedValue, e = strconv.ParseBool(stringValue)
		case entities.JSON:
			convertedValue = map[string]string{}
			e = json.Unmarshal([]byte(stringValue), &convertedValue)
		}
		if e != nil {
			o.logger.Warning("Problem with converting string value to proper type for notification builder")
		}

		variableMap := map[string]interface{}{
			"variableKey":   variableKey,
			"variableType":  variableType,
			"variableValue": convertedValue,
		}
		decisionNotification := decision.FeatureNotificationWithVariables(featureKey, featureDecision, &userContext, variableMap)
		decisionNotification.Type = notification.FeatureVariable

		if e = o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}()

	return stringValue, variableType, err
}

// GetAllFeatureVariablesWithDecision returns all the variables for a given feature along with the enabled state.
func (o *OptimizelyClient) GetAllFeatureVariablesWithDecision(featureKey string, userContext entities.UserContext) (enabled bool, variableMap map[string]interface{}, err error) {

	variableMap = make(map[string]interface{})
	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", userContext)
	if err != nil {
		o.logger.Error("Optimizely SDK tracking error", err)
		return enabled, variableMap, err
	}

	if featureDecision.Variation != nil {
		enabled = featureDecision.Variation.FeatureEnabled
	}

	feature := decisionContext.Feature
	if feature == nil {
		o.logger.Warning(fmt.Sprintf(`feature "%s" does not exist`, featureKey))
		return enabled, variableMap, nil
	}

	errs := new(multierror.Error)

	for _, v := range feature.VariableMap {
		val := v.DefaultValue

		if enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				val = variable.Value
			}
		}

		typedValue, typedError := o.getTypedValue(val, v.Type)
		errs = multierror.Append(errs, typedError)
		variableMap[v.Key] = typedValue
	}

	if o.notificationCenter != nil {
		decisionNotification := decision.FeatureNotificationWithVariables(featureKey, &featureDecision, &userContext,
			map[string]interface{}{"variableValues": variableMap})
		decisionNotification.Type = notification.AllFeatureVariables

		if err = o.notificationCenter.Send(notification.Decision, *decisionNotification); err != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}
	return enabled, variableMap, errs.ErrorOrNil()
}

// GetDetailedFeatureDecisionUnsafe triggers an impression event and returns all the variables
// for a given feature along with the experiment key, variation key and the enabled state.
// Usage of this method is unsafe and not recommended since it can be removed in any of the next releases.
func (o *OptimizelyClient) GetDetailedFeatureDecisionUnsafe(featureKey string, userContext entities.UserContext, disableTracking bool) (decisionInfo decision.UnsafeFeatureDecisionInfo, err error) {

	decisionInfo = decision.UnsafeFeatureDecisionInfo{}
	decisionInfo.VariableMap = make(map[string]interface{})
	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", userContext)
	if err != nil {
		o.logger.Error("Optimizely SDK tracking error", err)
		return decisionInfo, err
	}

	if featureDecision.Variation != nil {
		decisionInfo.Enabled = featureDecision.Variation.FeatureEnabled

		// This information is only necessary for feature tests.
		// For rollouts experiments and variations are an implementation detail only.
		if featureDecision.Source == decision.FeatureTest {
			decisionInfo.VariationKey = featureDecision.Variation.Key
			decisionInfo.ExperimentKey = featureDecision.Experiment.Key
		}

		// Triggers impression events when applicable
		if !disableTracking {
			// send impression event for feature tests
			if ue, ok := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment,
				featureDecision.Variation, userContext, featureKey, featureDecision.Experiment.Key, featureDecision.Source, decisionInfo.Enabled); ok {
				o.EventProcessor.ProcessEvent(ue)
			}
		}
	}

	feature := decisionContext.Feature
	if feature == nil {
		o.logger.Warning(fmt.Sprintf(`feature "%s" does not exist`, featureKey))
		return decisionInfo, nil
	}

	errs := new(multierror.Error)

	for _, v := range feature.VariableMap {
		val := v.DefaultValue

		if decisionInfo.Enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				val = variable.Value
			} else {
				o.logger.Warning(fmt.Sprintf(`variable with id "%s" does not exist`, v.ID))
			}
		}

		typedValue, typedError := o.getTypedValue(val, v.Type)
		errs = multierror.Append(errs, typedError)
		decisionInfo.VariableMap[v.Key] = typedValue
	}

	if o.notificationCenter != nil {
		decisionNotification := decision.FeatureNotificationWithVariables(featureKey, &featureDecision, &userContext,
			map[string]interface{}{"variableValues": decisionInfo.VariableMap})
		decisionNotification.Type = notification.AllFeatureVariables

		if err = o.notificationCenter.Send(notification.Decision, *decisionNotification); err != nil {
			o.logger.Warning(fmt.Sprintf("Problem with sending notification: %v", err))
		}
	}
	return decisionInfo, errs.ErrorOrNil()
}

// GetAllFeatureVariables returns all the variables as OptimizelyJSON object for a given feature.
func (o *OptimizelyClient) GetAllFeatureVariables(featureKey string, userContext entities.UserContext) (optlyJSON *optimizelyjson.OptimizelyJSON, err error) {
	_, variableMap, err := o.GetAllFeatureVariablesWithDecision(featureKey, userContext)
	if err != nil {
		return optlyJSON, err
	}
	optlyJSON = optimizelyjson.NewOptimizelyJSONfromMap(variableMap)
	return optlyJSON, nil
}

// GetVariation returns the key of the variation the user is bucketed into. Does not generate impression events.
func (o *OptimizelyClient) GetVariation(experimentKey string, userContext entities.UserContext) (result string, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("GetVariation call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, experimentDecision, err := o.getExperimentDecision(experimentKey, userContext)
	if err != nil {
		o.logger.Error("received an error while computing experiment decision", err)
	}

	if experimentDecision.Variation != nil {
		result = experimentDecision.Variation.Key
	}

	return result, err
}

// Track generates a conversion event with the given event key if it exists and queues it up to be sent to the Optimizely
// log endpoint for results processing.
func (o *OptimizelyClient) Track(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("Track call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		o.logger.Error("Optimizely SDK tracking error", e)
		return e
	}

	configEvent, e := projectConfig.GetEventByKey(eventKey)

	if e != nil {
		errorMessage := fmt.Sprintf(`Unable to get event for key "%s": %s`, eventKey, e)
		o.logger.Warning(errorMessage)
		return nil
	}

	userEvent := event.CreateConversionUserEvent(projectConfig, configEvent, userContext, eventTags)
	if o.EventProcessor.ProcessEvent(userEvent) && o.notificationCenter != nil {
		trackNotification := notification.TrackNotification{EventKey: eventKey, UserContext: userContext, EventTags: eventTags, ConversionEvent: *userEvent.Conversion}
		if err = o.notificationCenter.Send(notification.Track, trackNotification); err != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}

	return nil
}

func (o *OptimizelyClient) getFeatureDecision(featureKey, variableKey string, userContext entities.UserContext) (decisionContext decision.FeatureDecisionContext, featureDecision decision.FeatureDecision, err error) {

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case error:
				err = t
			case string:
				err = errors.New(t)
			default:
				err = errors.New("unexpected error")
			}
			errorMessage := fmt.Sprintf("getFeatureDecision call, optimizely SDK is panicking with the error:")
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	userID := userContext.ID
	o.logger.Debug(fmt.Sprintf(`Evaluating feature "%s" for user "%s".`, featureKey, userID))

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		o.logger.Error("Error calling getFeatureDecision", e)
		return decisionContext, featureDecision, e
	}

	decisionContext.ProjectConfig = projectConfig
	feature, e := projectConfig.GetFeatureByKey(featureKey)
	if e != nil {
		o.logger.Warning(fmt.Sprintf(`Could not get feature for key "%s": %s`, featureKey, e))
		return decisionContext, featureDecision, nil
	}

	decisionContext.Feature = &feature
	variable := entities.Variable{}
	if variableKey != "" {
		variable, err = projectConfig.GetVariableByKey(feature.Key, variableKey)
		if err != nil {
			o.logger.Warning(fmt.Sprintf(`Could not get variable for key "%s": %s`, variableKey, err))
			return decisionContext, featureDecision, nil
		}
	}

	decisionContext.Variable = variable
	featureDecision, err = o.DecisionService.GetFeatureDecision(decisionContext, userContext)
	if err != nil {
		o.logger.Warning(fmt.Sprintf(`Received error while making a decision for feature "%s": %s`, featureKey, err))
		return decisionContext, featureDecision, nil
	}

	return decisionContext, featureDecision, nil
}

func (o *OptimizelyClient) getExperimentDecision(experimentKey string, userContext entities.UserContext) (decisionContext decision.ExperimentDecisionContext, experimentDecision decision.ExperimentDecision, err error) {

	userID := userContext.ID
	o.logger.Debug(fmt.Sprintf(`Evaluating experiment "%s" for user "%s".`, experimentKey, userID))

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		return decisionContext, experimentDecision, e
	}

	experiment, e := projectConfig.GetExperimentByKey(experimentKey)
	if e != nil {
		o.logger.Warning(fmt.Sprintf(`Could not get experiment for key "%s": %s`, experimentKey, e))
		return decisionContext, experimentDecision, nil
	}

	decisionContext = decision.ExperimentDecisionContext{
		Experiment:    &experiment,
		ProjectConfig: projectConfig,
	}

	experimentDecision, err = o.DecisionService.GetExperimentDecision(decisionContext, userContext)
	if err != nil {
		o.logger.Warning(fmt.Sprintf(`Received error while making a decision for experiment "%s": %s`, experimentKey, err))
		return decisionContext, experimentDecision, nil
	}

	if experimentDecision.Variation != nil {
		result := experimentDecision.Variation.Key
		o.logger.Info(fmt.Sprintf(`User "%s" is bucketed into variation "%s" of experiment "%s".`, userContext.ID, result, experimentKey))
	} else {
		o.logger.Info(fmt.Sprintf(`User "%s" is not bucketed into any variation for experiment "%s": %s.`, userContext.ID, experimentKey, experimentDecision.Reason))
	}

	return decisionContext, experimentDecision, err
}

// OnTrack registers a handler for Track notifications
func (o *OptimizelyClient) OnTrack(callback func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent)) (int, error) {
	if o.notificationCenter == nil {
		return 0, fmt.Errorf("no notification center found")
	}

	handler := func(payload interface{}) {
		success := false
		if trackNotification, ok := payload.(notification.TrackNotification); ok {
			if conversionEvent, ok := trackNotification.ConversionEvent.(event.ConversionEvent); ok {
				success = true
				callback(trackNotification.EventKey, trackNotification.UserContext, trackNotification.EventTags, conversionEvent)
			}
		}
		if !success {
			o.logger.Warning(fmt.Sprintf("Unable to convert notification payload %v into TrackNotification", payload))
		}
	}
	id, err := o.notificationCenter.AddHandler(notification.Track, handler)
	if err != nil {
		o.logger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnTrack removes handler for Track notification with given id
func (o *OptimizelyClient) RemoveOnTrack(id int) error {
	if o.notificationCenter == nil {
		return fmt.Errorf("no notification center found")
	}
	if err := o.notificationCenter.RemoveHandler(id, notification.Track); err != nil {
		o.logger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}

func (o *OptimizelyClient) getTypedValue(value string, variableType entities.VariableType) (convertedValue interface{}, err error) {
	convertedValue = value
	switch variableType {
	case entities.Boolean:
		convertedValue, err = strconv.ParseBool(value)
	case entities.Double:
		convertedValue, err = strconv.ParseFloat(value, 64)
	case entities.Integer:
		convertedValue, err = strconv.Atoi(value)
	case entities.JSON:
		var optlyJSON *optimizelyjson.OptimizelyJSON
		optlyJSON, err = optimizelyjson.NewOptimizelyJSONfromString(value)
		convertedValue = optlyJSON.ToMap()
	case entities.String:
	default:
		o.logger.Warning(fmt.Sprintf(`type "%s" is unknown, returning string`, variableType))
	}
	return convertedValue, err
}

func (o *OptimizelyClient) getProjectConfig() (projectConfig config.ProjectConfig, err error) {

	if isNil(o.ConfigManager) {
		return nil, errors.New("project config manager is not initialized")
	}
	projectConfig, err = o.ConfigManager.GetConfig()
	if err != nil {
		return nil, err
	}

	return projectConfig, nil
}

// GetOptimizelyConfig returns OptimizelyConfig object
func (o *OptimizelyClient) GetOptimizelyConfig() (optimizelyConfig *config.OptimizelyConfig) {

	return o.ConfigManager.GetOptimizelyConfig()

}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components.
func (o *OptimizelyClient) Close() {
	o.execGroup.TerminateAndWait()
}

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
