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
	if !o.isValid {
		errorMessage := "Optimizely instance is not valid. Failing IsFeatureEnabled."
		err := errors.New(errorMessage)
		logger.Error(errorMessage, nil)
		return false, err
	}

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	projectConfig := o.configManager.GetConfig()

	if reflect.ValueOf(projectConfig).IsNil() {
		return false, fmt.Errorf("project config is null")
	}
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
	if !o.isValid {
		errorMessage := "Optimizely instance is not valid. Failing GetEnabledFeatures."
		err := errors.New(errorMessage)
		logger.Error(errorMessage, nil)
		return enabledFeatures, err
	}

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	projectConfig := o.configManager.GetConfig()

	if reflect.ValueOf(projectConfig).IsNil() {
		return enabledFeatures, fmt.Errorf("project config is null")
	}

	featureList := projectConfig.GetFeatureList()
	for _, feature := range featureList {
		isEnabled, _ := o.IsFeatureEnabled(feature.Key, userContext)

		if isEnabled {
			enabledFeatures = append(enabledFeatures, feature.Key)
		}
	}

	return enabledFeatures, nil
}

// GetFeatureVariableBoolean returns bool feature variable value
func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey string, variableKey string, userContext entities.UserContext) (value bool, err error) {
	val1, err1 := o.getFeatureVariable(false, featureKey, variableKey, userContext)
	if err1 == nil {
		switch val1.(type) {
		case bool:
			return val1.(bool), err1
		default:
			break
		}
	}
	return false, err1
}

// GetFeatureVariableDouble returns double feature variable value
func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey string, variableKey string, userContext entities.UserContext) (value float64, err error) {
	val1, err1 := o.getFeatureVariable(float64(0), featureKey, variableKey, userContext)
	if err1 == nil {
		switch val1.(type) {
		case float64:
			return val1.(float64), err1
		default:
			break
		}
	}
	return 0, err1
}

// GetFeatureVariableInteger returns integer feature variable value
func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey string, variableKey string, userContext entities.UserContext) (value int, err error) {
	val1, err1 := o.getFeatureVariable(int(0), featureKey, variableKey, userContext)
	if err1 == nil {
		switch val1.(type) {
		case int:
			return val1.(int), err1
		default:
			break
		}
	}
	return 0, err1
}

// GetFeatureVariableString returns string feature variable value
func (o *OptimizelyClient) GetFeatureVariableString(featureKey string, variableKey string, userContext entities.UserContext) (value string, err error) {
	val1, err1 := o.getFeatureVariable("", featureKey, variableKey, userContext)
	if err1 == nil {
		switch val1.(type) {
		case string:
			return val1.(string), err1
		default:
			break
		}
	}
	return "", err1
}

func (o *OptimizelyClient) getFeatureVariable(valueType interface{}, featureKey string, variableKey string, userContext entities.UserContext) (value interface{}, err error) {
	if !o.isValid {
		errorMessage := "Optimizely instance is not valid. Failing getFeatureVariable."
		err := errors.New(errorMessage)
		logger.Error(errorMessage, nil)
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
		}
	}()

	projectConfig := o.configManager.GetConfig()

	if reflect.ValueOf(projectConfig).IsNil() {
		return nil, fmt.Errorf("project config is null")
	}

	featureFlag, err := projectConfig.GetFeatureFlagByKey(featureKey)
	if err != nil {
		logger.Error("Error retrieving feature flag", err)
		return nil, err
	}
	variable, err := featureFlag.GetVariable(variableKey)
	if err != nil {
		logger.Error("Error retrieving variable", err)
		return nil, err
	}

	var featureValue interface{} = variable.DefaultValue

	feature, err := projectConfig.GetFeatureByKey(featureKey)
	if err != nil {
		logger.Error("Error retrieving feature", err)
		return nil, err
	}
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

	var typeName = ""
	var valueParsed interface{}
	switch valueType.(type) {
	case string:
		typeName = "string"
		convertedValue, ok := featureValue.(string)
		if ok {
			valueParsed = convertedValue
		}
	case int:
		typeName = "integer"
		convertedValue, ok := featureValue.(int)
		if ok {
			valueParsed = convertedValue
		}
	case float64:
		typeName = "double"
		convertedValue, ok := featureValue.(float64)
		if ok {
			valueParsed = convertedValue
		}
	case bool:
		typeName = "boolean"
		convertedValue, ok := featureValue.(bool)
		if ok {
			valueParsed = convertedValue
		}
	default:
		break
	}

	if valueParsed == nil || variable.Type != typeName {
		return nil, fmt.Errorf("Variable value for key %s is invalid or wrong type", variableKey)
	}

	// @TODO(yasir): send decision notification
	return valueParsed, nil
}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components
func (o *OptimizelyClient) Close() {
	o.cancelFunc()
}
