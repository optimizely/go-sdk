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

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/utils"
)

var logger = logging.GetLogger("Client")

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	configManager   optimizely.ProjectConfigManager
	decisionService decision.Service
	eventProcessor  event.Processor

	executionCtx utils.ExecutionCtx
}

// Activate returns the key of the variation the user is bucketed into and sends an impression event to the Optimizely log endpoint
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
		o.eventProcessor.ProcessEvent(impressionEvent)
	}

	return result, err
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
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

	context, featureDecision, err := o.getFeatureDecision(featureKey, userContext)
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
		impressionEvent := event.CreateImpressionUserEvent(context.ProjectConfig, featureDecision.Experiment, *featureDecision.Variation, userContext)
		o.eventProcessor.ProcessEvent(impressionEvent)
	}
	return result, err
}

// GetEnabledFeatures returns an array containing the keys of all features in the project that are enabled for the given user.
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
		isEnabled, _ := o.IsFeatureEnabled(feature.Key, userContext)

		if isEnabled {
			enabledFeatures = append(enabledFeatures, feature.Key)
		}
	}

	return enabledFeatures, err
}

// GetFeatureVariableBoolean returns boolean feature variable value
func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey, variableKey string, userContext entities.UserContext) (value bool, err error) {

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

// GetFeatureVariableDouble returns double feature variable value
func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey, variableKey string, userContext entities.UserContext) (value float64, err error) {

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

// GetFeatureVariableInteger returns integer feature variable value
func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey, variableKey string, userContext entities.UserContext) (value int, err error) {

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

// GetFeatureVariableString returns string feature variable value
func (o *OptimizelyClient) GetFeatureVariableString(featureKey, variableKey string, userContext entities.UserContext) (value string, err error) {

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

	value, valueType, err := o.GetFeatureVariable(featureKey, variableKey, userContext)
	if err != nil {
		return "", err
	}
	if valueType != entities.String {
		return "", fmt.Errorf("variable value for key %s is wrong type", variableKey)
	}
	return value, err
}

// GetFeatureVariable returns feature as a string along with it's associated type
func (o *OptimizelyClient) GetFeatureVariable(featureKey, variableKey string, userContext entities.UserContext) (value string, valueType entities.VariableType, err error) {

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

	context, featureDecision, err := o.getFeatureDecision(featureKey, userContext)
	if err != nil {
		return "", "", err
	}

	variable, err := context.ProjectConfig.GetVariableByKey(featureKey, variableKey)
	if err != nil {
		return "", "", err
	}

	if featureDecision.Variation != nil {
		if v, ok := featureDecision.Variation.Variables[variable.ID]; ok && featureDecision.Variation.FeatureEnabled {
			return v.Value, variable.Type, err
		}
	}

	return variable.DefaultValue, variable.Type, err
}

// GetAllFeatureVariables returns all the variables for a given feature along with the enabled state
func (o *OptimizelyClient) GetAllFeatureVariables(featureKey string, userContext entities.UserContext) (enabled bool, variableMap map[string]string, err error) {

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

	variableMap = make(map[string]string)
	decisionContext, featureDecision, err := o.getFeatureDecision(featureKey, userContext)
	if err != nil {
		logger.Error("Optimizely SDK tracking error", err)
		return enabled, variableMap, err
	}

	feature := decisionContext.Feature
	if featureDecision.Variation != nil {
		enabled = featureDecision.Variation.FeatureEnabled
	}

	for _, v := range feature.Variables {
		variableMap[v.Key] = v.DefaultValue

		if enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				variableMap[v.Key] = variable.Value
			}
		}
	}

	return enabled, variableMap, err
}

// GetVariation returns the key of the variation the user is bucketed into
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

// Track take and event key with event tags and if the event is part of the config, send to events backend.
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

	projectConfig, err := o.GetProjectConfig()
	if err != nil {
		logger.Error("Optimizely SDK tracking error", err)
		return err
	}

	configEvent, err := projectConfig.GetEventByKey(eventKey)

	if err == nil {
		userEvent := event.CreateConversionUserEvent(projectConfig, configEvent, userContext, eventTags)
		o.eventProcessor.ProcessEvent(userEvent)
	} else {
		errorMessage := fmt.Sprintf(`optimizely SDK track: error getting event with key "%s"`, eventKey)
		logger.Error(errorMessage, err)
		return err
	}

	return err
}

func (o *OptimizelyClient) getFeatureDecision(featureKey string, userContext entities.UserContext) (decisionContext decision.FeatureDecisionContext, featureDecision decision.FeatureDecision, err error) {

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
		logger.Error("Error calling getFeatureDecision", e)
		return decisionContext, featureDecision, e
	}

	decisionContext = decision.FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: projectConfig,
	}

	featureDecision, err = o.decisionService.GetFeatureDecision(decisionContext, userContext)
	if err != nil {
		err = nil
		logger.Warning("error making a decision")
		return decisionContext, featureDecision, err
	}

	return decisionContext, featureDecision, err
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
		logger.Error("Error calling getExperimentDecision", e)
		return decisionContext, experimentDecision, e
	}

	decisionContext = decision.ExperimentDecisionContext{
		Experiment:    &experiment,
		ProjectConfig: projectConfig,
	}

	experimentDecision, err = o.decisionService.GetExperimentDecision(decisionContext, userContext)
	if err != nil {
		err = nil
		logger.Warning(fmt.Sprintf(`error making a decision for experiment "%s"`, experimentKey))
		return decisionContext, experimentDecision, err
	}

	if experimentDecision.Variation != nil {
		result := experimentDecision.Variation.Key
		logger.Info(fmt.Sprintf(`User "%s" is bucketed into variation "%s" of experiment "%s".`, userContext.ID, result, experimentKey))
	} else {
		logger.Info(fmt.Sprintf(`User "%s" is not bucketed into any variation for experiment "%s".`, userContext.ID, experimentKey))
	}

	return decisionContext, experimentDecision, err
}

// GetProjectConfig returns the current ProjectConfig or nil if the instance is not valid
func (o *OptimizelyClient) GetProjectConfig() (projectConfig optimizely.ProjectConfig, err error) {

	projectConfig, err = o.configManager.GetConfig()
	if err != nil {
		return nil, err
	}

	return projectConfig, nil
}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components
func (o *OptimizelyClient) Close() {
	o.executionCtx.TerminateAndWait()
}
