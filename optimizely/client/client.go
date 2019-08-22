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

package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"

	"github.com/optimizely/go-sdk/optimizely/event"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

var logger = logging.GetLogger("Client")

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	configManager   optimizely.ProjectConfigManager
	decisionService decision.DecisionService
	eventProcessor  event.Processor
	isValid         bool

	cancelFunc context.CancelFunc
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
func (o *OptimizelyClient) IsFeatureEnabled(featureKey string, userContext entities.UserContext) (result bool, err error) {

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	if isValid, err := o.isValidClient("IsFeatureEnabled"); !isValid {
		return false, err
	}

	projectConfig := o.configManager.GetConfig()
	feature, err := projectConfig.GetFeatureByKey(featureKey)
	if err != nil {
		logger.Error("Error retrieving feature", err)
		return result, err
	}
	featureDecisionContext := decision.FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: projectConfig,
	}

	userID := userContext.ID
	logger.Debug(fmt.Sprintf(`Evaluating feature "%s" for user "%s".`, featureKey, userID))
	featureDecision, err := o.decisionService.GetFeatureDecision(featureDecisionContext, userContext)

	if err != nil {
		logger.Error("Received an error while computing feature decision", err)
		return result, err
	}

	result = (featureDecision.Variation.FeatureEnabled == true)
	if result {
		logger.Info(fmt.Sprintf(`Feature "%s" is enabled for user "%s".`, featureKey, userID))
	} else {
		logger.Info(fmt.Sprintf(`Feature "%s" is not enabled for user "%s".`, featureKey, userID))
	}

	if featureDecision.Source == decision.FeatureTest {
		// send impression event for feature tests
		impressionEvent := event.CreateImpressionUserEvent(projectConfig, featureDecision.Experiment, featureDecision.Variation, userContext)
		o.eventProcessor.ProcessEvent(impressionEvent)
	}
	return result, nil
}

// GetEnabledFeatures returns an array containing the keys of all features in the project that are enabled for the given user.
func (o *OptimizelyClient) GetEnabledFeatures(userContext entities.UserContext) (enabledFeatures []string, err error) {

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	if isValid, err := o.isValidClient("GetEnabledFeatures"); !isValid {
		return enabledFeatures, err
	}

	projectConfig := o.configManager.GetConfig()
	featureList := projectConfig.GetFeatureList()
	for _, feature := range featureList {
		isEnabled, _ := o.IsFeatureEnabled(feature.Key, userContext)

		if isEnabled {
			enabledFeatures = append(enabledFeatures, feature.Key)
		}
	}

	return enabledFeatures, nil
}

// GetFeatureVariableString returns string feature variable value
func (o *OptimizelyClient) GetFeatureVariableString(featureKey string, variableKey string, userContext entities.UserContext) (value string, err error) {
	val, err := o.getFeatureVariable("string", featureKey, variableKey, userContext)
	if returnValue, ok := val.(string); ok {
		value = returnValue
	}
	return value, err
}

func (o *OptimizelyClient) getFeatureVariable(valueType string, featureKey string, variableKey string, userContext entities.UserContext) (value interface{}, err error) {

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	if isValid, err := o.isValidClient("getFeatureVariable"); !isValid {
		return nil, err
	}

	projectConfig := o.configManager.GetConfig()

	feature, err := projectConfig.GetFeatureByKey(featureKey)
	if err != nil {
		logger.Error("Error retrieving feature", err)
		return nil, err
	}

	variable, err := projectConfig.GetVariableByKey(featureKey, variableKey)
	if err != nil {
		logger.Error("Error retrieving variable", err)
		return nil, err
	}

	var featureValue = variable.DefaultValue

	featureDecisionContext := decision.FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: projectConfig,
	}

	featureDecision, err := o.decisionService.GetFeatureDecision(featureDecisionContext, userContext)
	if err == nil {
		for _, v := range featureDecision.Variation.Variables {
			if v.ID == variable.ID && featureDecision.Variation.FeatureEnabled {
				featureValue = v.Value
				break
			}
		}
	}

	var valueParsed interface{}
	switch valueType {
	case "string":
		valueParsed = featureValue
		break
	case "integer":
		convertedValue, err := strconv.Atoi(featureValue)
		if err == nil {
			valueParsed = convertedValue
		}
		break
	case "double":
		convertedValue, err := strconv.ParseFloat(featureValue, 64)
		if err == nil {
			valueParsed = convertedValue
		}
		break
	case "boolean":
		convertedValue, err := strconv.ParseBool(featureValue)
		if err == nil {
			valueParsed = convertedValue
		}
		break
	default:
		break
	}

	if valueParsed == nil || variable.Type != valueType {
		return nil, fmt.Errorf("Variable value for key %s is invalid or wrong type", variableKey)
	}

	// @TODO(yasir): send decision notification
	return valueParsed, nil
}

func (o *OptimizelyClient) isValidClient(methodName string) (result bool, err error) {
	if !o.isValid {
		errorMessage := fmt.Sprintf("Optimizely instance is not valid. Failing %s.", methodName)
		err := errors.New(errorMessage)
		logger.Error(errorMessage, nil)
		return false, err
	}

	if reflect.ValueOf(o.configManager.GetConfig()).IsNil() {
		return false, fmt.Errorf("project config is null")
	}
	return true, nil
}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components
func (o *OptimizelyClient) Close() {
	o.cancelFunc()
}
