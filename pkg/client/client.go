/****************************************************************************
 * Copyright 2019-2024, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"

	"github.com/hashicorp/go-multierror"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	pkgReasons "github.com/optimizely/go-sdk/v2/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
	"github.com/optimizely/go-sdk/v2/pkg/notification"
	"github.com/optimizely/go-sdk/v2/pkg/odp"
	pkgOdpSegment "github.com/optimizely/go-sdk/v2/pkg/odp/segment"
	pkgOdpUtils "github.com/optimizely/go-sdk/v2/pkg/odp/utils"
	"github.com/optimizely/go-sdk/v2/pkg/optimizelyjson"
	"github.com/optimizely/go-sdk/v2/pkg/tracing"
	"github.com/optimizely/go-sdk/v2/pkg/utils"
)

const (
	// DefaultTracerName is the name of the tracer used by the Optimizely SDK
	DefaultTracerName = "OptimizelySDK"
	// SpanNameDecide is the name of the span used by the Optimizely SDK for tracing decide call
	SpanNameDecide = "decide"
	// SpanNameDecideForKeys is the name of the span used by the Optimizely SDK for tracing decideForKeys call
	SpanNameDecideForKeys = "decideForKeys"
	// SpanNameDecideAll is the name of the span used by the Optimizely SDK for tracing decideAll call
	SpanNameDecideAll = "decideAll"
	// SpanNameActivate is the name of the span used by the Optimizely SDK for tracing Activate call
	SpanNameActivate = "Activate"
	// SpanNameFetchQualifiedSegments is the name of the span used by the Optimizely SDK for tracing fetchQualifiedSegments call
	SpanNameFetchQualifiedSegments = "fetchQualifiedSegments"
	// SpanNameSendOdpEvent is the name of the span used by the Optimizely SDK for tracing SendOdpEvent call
	SpanNameSendOdpEvent = "SendOdpEvent"
	// SpanNameIsFeatureEnabled is the name of the span used by the Optimizely SDK for tracing IsFeatureEnabled call
	SpanNameIsFeatureEnabled = "IsFeatureEnabled"
	// SpanNameGetEnabledFeatures is the name of the span used by the Optimizely SDK for tracing GetEnabledFeatures call
	SpanNameGetEnabledFeatures = "GetEnabledFeatures"
	// SpanNameGetFeatureVariableBoolean is the name of the span used by the Optimizely SDK for tracing GetFeatureVariableBoolean call
	SpanNameGetFeatureVariableBoolean = "GetFeatureVariableBoolean"
	// SpanNameGetFeatureVariableDouble is the name of the span used by the Optimizely SDK for tracing GetFeatureVariableDouble call
	SpanNameGetFeatureVariableDouble = "GetFeatureVariableDouble"
	// SpanNameGetFeatureVariableInteger is the name of the span used by the Optimizely SDK for tracing GetFeatureVariableInteger call
	SpanNameGetFeatureVariableInteger = "GetFeatureVariableInteger"
	// SpanNameGetFeatureVariableString is the name of the span used by the Optimizely SDK for tracing GetFeatureVariableString call
	SpanNameGetFeatureVariableString = "GetFeatureVariableString"
	// SpanNameGetFeatureVariableJSON is the name of the span used by the Optimizely SDK for tracing GetFeatureVariableJSON call
	SpanNameGetFeatureVariableJSON = "GetFeatureVariableJSON"
	// SpanNameGetFeatureVariablePrivate is the name of the span used by the Optimizely SDK for tracing getFeatureVariable call
	SpanNameGetFeatureVariablePrivate = "getFeatureVariable"
	// SpanNameGetFeatureVariablePublic is the name of the span used by the Optimizely SDK for tracing GetFeatureVariable call
	SpanNameGetFeatureVariablePublic = "GetFeatureVariable"
	// SpanNameGetAllFeatureVariablesWithDecision is the name of the span used by the Optimizely SDK for tracing GetAllFeatureVariablesWithDecision call
	SpanNameGetAllFeatureVariablesWithDecision = "GetAllFeatureVariablesWithDecision"
	// SpanNameGetDetailedFeatureDecisionUnsafe is the name of the span used by the Optimizely SDK for tracing GetDetailedFeatureDecisionUnsafe call
	SpanNameGetDetailedFeatureDecisionUnsafe = "GetDetailedFeatureDecisionUnsafe"
	// SpanNameGetAllFeatureVariables is the name of the span used by the Optimizely SDK for tracing GetAllFeatureVariables call
	SpanNameGetAllFeatureVariables = "GetAllFeatureVariables"
	// SpanNameGetVariation is the name of the span used by the Optimizely SDK for tracing GetVariation call
	SpanNameGetVariation = "GetVariation"
	// SpanNameTrack is the name of the span used by the Optimizely SDK for tracing Track call
	SpanNameTrack = "Track"
	// SpanNameGetFeatureDecision is the name of the span used by the Optimizely SDK for tracing getFeatureDecision call
	SpanNameGetFeatureDecision = "getFeatureDecision"
	// SpanNameGetExperimentDecision is the name of the span used by the Optimizely SDK for tracing getExperimentDecision call
	SpanNameGetExperimentDecision = "getExperimentDecision"
	// SpanNameGetProjectConfig is the name of the span used by the Optimizely SDK for tracing getProjectConfig call
	SpanNameGetProjectConfig = "getProjectConfig"
	// SpanNameGetOptimizelyConfig is the name of the span used by the Optimizely SDK for tracing GetOptimizelyConfig call
	SpanNameGetOptimizelyConfig = "GetOptimizelyConfig"
	// SpanNameGetDecisionVariableMap is the name of the span used by the Optimizely SDK for tracing getDecisionVariableMap call
	SpanNameGetDecisionVariableMap = "getDecisionVariableMap"
)

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	ctx                  context.Context
	ConfigManager        config.ProjectConfigManager
	DecisionService      decision.Service
	EventProcessor       event.Processor
	OdpManager           odp.Manager
	notificationCenter   notification.Center
	execGroup            *utils.ExecGroup
	logger               logging.OptimizelyLogProducer
	defaultDecideOptions *decide.Options
	tracer               tracing.Tracer
}

// CreateUserContext creates a context of the user for which decision APIs will be called.
// A user context will be created successfully even when the SDK is not fully configured yet.
func (o *OptimizelyClient) CreateUserContext(userID string, attributes map[string]interface{}) OptimizelyUserContext {
	if o.OdpManager != nil {
		// Identify user to odp server
		o.OdpManager.IdentifyUser(userID)
	}
	// Passing qualified segments as nil initially since they will be fetched later
	return newOptimizelyUserContext(o, userID, attributes, nil, nil)
}

// WithTraceContext sets the context for the OptimizelyClient which can be used to propagate trace information
func (o *OptimizelyClient) WithTraceContext(ctx context.Context) *OptimizelyClient {
	o.ctx = ctx
	return o
}

func (o *OptimizelyClient) decide(userContext OptimizelyUserContext, key string, options *decide.Options) OptimizelyDecision {
	var err error
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
			errorMessage := "decide call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameDecide)
	defer span.End()

	decisionContext := decision.FeatureDecisionContext{
		ForcedDecisionService: userContext.forcedDecisionService,
	}
	projectConfig, err := o.getProjectConfig()
	if err != nil {
		return NewErrorDecision(key, userContext, decide.GetDecideError(decide.SDKNotReady))
	}
	decisionContext.ProjectConfig = projectConfig

	feature, err := projectConfig.GetFeatureByKey(key)
	if err != nil {
		return NewErrorDecision(key, userContext, decide.GetDecideError(decide.FlagKeyInvalid, key))
	}
	decisionContext.Feature = &feature

	usrContext := entities.UserContext{
		ID:                userContext.GetUserID(),
		Attributes:        userContext.GetUserAttributes(),
		QualifiedSegments: userContext.GetQualifiedSegments(),
	}
	var variationKey string
	var eventSent, flagEnabled bool
	allOptions := o.getAllOptions(options)
	decisionReasons := decide.NewDecisionReasons(&allOptions)
	decisionContext.Variable = entities.Variable{}
	var featureDecision decision.FeatureDecision
	var reasons decide.DecisionReasons

	// To avoid cyclo-complexity warning
	findRegularDecision := func() {
		// regular decision
		featureDecision, reasons, err = o.DecisionService.GetFeatureDecision(decisionContext, usrContext, &allOptions)
		decisionReasons.Append(reasons)
	}

	// check forced-decisions first
	// Passing empty rule-key because checking mapping with flagKey only
	if userContext.forcedDecisionService != nil {
		var variation *entities.Variation
		variation, reasons, err = userContext.forcedDecisionService.FindValidatedForcedDecision(projectConfig, decision.OptimizelyDecisionContext{FlagKey: key, RuleKey: ""}, &allOptions)
		decisionReasons.Append(reasons)
		if err != nil {
			findRegularDecision()
		} else {
			featureDecision = decision.FeatureDecision{Decision: decision.Decision{Reason: pkgReasons.ForcedDecisionFound}, Variation: variation, Source: decision.FeatureTest}
		}
	} else {
		findRegularDecision()
	}

	if err != nil {
		o.logger.Warning(fmt.Sprintf(`Received error while making a decision for feature %q: %s`, key, err))
	}

	if featureDecision.Variation != nil {
		variationKey = featureDecision.Variation.Key
		flagEnabled = featureDecision.Variation.FeatureEnabled
	}

	if !allOptions.DisableDecisionEvent {
		if ue, ok := event.CreateImpressionUserEvent(decisionContext.ProjectConfig, featureDecision.Experiment,
			featureDecision.Variation, usrContext, key, featureDecision.Experiment.Key, featureDecision.Source, flagEnabled); ok {
			o.EventProcessor.ProcessEvent(ue)
			eventSent = true
		}
	}

	variableMap := map[string]interface{}{}
	if !allOptions.ExcludeVariables {
		variableMap, reasons = o.getDecisionVariableMap(feature, featureDecision.Variation, flagEnabled)
		decisionReasons.Append(reasons)
	}
	optimizelyJSON := optimizelyjson.NewOptimizelyJSONfromMap(variableMap)
	reasonsToReport := decisionReasons.ToReport()
	ruleKey := featureDecision.Experiment.Key

	if o.notificationCenter != nil {
		decisionNotification := decision.FlagNotification(key, variationKey, ruleKey, flagEnabled, eventSent, usrContext, variableMap, reasonsToReport)
		o.logger.Debug(fmt.Sprintf(`Feature %q is enabled for user %q? %v`, key, usrContext.ID, flagEnabled))
		if e := o.notificationCenter.Send(notification.Decision, *decisionNotification); e != nil {
			o.logger.Warning("Problem with sending notification")
		}
	}

	decision := NewOptimizelyDecision(variationKey, ruleKey, key, flagEnabled, optimizelyJSON, userContext, reasonsToReport)
	decision.IsEveryoneElseVariation = featureDecision.Experiment.IsEveryoneElseVariation
	return decision
}

func (o *OptimizelyClient) decideForKeys(userContext OptimizelyUserContext, keys []string, options *decide.Options) map[string]OptimizelyDecision {
	var err error
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
			errorMessage := "decideForKeys call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameDecideForKeys)
	defer span.End()

	decisionMap := map[string]OptimizelyDecision{}
	if _, err = o.getProjectConfig(); err != nil {
		o.logger.Error("Optimizely instance is not valid, failing decideForKeys call.", err)
		return decisionMap
	}

	if len(keys) == 0 {
		return decisionMap
	}

	enabledFlagsOnly := o.getAllOptions(options).EnabledFlagsOnly
	for _, key := range keys {
		optimizelyDecision := o.decide(userContext, key, options)
		if !enabledFlagsOnly || optimizelyDecision.Enabled {
			decisionMap[key] = optimizelyDecision
		}
	}

	return decisionMap
}

func (o *OptimizelyClient) decideAll(userContext OptimizelyUserContext, options *decide.Options) map[string]OptimizelyDecision {

	var err error
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
			errorMessage := "decideAll call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameDecideAll)
	defer span.End()

	projectConfig, err := o.getProjectConfig()
	if err != nil {
		o.logger.Error("Optimizely instance is not valid, failing decideAll call.", err)
		return map[string]OptimizelyDecision{}
	}

	allFlagKeys := []string{}
	for _, flag := range projectConfig.GetFeatureList() {
		allFlagKeys = append(allFlagKeys, flag.Key)
	}

	return o.decideForKeys(userContext, allFlagKeys, options)
}

// fetchQualifiedSegments fetches all qualified segments for the user context.
// request is performed asynchronously only when callback is provided
func (o *OptimizelyClient) fetchQualifiedSegments(userContext *OptimizelyUserContext, options []pkgOdpSegment.OptimizelySegmentOption, callback func(success bool)) {
	var err error
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
			o.logger.Error("fetchQualifiedSegments call, optimizely SDK is panicking with the error:", err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameFetchQualifiedSegments)
	defer span.End()

	// on failure, qualifiedSegments should be reset if a previous value exists.
	userContext.SetQualifiedSegments(nil)

	if _, err = o.getProjectConfig(); err != nil {
		o.logger.Error("fetchQualifiedSegments failed with error:", decide.GetDecideError(decide.SDKNotReady))
		if callback != nil {
			callback(false)
		}
		return
	}

	qualifiedSegments, segmentsError := o.OdpManager.FetchQualifiedSegments(userContext.GetUserID(), options)
	success := segmentsError == nil

	if success {
		userContext.SetQualifiedSegments(qualifiedSegments)
	} else {
		o.logger.Error("fetchQualifiedSegments failed with error:", segmentsError)
	}

	if callback != nil {
		callback(success)
	}
}

// SendOdpEvent sends an event to the ODP server.
func (o *OptimizelyClient) SendOdpEvent(eventType, action string, identifiers map[string]string, data map[string]interface{}) (err error) {

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
			errorMessage := "SendOdpEvent call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameSendOdpEvent)
	defer span.End()

	if _, err = o.getProjectConfig(); err != nil {
		o.logger.Error("SendOdpEvent failed with error:", decide.GetDecideError(decide.SDKNotReady))
		return err
	}

	// the event type (default = "fullstack").
	if eventType == "" {
		eventType = pkgOdpUtils.OdpEventType
	}

	if len(identifiers) == 0 {
		err = errors.New("ODP events must have at least one key-value pair in identifiers")
		o.logger.Error("received an error while sending ODP event", err)
		return err
	}

	return o.OdpManager.SendOdpEvent(eventType, action, identifiers, data)
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
			errorMessage := "Activate call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameActivate)
	defer span.End()

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
			errorMessage := "IsFeatureEnabled call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameIsFeatureEnabled)
	defer span.End()

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
		o.logger.Debug(fmt.Sprintf(`Feature %q is enabled for user %q.`, featureKey, userContext.ID))
	} else {
		o.logger.Debug(fmt.Sprintf(`Feature %q is not enabled for user %q.`, featureKey, userContext.ID))
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
			errorMessage := "GetEnabledFeatures call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetEnabledFeatures)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariableBoolean)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariableDouble)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariableInteger)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariableString)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariableJSON)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariablePrivate)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureVariablePublic)
	defer span.End()

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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetAllFeatureVariablesWithDecision)
	defer span.End()

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
		o.logger.Warning(fmt.Sprintf(`feature %q does not exist`, featureKey))
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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetDetailedFeatureDecisionUnsafe)
	defer span.End()

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
		o.logger.Warning(fmt.Sprintf(`feature %q does not exist`, featureKey))
		return decisionInfo, nil
	}

	errs := new(multierror.Error)

	for _, v := range feature.VariableMap {
		val := v.DefaultValue

		if decisionInfo.Enabled {
			if variable, ok := featureDecision.Variation.Variables[v.ID]; ok {
				val = variable.Value
			} else {
				o.logger.Warning(fmt.Sprintf(`variable with id %q does not exist`, v.ID))
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
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetAllFeatureVariables)
	defer span.End()

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
			errorMessage := "GetVariation call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetVariation)
	defer span.End()

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
			errorMessage := "Track call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameTrack)
	defer span.End()

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		o.logger.Error("Optimizely SDK tracking error", e)
		return e
	}

	configEvent, e := projectConfig.GetEventByKey(eventKey)

	if e != nil {
		errorMessage := fmt.Sprintf(`Unable to get event for key %q: %s`, eventKey, e)
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
			errorMessage := "getFeatureDecision call, optimizely SDK is panicking with the error:"
			o.logger.Error(errorMessage, err)
			o.logger.Debug(string(debug.Stack()))
		}
	}()

	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetFeatureDecision)
	defer span.End()

	userID := userContext.ID
	o.logger.Debug(fmt.Sprintf(`Evaluating feature %q for user %q.`, featureKey, userID))

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		o.logger.Error("Error calling getFeatureDecision", e)
		return decisionContext, featureDecision, e
	}

	decisionContext.ProjectConfig = projectConfig
	feature, e := projectConfig.GetFeatureByKey(featureKey)
	if e != nil {
		o.logger.Warning(fmt.Sprintf(`Could not get feature for key %q: %s`, featureKey, e))
		return decisionContext, featureDecision, nil
	}

	decisionContext.Feature = &feature
	variable := entities.Variable{}
	if variableKey != "" {
		variable, err = projectConfig.GetVariableByKey(feature.Key, variableKey)
		if err != nil {
			o.logger.Warning(fmt.Sprintf(`Could not get variable for key %q: %s`, variableKey, err))
			return decisionContext, featureDecision, nil
		}
	}

	decisionContext.Variable = variable
	options := &decide.Options{}
	featureDecision, _, err = o.DecisionService.GetFeatureDecision(decisionContext, userContext, options)
	if err != nil {
		o.logger.Warning(fmt.Sprintf(`Received error while making a decision for feature %q: %s`, featureKey, err))
		return decisionContext, featureDecision, nil
	}

	return decisionContext, featureDecision, nil
}

func (o *OptimizelyClient) getExperimentDecision(experimentKey string, userContext entities.UserContext) (decisionContext decision.ExperimentDecisionContext, experimentDecision decision.ExperimentDecision, err error) {
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetExperimentDecision)
	defer span.End()

	userID := userContext.ID
	o.logger.Debug(fmt.Sprintf(`Evaluating experiment %q for user %q.`, experimentKey, userID))

	projectConfig, e := o.getProjectConfig()
	if e != nil {
		return decisionContext, experimentDecision, e
	}

	experiment, e := projectConfig.GetExperimentByKey(experimentKey)
	if e != nil {
		o.logger.Warning(fmt.Sprintf(`Could not get experiment for key %q: %s`, experimentKey, e))
		return decisionContext, experimentDecision, nil
	}

	decisionContext = decision.ExperimentDecisionContext{
		Experiment:    &experiment,
		ProjectConfig: projectConfig,
	}

	options := &decide.Options{}
	experimentDecision, _, err = o.DecisionService.GetExperimentDecision(decisionContext, userContext, options)
	if err != nil {
		o.logger.Warning(fmt.Sprintf(`Received error while making a decision for experiment %q: %s`, experimentKey, err))
		return decisionContext, experimentDecision, nil
	}

	if experimentDecision.Variation != nil {
		result := experimentDecision.Variation.Key
		o.logger.Debug(fmt.Sprintf(`User %q is bucketed into variation %q of experiment %q.`, userContext.ID, result, experimentKey))
	} else {
		o.logger.Debug(fmt.Sprintf(`User %q is not bucketed into any variation for experiment %q: %s.`, userContext.ID, experimentKey, experimentDecision.Reason))
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
		o.logger.Warning(fmt.Sprintf(`type %q is unknown, returning string`, variableType))
	}
	return convertedValue, err
}

func (o *OptimizelyClient) getProjectConfig() (projectConfig config.ProjectConfig, err error) {
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetProjectConfig)
	defer span.End()

	if isNil(o.ConfigManager) {
		return nil, errors.New("project config manager is not initialized")
	}
	projectConfig, err = o.ConfigManager.GetConfig()
	if err != nil {
		return nil, err
	}

	return projectConfig, nil
}

func (o *OptimizelyClient) getAllOptions(options *decide.Options) decide.Options {
	return decide.Options{
		DisableDecisionEvent:     o.defaultDecideOptions.DisableDecisionEvent || options.DisableDecisionEvent,
		EnabledFlagsOnly:         o.defaultDecideOptions.EnabledFlagsOnly || options.EnabledFlagsOnly,
		ExcludeVariables:         o.defaultDecideOptions.ExcludeVariables || options.ExcludeVariables,
		IgnoreUserProfileService: o.defaultDecideOptions.IgnoreUserProfileService || options.IgnoreUserProfileService,
		IncludeReasons:           o.defaultDecideOptions.IncludeReasons || options.IncludeReasons,
	}
}

// GetOptimizelyConfig returns OptimizelyConfig object
func (o *OptimizelyClient) GetOptimizelyConfig() (optimizelyConfig *config.OptimizelyConfig) {
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetOptimizelyConfig)
	defer span.End()
	return o.ConfigManager.GetOptimizelyConfig()
}

// GetNotificationCenter returns Optimizely Notification Center interface
func (o *OptimizelyClient) GetNotificationCenter() notification.Center {
	return o.notificationCenter
}

// Close closes the Optimizely instance and stops any ongoing tasks from its children components.
func (o *OptimizelyClient) Close() {
	o.execGroup.TerminateAndWait()
}

func (o *OptimizelyClient) getDecisionVariableMap(feature entities.Feature, variation *entities.Variation, featureEnabled bool) (map[string]interface{}, decide.DecisionReasons) {
	_, span := o.tracer.StartSpan(o.ctx, DefaultTracerName, SpanNameGetDecisionVariableMap)
	defer span.End()

	reasons := decide.NewDecisionReasons(nil)
	valuesMap := map[string]interface{}{}

	for _, v := range feature.VariableMap {
		val := v.DefaultValue

		if featureEnabled {
			if variable, ok := variation.Variables[v.ID]; ok {
				val = variable.Value
			}
		}

		typedValue, typedError := o.getTypedValue(val, v.Type)
		if typedError != nil {
			reasons.AddError(decide.GetDecideMessage(decide.VariableValueInvalid, v.Key))
		}
		valuesMap[v.Key] = typedValue
	}

	return valuesMap, reasons
}

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}
