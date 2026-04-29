/****************************************************************************
 * Copyright 2026, Optimizely, Inc. and contributors                       *
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

package openfeature

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	of "github.com/open-feature/go-sdk/openfeature"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
)

// BooleanEvaluation evaluates a boolean feature flag by delegating to the
// Optimizely Decide API and returning the Enabled field.
func (p *Provider) BooleanEvaluation(_ context.Context, flag string, defaultValue bool, flatCtx of.FlattenedContext) of.BoolResolutionDetail {
	detail, err := p.evaluate(flag, flatCtx)
	if err != nil {
		return of.BoolResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}
	return of.BoolResolutionDetail{
		Value:                    detail.decision.Enabled,
		ProviderResolutionDetail: detail.toProviderDetail(),
	}
}

// StringEvaluation evaluates a string feature variable by extracting the
// variable specified by the variableKey evaluation context attribute.
func (p *Provider) StringEvaluation(_ context.Context, flag string, defaultValue string, flatCtx of.FlattenedContext) of.StringResolutionDetail {
	detail, ctxResult, err := p.evaluateVariable(flag, flatCtx)
	if err != nil {
		return of.StringResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	if !ctxResult.hasVariableKey {
		return of.StringResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: errorDetail(&evaluationError{
				code:    of.GeneralCode,
				message: "variableKey is required for string evaluation",
			}),
		}
	}

	val, err := getVariable(detail, ctxResult.variableKey)
	if err != nil {
		return of.StringResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	strVal, ok := val.(string)
	if !ok {
		strVal = fmt.Sprintf("%v", val)
	}

	return of.StringResolutionDetail{
		Value:                    strVal,
		ProviderResolutionDetail: detail.toProviderDetail(),
	}
}

// IntEvaluation evaluates an integer feature variable.
func (p *Provider) IntEvaluation(_ context.Context, flag string, defaultValue int64, flatCtx of.FlattenedContext) of.IntResolutionDetail {
	detail, ctxResult, err := p.evaluateVariable(flag, flatCtx)
	if err != nil {
		return of.IntResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	if !ctxResult.hasVariableKey {
		return of.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: errorDetail(&evaluationError{
				code:    of.GeneralCode,
				message: "variableKey is required for integer evaluation",
			}),
		}
	}

	val, err := getVariable(detail, ctxResult.variableKey)
	if err != nil {
		return of.IntResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	intVal, parseErr := toInt64(val)
	if parseErr != nil {
		return of.IntResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: errorDetail(&evaluationError{
				code:    of.ParseErrorCode,
				message: fmt.Sprintf("cannot parse '%v' as int64: %v", val, parseErr),
			}),
		}
	}

	return of.IntResolutionDetail{
		Value:                    intVal,
		ProviderResolutionDetail: detail.toProviderDetail(),
	}
}

// FloatEvaluation evaluates a float feature variable.
func (p *Provider) FloatEvaluation(_ context.Context, flag string, defaultValue float64, flatCtx of.FlattenedContext) of.FloatResolutionDetail {
	detail, ctxResult, err := p.evaluateVariable(flag, flatCtx)
	if err != nil {
		return of.FloatResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	if !ctxResult.hasVariableKey {
		return of.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: errorDetail(&evaluationError{
				code:    of.GeneralCode,
				message: "variableKey is required for float evaluation",
			}),
		}
	}

	val, err := getVariable(detail, ctxResult.variableKey)
	if err != nil {
		return of.FloatResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	floatVal, parseErr := toFloat64(val)
	if parseErr != nil {
		return of.FloatResolutionDetail{
			Value: defaultValue,
			ProviderResolutionDetail: errorDetail(&evaluationError{
				code:    of.ParseErrorCode,
				message: fmt.Sprintf("cannot parse '%v' as float64: %v", val, parseErr),
			}),
		}
	}

	return of.FloatResolutionDetail{
		Value:                    floatVal,
		ProviderResolutionDetail: detail.toProviderDetail(),
	}
}

// ObjectEvaluation evaluates an object (JSON) feature variable. If variableKey
// is provided, returns that specific variable parsed as JSON. If omitted,
// returns the full variables map.
func (p *Provider) ObjectEvaluation(_ context.Context, flag string, defaultValue interface{}, flatCtx of.FlattenedContext) of.InterfaceResolutionDetail {
	detail, ctxResult, err := p.evaluateVariable(flag, flatCtx)
	if err != nil {
		return of.InterfaceResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	// If no variableKey, return the full variables map
	if !ctxResult.hasVariableKey {
		varsMap := detail.decision.Variables.ToMap()
		return of.InterfaceResolutionDetail{
			Value:                    varsMap,
			ProviderResolutionDetail: detail.toProviderDetail(),
		}
	}

	val, err := getVariable(detail, ctxResult.variableKey)
	if err != nil {
		return of.InterfaceResolutionDetail{
			Value:                    defaultValue,
			ProviderResolutionDetail: errorDetail(err),
		}
	}

	// If it's a string, try to parse as JSON
	if strVal, ok := val.(string); ok {
		var parsed interface{}
		if jsonErr := json.Unmarshal([]byte(strVal), &parsed); jsonErr == nil {
			val = parsed
		}
	}

	return of.InterfaceResolutionDetail{
		Value:                    val,
		ProviderResolutionDetail: detail.toProviderDetail(),
	}
}

// evaluationDetail holds the result of an Optimizely Decide call.
type evaluationDetail struct {
	decision decisionResult
}

// decisionResult wraps the fields we need from OptimizelyDecision.
type decisionResult struct {
	VariationKey string
	Enabled      bool
	Variables    variablesAccessor
	FlagKey      string
	RuleKey      string
	Reasons      []string
}

// variablesAccessor abstracts access to decision variables.
type variablesAccessor interface {
	ToMap() map[string]interface{}
}

func (d *evaluationDetail) toProviderDetail() of.ProviderResolutionDetail {
	metadata := of.FlagMetadata{
		"flagKey": d.decision.FlagKey,
	}
	if d.decision.RuleKey != "" {
		metadata["ruleKey"] = d.decision.RuleKey
	}
	if len(d.decision.Reasons) > 0 {
		metadata["reasons"] = d.decision.Reasons
	}

	return of.ProviderResolutionDetail{
		Reason:       mapReason(d.decision.VariationKey, d.decision.Enabled, false),
		Variant:      d.decision.VariationKey,
		FlagMetadata: metadata,
	}
}

// evaluationError wraps an error code and message for evaluation failures.
type evaluationError struct {
	code    of.ErrorCode
	message string
}

func (e *evaluationError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// errorDetail builds a ProviderResolutionDetail for an error case.
func errorDetail(err error) of.ProviderResolutionDetail {
	if evalErr, ok := err.(*evaluationError); ok {
		return of.ProviderResolutionDetail{
			ResolutionError: makeResolutionError(evalErr.code, evalErr.message),
			Reason:          of.ErrorReason,
		}
	}
	if ctxErr, ok := err.(*contextError); ok {
		return of.ProviderResolutionDetail{
			ResolutionError: makeResolutionError(ctxErr.code, ctxErr.message),
			Reason:          of.ErrorReason,
		}
	}
	return of.ProviderResolutionDetail{
		ResolutionError: of.NewGeneralResolutionError(err.Error()),
		Reason:          of.ErrorReason,
	}
}

// evaluate performs the common evaluation logic: check readiness, extract
// context, call Decide, and detect errors.
func (p *Provider) evaluate(flag string, flatCtx of.FlattenedContext) (*evaluationDetail, error) {
	_, detail, err := p.evaluateWithContext(flag, flatCtx)
	return detail, err
}

// evaluateVariable performs evaluate and also returns the parsed context
// (needed for variableKey extraction).
func (p *Provider) evaluateVariable(flag string, flatCtx of.FlattenedContext) (*evaluationDetail, *contextResult, error) {
	ctxResult, detail, err := p.evaluateWithContext(flag, flatCtx)
	return detail, ctxResult, err
}

// evaluateWithContext is the shared evaluation core. It checks readiness,
// extracts the OpenFeature context, calls the Optimizely Decide API, and
// detects flag-not-found conditions. Both evaluate and evaluateVariable
// delegate here.
func (p *Provider) evaluateWithContext(flag string, flatCtx of.FlattenedContext) (*contextResult, *evaluationDetail, error) {
	if !p.ready.Load() || p.client == nil {
		return nil, nil, &evaluationError{
			code:    of.ProviderNotReadyCode,
			message: "provider not ready: Optimizely client is not initialized",
		}
	}

	ctxResult, err := extractContext(flatCtx)
	if err != nil {
		return nil, nil, err
	}

	userCtx := p.client.CreateUserContext(ctxResult.userID, ctxResult.attributes)
	decision := userCtx.Decide(flag, []decide.OptimizelyDecideOptions{decide.IncludeReasons})

	// Detect flag-not-found: the Optimizely SDK returns a decision with
	// empty VariationKey and a reason containing the substring below.
	// This is coupled to the SDK's internal reason text in
	// pkg/decision/reasons/reasons.go (FailNoFlagFoundForFlagKey).
	// If the Go SDK changes that wording, TestBooleanEvaluation/"flag
	// not found returns default" will fail, surfacing the breakage.
	const flagNotFoundSubstring = "No flag was found"
	if decision.VariationKey == "" && !decision.Enabled {
		for _, reason := range decision.Reasons {
			if strings.Contains(reason, flagNotFoundSubstring) {
				return nil, nil, &evaluationError{
					code:    of.FlagNotFoundCode,
					message: fmt.Sprintf("flag '%s' not found", flag),
				}
			}
		}
	}

	detail := &evaluationDetail{
		decision: decisionResult{
			VariationKey: decision.VariationKey,
			Enabled:      decision.Enabled,
			Variables:    decision.Variables,
			FlagKey:      decision.FlagKey,
			RuleKey:      decision.RuleKey,
			Reasons:      decision.Reasons,
		},
	}
	return ctxResult, detail, nil
}

// getVariable extracts a named variable from the decision's variables map.
func getVariable(detail *evaluationDetail, varKey string) (interface{}, error) {
	varsMap := detail.decision.Variables.ToMap()
	val, ok := varsMap[varKey]
	if !ok {
		return nil, &evaluationError{
			code:    of.GeneralCode,
			message: fmt.Sprintf("variable '%s' not found in decision variables", varKey),
		}
	}
	return val, nil
}

// toInt64 converts a variable value to int64.
func toInt64(val interface{}) (int64, error) {
	switch v := val.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case json.Number:
		return v.Int64()
	default:
		return 0, fmt.Errorf("unsupported type %T", val)
	}
}

// toFloat64 converts a variable value to float64.
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case json.Number:
		return v.Float64()
	default:
		return 0, fmt.Errorf("unsupported type %T", val)
	}
}
