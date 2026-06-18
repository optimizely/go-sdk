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

func TestMapReason(t *testing.T) {
	tests := []struct {
		name         string
		variationKey string
		enabled      bool
		hasError     bool
		wantReason   openfeature.Reason
	}{
		{
			name:         "variation set and enabled returns TARGETING_MATCH",
			variationKey: "variation_1",
			enabled:      true,
			wantReason:   openfeature.TargetingMatchReason,
		},
		{
			name:         "variation set and disabled returns DISABLED",
			variationKey: "variation_1",
			enabled:      false,
			wantReason:   openfeature.DisabledReason,
		},
		{
			name:         "no variation returns DEFAULT",
			variationKey: "",
			enabled:      false,
			wantReason:   openfeature.DefaultReason,
		},
		{
			name:       "error returns ERROR",
			hasError:   true,
			wantReason: openfeature.ErrorReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapReason(tt.variationKey, tt.enabled, tt.hasError)
			assert.Equal(t, tt.wantReason, got)
		})
	}
}

func TestMakeResolutionError(t *testing.T) {
	tests := []struct {
		name    string
		code    openfeature.ErrorCode
		message string
	}{
		{
			name:    "flag not found",
			code:    openfeature.FlagNotFoundCode,
			message: "flag 'my_flag' not found",
		},
		{
			name:    "targeting key missing",
			code:    openfeature.TargetingKeyMissingCode,
			message: "targeting key is required",
		},
		{
			name:    "type mismatch",
			code:    openfeature.TypeMismatchCode,
			message: "expected string",
		},
		{
			name:    "parse error",
			code:    openfeature.ParseErrorCode,
			message: "cannot parse value",
		},
		{
			name:    "provider not ready",
			code:    openfeature.ProviderNotReadyCode,
			message: "provider not initialized",
		},
		{
			name:    "general error",
			code:    openfeature.GeneralCode,
			message: "unexpected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resErr := makeResolutionError(tt.code, tt.message)
			assert.Contains(t, resErr.Error(), tt.message)
		})
	}
}
