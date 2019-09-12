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
	isValid         bool

	executionCtx utils.ExecutionCtx
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
func (o *OptimizelyClient) IsFeatureEnabled(featureKey string, userContext entities.UserContext) (result bool, err error) {

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
			errorMessage := fmt.Sprintf(`optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
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

// Track take and event key with event tags and if the event is part of the config, send to events backend.
func (o *OptimizelyClient) Track(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}) (err error) {

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
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

// GetFeatureVariableBoolean returns boolean feature variable value
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

// GetFeatureVariableDouble returns double feature variable value
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

// GetFeatureVariableInteger returns integer feature variable value
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

// GetFeatureVariableString returns string feature variable value
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

// GetFeatureVariable returns feature as a string along with it's associated type
func (o *OptimizelyClient) GetFeatureVariable(featureKey, variableKey string, userContext entities.UserContext) (value string, valueType entities.VariableType, err error) {

	context, featureDecision, err := o.getFeatureDecision(featureKey, userContext)
	if err != nil {
		return "", "", errors.New("error fetching project config")
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

// GetFeatureVariableMap returns variation map based on the decision service
func (o *OptimizelyClient) GetFeatureVariableMap(featureKey string, userContext entities.UserContext) (enabled bool, variableMap map[string]string, err error) {
	variableMap = make(map[string]string)
	variableIDMap := make(map[string]string)
	context, featureDecision, err := o.getFeatureDecision(featureKey, userContext)
	if err != nil {
		logger.Error("Optimizely SDK tracking error", err)
		return enabled, variableMap, err
	}

	feature := context.Feature
	for _, v := range feature.Variables {
		variableMap[v.Key] = v.DefaultValue
		variableIDMap[v.ID] = v.Key
	}

	enabled = featureDecision.Variation.FeatureEnabled

	if featureDecision.Variation == nil || !featureDecision.Variation.FeatureEnabled {
		return enabled, variableMap, err
	}

	for _, v := range featureDecision.Variation.Variables {
		if key, ok := variableIDMap[v.ID]; ok {
			variableMap[key] = v.Value
		}
	}

	return enabled, variableMap, err
}

func (o *OptimizelyClient) getFeatureDecision(featureKey string, userContext entities.UserContext) (decisionContext decision.FeatureDecisionContext, featureDecision decision.FeatureDecision, err error) {

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			logger.Error(errorMessage, err)

			// If we have a feature, then we can recover w/o throwing
			if decisionContext.Feature == nil {
				err = errors.New(errorMessage)
			}
		}
	}()

	userID := userContext.ID
	logger.Debug(fmt.Sprintf(`Evaluating feature "%s" for user "%s".`, featureKey, userID))

	projectConfig, err := o.GetProjectConfig()
	if err != nil {
		logger.Error("Error calling getFeatureDecision", err)
		return decisionContext, featureDecision, err
	}

	feature, err := projectConfig.GetFeatureByKey(featureKey)
	if err != nil {
		logger.Error("Error calling getFeatureDecision", err)
		return decisionContext, featureDecision, err
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

	// @TODO(yasir): send decision notification
	return decisionContext, featureDecision, err
}

// GetProjectConfig returns the current ProjectConfig or nil if the instance is not valid
func (o *OptimizelyClient) GetProjectConfig() (projectConfig optimizely.ProjectConfig, err error) {
	if !o.isValid {
		return nil, fmt.Errorf("optimizely instance is not valid")
	}

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
