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

// Package client has client definitions
package client

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"

	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/utils"
)

var logger = logging.GetLogger("Client")

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	ConfigManager      pkg.ProjectConfigManager
	DecisionService    decision.Service
	EventProcessor     event.Processor
	NotificationCenter notification.Center
	executionCtx       utils.ExecutionCtx
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	decisionContext, experimentDecision, err := o.getExperimentDecision(experimentKey, userContext)
	if err != nil {
		logger.Error("received an error while computing experiment decision", err)
		return result, err
	}

	if experimentDecision.Variation != nil {
		// send an impression event
		result = experimentDecision.Variation.Key
		impressionEvent := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, *decisionContext.Experiment, *experimentDecision.Variation, userContext)
		o.EventProcessor.ProcessEvent(impressionEvent)
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", userContext)
	if err != nil {
		logger.Error("received an error while computing feature decision", err)
		return result, err
	}

	if featureDecision.Variation == nil {
		result = false
	} else {
		result = featureDecision.Variation.FeatureEnabled
	}

	if result {
		logger.Info(fmt.Sprintf(`Feature "%s" is enabled for user "%s".`, featureKey, userContext.ID))
	} else {
		logger.Info(fmt.Sprintf(`Feature "%s" is not enabled for user "%s".`, featureKey, userContext.ID))
	}

	if featureDecision.Source == decision.FeatureTest {
		// send impression event for feature tests
		impressionEvent := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment, *featureDecision.Variation, userContext)
		o.EventProcessor.ProcessEvent(impressionEvent)
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	projectConfig, err := o.GetProjectConfig()
	if err != nil {
		logger.Error("Error retrieving ProjectConfig", err)
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
func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey, variableKey string, userContext entities.UserContext) (value bool, err error) {

	val, valueType, err := o.GetFeatureVariable(featureKey, variableKey, userContext)
	if err != nil {
		return false, err
	}
	convertedValue, err := strconv.ParseBool(val)
	if err != nil || valueType != entities.Boolean {
		return false, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}
	return convertedValue, err
}

// GetFeatureVariableDouble returns the feature variable value of type double associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey, variableKey string, userContext entities.UserContext) (value float64, err error) {

	val, valueType, err := o.GetFeatureVariable(featureKey, variableKey, userContext)
	if err != nil {
		return 0, err
	}
	convertedValue, err := strconv.ParseFloat(val, 64)
	if err != nil || valueType != entities.Double {
		return 0, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}
	return convertedValue, err
}

// GetFeatureVariableInteger returns the feature variable value of type int associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey, variableKey string, userContext entities.UserContext) (value int, err error) {

	val, valueType, err := o.GetFeatureVariable(featureKey, variableKey, userContext)
	if err != nil {
		return 0, err
	}
	convertedValue, err := strconv.Atoi(val)
	if err != nil || valueType != entities.Integer {
		return 0, fmt.Errorf("variable value for key %s is invalid or wrong type", variableKey)
	}
	return convertedValue, err
}

// GetFeatureVariableString returns the feature variable value of type string associated with the given feature and variable keys.
func (o *OptimizelyClient) GetFeatureVariableString(featureKey, variableKey string, userContext entities.UserContext) (value string, err error) {

	value, valueType, err := o.GetFeatureVariable(featureKey, variableKey, userContext)
	if err != nil {
		return "", err
	}
	if valueType != entities.String {
		return "", fmt.Errorf("variable value for key %s is wrong type", variableKey)
	}
	return value, err
}

// GetFeatureVariable returns feature variable as a string along with it's associated type.
func (o *OptimizelyClient) GetFeatureVariable(featureKey, variableKey string, userContext entities.UserContext) (value string, valueType entities.VariableType, err error) {

	featureDecisionContext, featureDecision, err := o.getFeatureDecision(featureKey, variableKey, userContext)
	if err != nil {
		return "", "", err
	}

	variable := featureDecisionContext.Variable

	if featureDecision.Variation != nil {
		if v, ok := featureDecision.Variation.Variables[variable.ID]; ok && featureDecision.Variation.FeatureEnabled {
			return v.Value, variable.Type, err
		}
	}

	return variable.DefaultValue, variable.Type, err
}

// GetAllFeatureVariables returns all the variables for a given feature along with the enabled state.
func (o *OptimizelyClient) GetAllFeatureVariables(featureKey string, userContext entities.UserContext) (enabled bool, variableMap map[string]string, err error) {

	variableMap = make(map[string]string)
	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, "", userContext)
	if err != nil {
		logger.Error("Optimizely SDK tracking error", err)
		return enabled, variableMap, err
	}

	if featureDecision.Variation != nil {
		enabled = featureDecision.Variation.FeatureEnabled
	}

	feature := decisionContext.Feature
	if feature == nil {
		logger.Warning(fmt.Sprintf(`feature "%s" does not exist`, featureKey))
		return enabled, variableMap, nil
	}

	for _, v := range feature.VariableMap {
		variableMap[v.Key] = v.DefaultValue

		if enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				variableMap[v.Key] = variable.Value
			}
		}
	}

	return enabled, variableMap, err
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	_, experimentDecision, err := o.getExperimentDecision(experimentKey, userContext)
	if err != nil {
		logger.Error("received an error while computing experiment decision", err)
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	projectConfig, e := o.GetProjectConfig()
	if e != nil {
		logger.Error("Optimizely SDK tracking error", e)
		return e
	}

	configEvent, e := projectConfig.GetEventByKey(eventKey)

	if e != nil {
		errorMessage := fmt.Sprintf(`Unable to get event for key "%s": %s`, eventKey, e)
		logger.Warning(errorMessage)
		return nil
	}

	userEvent := event.CreateConversionUserEvent(projectConfig, configEvent, userContext, eventTags)
	if o.EventProcessor.ProcessEvent(userEvent) && o.NotificationCenter != nil {
		trackNotification := notification.TrackNotification{Type: notification.Track, EventKey: eventKey, UserContext: userContext, EventTags: eventTags}
		var payload []interface{}
		payload = append(payload, trackNotification, userEvent)

		if err = o.NotificationCenter.Send(notification.Track, payload); err != nil {
			logger.Warning("Problem with sending notification")
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
			errorMessage := fmt.Sprintf("optimizely SDK is panicking with the error:")
			logger.Error(errorMessage, err)
			logger.Debug(string(debug.Stack()))
		}
	}()

	userID := userContext.ID
	logger.Debug(fmt.Sprintf(`Evaluating feature "%s" for user "%s".`, featureKey, userID))

	projectConfig, e := o.GetProjectConfig()
	if e != nil {
		logger.Error("Error calling getFeatureDecision", e)
		return decisionContext, featureDecision, e
	}

	feature, e := projectConfig.GetFeatureByKey(featureKey)
	if e != nil {
		logger.Warning(fmt.Sprintf(`Could not get feature for key "%s": %s`, featureKey, e))
		return decisionContext, featureDecision, nil
	}

	variable := entities.Variable{}
	if variableKey != "" {
		variable, err = projectConfig.GetVariableByKey(feature.Key, variableKey)
		if err != nil {
			logger.Warning(fmt.Sprintf(`Could not get variable for key "%s": %s`, variableKey, err))
			return decisionContext, featureDecision, nil
		}
	}

	decisionContext = decision.FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: projectConfig,
		Variable:      variable,
	}

	featureDecision, err = o.DecisionService.GetFeatureDecision(decisionContext, userContext)
	if err != nil {
		logger.Warning(fmt.Sprintf(`Received error while making a decision for feature "%s": %s`, featureKey, err))
		return decisionContext, featureDecision, nil
	}

	return decisionContext, featureDecision, nil
}

func (o *OptimizelyClient) getExperimentDecision(experimentKey string, userContext entities.UserContext) (decisionContext decision.ExperimentDecisionContext, experimentDecision decision.ExperimentDecision, err error) {

	userID := userContext.ID
	logger.Debug(fmt.Sprintf(`Evaluating experiment "%s" for user "%s".`, experimentKey, userID))

	projectConfig, e := o.GetProjectConfig()
	if e != nil {
		logger.Error("Error calling getExperimentDecision", e)
		return decisionContext, experimentDecision, e
	}

	experiment, e := projectConfig.GetExperimentByKey(experimentKey)
	if e != nil {
		logger.Warning(fmt.Sprintf(`Could not get experiment for key "%s": %s`, experimentKey, e))
		return decisionContext, experimentDecision, nil
	}

	decisionContext = decision.ExperimentDecisionContext{
		Experiment:    &experiment,
		ProjectConfig: projectConfig,
	}

	experimentDecision, err = o.DecisionService.GetExperimentDecision(decisionContext, userContext)
	if err != nil {
		logger.Warning(fmt.Sprintf(`Received error while making a decision for experiment "%s": %s`, experimentKey, err))
		return decisionContext, experimentDecision, nil
	}

	if experimentDecision.Variation != nil {
		result := experimentDecision.Variation.Key
		logger.Info(fmt.Sprintf(`User "%s" is bucketed into variation "%s" of experiment "%s".`, userContext.ID, result, experimentKey))
	} else {
		logger.Info(fmt.Sprintf(`User "%s" is not bucketed into any variation for experiment "%s": %s.`, userContext.ID, experimentKey, experimentDecision.Reason))
	}

	return decisionContext, experimentDecision, err
}

// OnTrack registers a handler for Track notifications
func (o *OptimizelyClient) OnTrack(callback func(notification notification.TrackNotification, userEvent event.UserEvent)) (int, error) {
	if o.NotificationCenter == nil {
		return 0, fmt.Errorf("no notification center found")
	}
	handler := func(payload interface{}) {
		success := false
		if trackPayload, ok := payload.([]interface{}); ok {
			if parsedNotification, ok := trackPayload[0].(notification.TrackNotification); ok {
				if parsedEvent, ok := trackPayload[1].(event.UserEvent); ok {
					success = true
					callback(parsedNotification, parsedEvent)
				}
			}
		}
		if !success {
			logger.Warning(fmt.Sprintf("Unable to convert notification payload %v into TrackNotification", payload))
		}
	}
	id, err := o.NotificationCenter.AddHandler(notification.Track, handler)
	if err != nil {
		logger.Warning("Problem with adding notification handler")
		return 0, err
	}
	return id, nil
}

// RemoveOnTrack removes handler for Track notification with given id
func (o *OptimizelyClient) RemoveOnTrack(id int) error {
	if o.NotificationCenter == nil {
		return fmt.Errorf("no notification center found")
	}
	if err := o.NotificationCenter.RemoveHandler(id, notification.Track); err != nil {
		logger.Warning("Problem with removing notification handler")
		return err
	}
	return nil
}

// GetProjectConfig returns the current ProjectConfig or nil if the instance is not valid.
func (o *OptimizelyClient) GetProjectConfig() (projectConfig pkg.ProjectConfig, err error) {

	projectConfig, err = o.ConfigManager.GetConfig()
	if err != nil {
		return nil, err
	}

	return projectConfig, nil
}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components.
func (o *OptimizelyClient) Close() {
	o.executionCtx.TerminateAndWait()
}
