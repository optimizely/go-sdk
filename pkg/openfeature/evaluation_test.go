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
	"testing"

	of "github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/v2/pkg/client"
)

// testDatafile is a minimal datafile with a feature flag "test_feature" that
// has a rollout delivering variation "var_1" to all traffic, with variables
// of each type.
const testDatafile = `{
  "version": "4",
  "rollouts": [{
    "id": "rollout_1",
    "experiments": [{
      "id": "rollout_exp_1",
      "key": "rollout_exp_1",
      "status": "Running",
      "layerId": "rollout_layer_1",
      "variations": [{
        "id": "var_1",
        "key": "variation_on",
        "featureEnabled": true,
        "variables": [
          {"id": "v1", "value": "hello"},
          {"id": "v2", "value": "42"},
          {"id": "v3", "value": "3.14"},
          {"id": "v4", "value": "{\"key\":\"val\"}"}
        ]
      }],
      "trafficAllocation": [{"entityId": "var_1", "endOfRange": 10000}],
      "audienceIds": [],
      "forcedVariations": {}
    }]
  }],
  "featureFlags": [{
    "id": "feat_1",
    "key": "test_feature",
    "experimentIds": [],
    "rolloutId": "rollout_1",
    "variables": [
      {"id": "v1", "key": "str_var", "type": "string", "defaultValue": "default_str"},
      {"id": "v2", "key": "int_var", "type": "integer", "defaultValue": "0"},
      {"id": "v3", "key": "dbl_var", "type": "double", "defaultValue": "0.0"},
      {"id": "v4", "key": "json_var", "type": "json", "defaultValue": "{}"}
    ]
  }],
  "experiments": [],
  "events": [{"id": "evt_1", "key": "purchase", "experimentIds": []}],
  "audiences": [],
  "typedAudiences": [],
  "groups": [],
  "attributes": [],
  "accountId": "12345",
  "projectId": "67890",
  "revision": "1",
  "anonymizeIP": false,
  "botFiltering": false
}`

// newTestProvider creates a provider wrapping a real static OptimizelyClient
// initialized with the testDatafile. The returned client must be closed by the
// caller.
func newTestProvider(t *testing.T) (*Provider, *client.OptimizelyClient) {
	t.Helper()
	factory := client.OptimizelyFactory{Datafile: []byte(testDatafile)}
	c, err := factory.StaticClient()
	if err != nil {
		t.Fatalf("failed to create static client: %v", err)
	}
	p := NewProviderWithClient(c)
	p.ready.Store(true)
	return p, c
}

func TestBooleanEvaluation(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	tests := []struct {
		name       string
		flag       string
		flatCtx    of.FlattenedContext
		defaultVal bool
		wantVal    bool
		wantReason of.Reason
		wantErr    bool
	}{
		{
			name:       "user qualifies - feature enabled",
			flag:       "test_feature",
			flatCtx:    of.FlattenedContext{of.TargetingKey: "user-123"},
			defaultVal: false,
			wantVal:    true,
			wantReason: of.TargetingMatchReason,
		},
		{
			name:       "flag not found returns default",
			flag:       "nonexistent_flag",
			flatCtx:    of.FlattenedContext{of.TargetingKey: "user-123"},
			defaultVal: false,
			wantVal:    false,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
		{
			name:       "missing targeting key returns default",
			flag:       "test_feature",
			flatCtx:    of.FlattenedContext{"plan": "enterprise"},
			defaultVal: true,
			wantVal:    true,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.BooleanEvaluation(ctx, tt.flag, tt.defaultVal, tt.flatCtx)
			assert.Equal(t, tt.wantVal, result.Value)
			assert.Equal(t, tt.wantReason, result.ProviderResolutionDetail.Reason)
			if tt.wantErr {
				assert.NotEmpty(t, result.ProviderResolutionDetail.ResolutionError.Error())
			}
		})
	}
}

func TestBooleanEvaluationProviderNotReady(t *testing.T) {
	p := NewProvider("fake-key")
	// provider not initialized — client is nil, ready is false
	ctx := context.Background()
	flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}

	result := p.BooleanEvaluation(ctx, "test_feature", false, flatCtx)
	assert.Equal(t, false, result.Value)
	assert.Equal(t, of.ErrorReason, result.ProviderResolutionDetail.Reason)
	assert.Contains(t, result.ProviderResolutionDetail.ResolutionError.Error(), "not ready")
}

func TestStringEvaluation(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	tests := []struct {
		name       string
		flag       string
		flatCtx    of.FlattenedContext
		defaultVal string
		wantVal    string
		wantReason of.Reason
		wantErr    bool
	}{
		{
			name: "valid string variable with variableKey",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "str_var",
			},
			defaultVal: "fallback",
			wantVal:    "hello",
			wantReason: of.TargetingMatchReason,
		},
		{
			name: "non-string variable coerced via Sprintf",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "int_var",
			},
			defaultVal: "fallback",
			wantVal:    "42",
			wantReason: of.TargetingMatchReason,
		},
		{
			name:       "missing variableKey returns error",
			flag:       "test_feature",
			flatCtx:    of.FlattenedContext{of.TargetingKey: "user-123"},
			defaultVal: "fallback",
			wantVal:    "fallback",
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
		{
			name: "variable not found returns default",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "nonexistent_var",
			},
			defaultVal: "fallback",
			wantVal:    "fallback",
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.StringEvaluation(ctx, tt.flag, tt.defaultVal, tt.flatCtx)
			assert.Equal(t, tt.wantVal, result.Value)
			assert.Equal(t, tt.wantReason, result.ProviderResolutionDetail.Reason)
			if tt.wantErr {
				assert.NotEmpty(t, result.ProviderResolutionDetail.ResolutionError.Error())
			}
		})
	}
}

func TestIntEvaluation(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	tests := []struct {
		name       string
		flag       string
		flatCtx    of.FlattenedContext
		defaultVal int64
		wantVal    int64
		wantReason of.Reason
		wantErr    bool
	}{
		{
			name: "valid int variable",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "int_var",
			},
			defaultVal: 0,
			wantVal:    42,
			wantReason: of.TargetingMatchReason,
		},
		{
			name:       "missing variableKey returns error",
			flag:       "test_feature",
			flatCtx:    of.FlattenedContext{of.TargetingKey: "user-123"},
			defaultVal: 99,
			wantVal:    99,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
		{
			name: "non-numeric variable returns parse error",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "str_var",
			},
			defaultVal: -1,
			wantVal:    -1,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.IntEvaluation(ctx, tt.flag, tt.defaultVal, tt.flatCtx)
			assert.Equal(t, tt.wantVal, result.Value)
			assert.Equal(t, tt.wantReason, result.ProviderResolutionDetail.Reason)
			if tt.wantErr {
				assert.NotEmpty(t, result.ProviderResolutionDetail.ResolutionError.Error())
			}
		})
	}
}

func TestFloatEvaluation(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	tests := []struct {
		name       string
		flag       string
		flatCtx    of.FlattenedContext
		defaultVal float64
		wantVal    float64
		wantReason of.Reason
		wantErr    bool
	}{
		{
			name: "valid float variable",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "dbl_var",
			},
			defaultVal: 0.0,
			wantVal:    3.14,
			wantReason: of.TargetingMatchReason,
		},
		{
			name:       "missing variableKey returns error",
			flag:       "test_feature",
			flatCtx:    of.FlattenedContext{of.TargetingKey: "user-123"},
			defaultVal: 9.99,
			wantVal:    9.99,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
		{
			name: "non-numeric variable returns parse error",
			flag: "test_feature",
			flatCtx: of.FlattenedContext{
				of.TargetingKey: "user-123",
				"variableKey":   "str_var",
			},
			defaultVal: -1.0,
			wantVal:    -1.0,
			wantReason: of.ErrorReason,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.FloatEvaluation(ctx, tt.flag, tt.defaultVal, tt.flatCtx)
			assert.InDelta(t, tt.wantVal, result.Value, 0.001)
			assert.Equal(t, tt.wantReason, result.ProviderResolutionDetail.Reason)
			if tt.wantErr {
				assert.NotEmpty(t, result.ProviderResolutionDetail.ResolutionError.Error())
			}
		})
	}
}

func TestParityWithNativeDecideAPI(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	userID := "parity-user-42"
	attrs := map[string]interface{}{"plan": "enterprise"}

	// Native Decide API
	userCtx := c.CreateUserContext(userID, attrs)
	nativeDecision := userCtx.Decide("test_feature", nil)

	// OpenFeature provider
	flatCtx := of.FlattenedContext{
		of.TargetingKey: userID,
		"plan":          "enterprise",
	}
	ofResult := p.BooleanEvaluation(ctx, "test_feature", false, flatCtx)

	// Parity: enabled state must match
	assert.Equal(t, nativeDecision.Enabled, ofResult.Value,
		"boolean value must match native Decide.Enabled")

	// Parity: variation key must match
	assert.Equal(t, nativeDecision.VariationKey, ofResult.ProviderResolutionDetail.Variant,
		"variant must match native Decide.VariationKey")

	// Parity: variables must match
	flatCtxWithVar := of.FlattenedContext{
		of.TargetingKey: userID,
		"plan":          "enterprise",
		"variableKey":   "str_var",
	}
	strResult := p.StringEvaluation(ctx, "test_feature", "", flatCtxWithVar)
	nativeVars := nativeDecision.Variables.ToMap()
	assert.Equal(t, nativeVars["str_var"], strResult.Value,
		"string variable must match native decision variables")
}

func TestObjectEvaluation(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	t.Run("specific variable via variableKey returns parsed JSON", func(t *testing.T) {
		flatCtx := of.FlattenedContext{
			of.TargetingKey: "user-123",
			"variableKey":   "json_var",
		}
		result := p.ObjectEvaluation(ctx, "test_feature", nil, flatCtx)
		assert.Equal(t, of.TargetingMatchReason, result.ProviderResolutionDetail.Reason)
		valMap, ok := result.Value.(map[string]interface{})
		assert.True(t, ok, "value should be map[string]interface{}")
		assert.Equal(t, "val", valMap["key"])
	})

	t.Run("full variables map when variableKey omitted", func(t *testing.T) {
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.ObjectEvaluation(ctx, "test_feature", nil, flatCtx)
		assert.Equal(t, of.TargetingMatchReason, result.ProviderResolutionDetail.Reason)
		valMap, ok := result.Value.(map[string]interface{})
		assert.True(t, ok, "value should be map[string]interface{}")
		// Should contain all variables
		assert.Contains(t, valMap, "str_var")
		assert.Contains(t, valMap, "int_var")
	})

	t.Run("non-JSON string variable returned as-is", func(t *testing.T) {
		flatCtx := of.FlattenedContext{
			of.TargetingKey: "user-123",
			"variableKey":   "str_var",
		}
		result := p.ObjectEvaluation(ctx, "test_feature", nil, flatCtx)
		assert.Equal(t, of.TargetingMatchReason, result.ProviderResolutionDetail.Reason)
		// "hello" is not valid JSON, so it should be returned as the original string
		assert.Equal(t, "hello", result.Value)
	})

	t.Run("flag not found returns default", func(t *testing.T) {
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.ObjectEvaluation(ctx, "nonexistent_flag", "default_obj", flatCtx)
		assert.Equal(t, "default_obj", result.Value)
		assert.Equal(t, of.ErrorReason, result.ProviderResolutionDetail.Reason)
	})
}

func TestFlagMetadata(t *testing.T) {
	p, c := newTestProvider(t)
	defer c.Close()
	ctx := context.Background()

	t.Run("flagKey populated on successful boolean evaluation", func(t *testing.T) {
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.BooleanEvaluation(ctx, "test_feature", false, flatCtx)
		assert.NotNil(t, result.ProviderResolutionDetail.FlagMetadata)
		flagKey, err := result.ProviderResolutionDetail.FlagMetadata.GetString("flagKey")
		assert.NoError(t, err)
		assert.Equal(t, "test_feature", flagKey)
	})

	t.Run("ruleKey populated when rule matches", func(t *testing.T) {
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.BooleanEvaluation(ctx, "test_feature", false, flatCtx)
		assert.NotNil(t, result.ProviderResolutionDetail.FlagMetadata)
		// The test datafile has a rollout rule — ruleKey should be present
		ruleKey, err := result.ProviderResolutionDetail.FlagMetadata.GetString("ruleKey")
		assert.NoError(t, err)
		assert.NotEmpty(t, ruleKey)
	})

	t.Run("reasons populated as string slice", func(t *testing.T) {
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.BooleanEvaluation(ctx, "test_feature", false, flatCtx)
		assert.NotNil(t, result.ProviderResolutionDetail.FlagMetadata)
		reasons := result.ProviderResolutionDetail.FlagMetadata["reasons"]
		reasonsSlice, ok := reasons.([]string)
		assert.True(t, ok, "reasons should be []string")
		assert.NotEmpty(t, reasonsSlice)
	})

	t.Run("FlagMetadata nil on error path", func(t *testing.T) {
		// Flag not found — error path
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := p.BooleanEvaluation(ctx, "nonexistent_flag", false, flatCtx)
		assert.Equal(t, of.ErrorReason, result.ProviderResolutionDetail.Reason)
		assert.Nil(t, result.ProviderResolutionDetail.FlagMetadata)
	})

	t.Run("FlagMetadata nil when provider not ready", func(t *testing.T) {
		notReadyProvider := NewProvider("fake-key")
		flatCtx := of.FlattenedContext{of.TargetingKey: "user-123"}
		result := notReadyProvider.BooleanEvaluation(ctx, "test_feature", false, flatCtx)
		assert.Equal(t, of.ErrorReason, result.ProviderResolutionDetail.Reason)
		assert.Nil(t, result.ProviderResolutionDetail.FlagMetadata)
	})
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    int64
		wantErr bool
	}{
		{name: "float64", input: float64(42), want: 42},
		{name: "int64", input: int64(99), want: 99},
		{name: "int", input: int(7), want: 7},
		{name: "string", input: "123", want: 123},
		{name: "json.Number", input: json.Number("456"), want: 456},
		{name: "invalid string", input: "not_a_number", wantErr: true},
		{name: "unsupported type", input: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toInt64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantErr bool
	}{
		{name: "float64", input: float64(3.14), want: 3.14},
		{name: "int64", input: int64(10), want: 10.0},
		{name: "int", input: int(5), want: 5.0},
		{name: "string", input: "2.718", want: 2.718},
		{name: "json.Number", input: json.Number("1.5"), want: 1.5},
		{name: "invalid string", input: "not_a_number", wantErr: true},
		{name: "unsupported type", input: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toFloat64(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.InDelta(t, tt.want, got, 0.001)
		})
	}
}

func TestErrorDetailGenericError(t *testing.T) {
	err := fmt.Errorf("something unexpected")
	detail := errorDetail(err)
	assert.Equal(t, of.ErrorReason, detail.Reason)
	assert.Contains(t, detail.ResolutionError.Error(), "something unexpected")
}
