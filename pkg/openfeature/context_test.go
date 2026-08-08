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
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/stretchr/testify/assert"
)

func TestExtractContext(t *testing.T) {
	tests := []struct {
		name           string
		flatCtx        openfeature.FlattenedContext
		wantUserID     string
		wantAttrs      map[string]interface{}
		wantVarKey     string
		wantHasVarKey  bool
		wantErr        bool
		wantErrCode    openfeature.ErrorCode
	}{
		{
			name: "valid targeting key and attributes",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: "user-123",
				"plan":                  "enterprise",
				"age":                   float64(30),
			},
			wantUserID:    "user-123",
			wantAttrs:     map[string]interface{}{"plan": "enterprise", "age": float64(30)},
			wantHasVarKey: false,
		},
		{
			name:        "missing targeting key",
			flatCtx:     openfeature.FlattenedContext{"plan": "enterprise"},
			wantErr:     true,
			wantErrCode: openfeature.TargetingKeyMissingCode,
		},
		{
			name:        "empty targeting key",
			flatCtx:     openfeature.FlattenedContext{openfeature.TargetingKey: ""},
			wantErr:     true,
			wantErrCode: openfeature.TargetingKeyMissingCode,
		},
		{
			name: "targeting key wrong type",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: 12345,
			},
			wantErr:     true,
			wantErrCode: openfeature.InvalidContextCode,
		},
		{
			name: "variableKey extracted and stripped",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: "user-123",
				variableKeyAttr:         "banner_text",
				"plan":                  "enterprise",
			},
			wantUserID:    "user-123",
			wantAttrs:     map[string]interface{}{"plan": "enterprise"},
			wantVarKey:    "banner_text",
			wantHasVarKey: true,
		},
		{
			name: "variableKey wrong type ignored",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: "user-123",
				variableKeyAttr:         12345,
			},
			wantUserID:    "user-123",
			wantAttrs:     map[string]interface{}{},
			wantHasVarKey: false,
		},
		{
			name: "unsupported attribute types filtered",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: "user-123",
				"name":                  "Alice",
				"nested_obj":            map[string]interface{}{"key": "val"},
				"slice_val":             []string{"a", "b"},
				"count":                 float64(5),
				"enabled":               true,
			},
			wantUserID: "user-123",
			wantAttrs: map[string]interface{}{
				"name":    "Alice",
				"count":   float64(5),
				"enabled": true,
			},
			wantHasVarKey: false,
		},
		{
			name:       "nil context",
			flatCtx:    nil,
			wantErr:    true,
			wantErrCode: openfeature.TargetingKeyMissingCode,
		},
		{
			name: "targeting key only no attributes",
			flatCtx: openfeature.FlattenedContext{
				openfeature.TargetingKey: "user-456",
			},
			wantUserID:    "user-456",
			wantAttrs:     map[string]interface{}{},
			wantHasVarKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractContext(tt.flatCtx)
			if tt.wantErr {
				assert.Error(t, err)
				ctxErr, ok := err.(*contextError)
				assert.True(t, ok, "error should be *contextError")
				assert.Equal(t, tt.wantErrCode, ctxErr.code)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantUserID, result.userID)
			assert.Equal(t, tt.wantAttrs, result.attributes)
			assert.Equal(t, tt.wantVarKey, result.variableKey)
			assert.Equal(t, tt.wantHasVarKey, result.hasVariableKey)
		})
	}
}
