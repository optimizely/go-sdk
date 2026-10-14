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
	"fmt"

	of "github.com/open-feature/go-sdk/openfeature"
)

const variableKeyAttr = "variableKey"

// contextResult holds the parsed evaluation context.
type contextResult struct {
	userID         string
	attributes     map[string]interface{}
	variableKey    string
	hasVariableKey bool
}

// contextError represents an error during context extraction.
type contextError struct {
	code    of.ErrorCode
	message string
}

func (e *contextError) Error() string {
	return fmt.Sprintf("openfeature context error (%s): %s", e.code, e.message)
}

// extractContext converts an OpenFeature FlattenedContext into Optimizely
// user ID, attributes, and an optional variableKey. The variableKey is
// stripped from attributes before passing to Optimizely. Unsupported
// attribute types (maps, slices) are silently dropped.
func extractContext(flatCtx of.FlattenedContext) (*contextResult, error) {
	if flatCtx == nil {
		return nil, &contextError{
			code:    of.TargetingKeyMissingCode,
			message: "evaluation context is nil",
		}
	}

	// Extract targeting key
	tkRaw, ok := flatCtx[of.TargetingKey]
	if !ok {
		return nil, &contextError{
			code:    of.TargetingKeyMissingCode,
			message: "targeting key is required",
		}
	}

	userID, ok := tkRaw.(string)
	if !ok {
		return nil, &contextError{
			code:    of.InvalidContextCode,
			message: "targeting key must be a string",
		}
	}

	if userID == "" {
		return nil, &contextError{
			code:    of.TargetingKeyMissingCode,
			message: "targeting key must not be empty",
		}
	}

	result := &contextResult{
		userID:     userID,
		attributes: make(map[string]interface{}),
	}

	for key, val := range flatCtx {
		if key == of.TargetingKey {
			continue
		}

		// Extract and strip variableKey
		if key == variableKeyAttr {
			if vk, ok := val.(string); ok {
				result.variableKey = vk
				result.hasVariableKey = true
			}
			continue
		}

		// Filter unsupported types: only pass scalars
		if isSupportedAttributeType(val) {
			result.attributes[key] = val
		}
	}

	return result, nil
}

// isSupportedAttributeType returns true for types the Optimizely SDK
// supports as user attributes: string, bool, int, float64.
func isSupportedAttributeType(val interface{}) bool {
	switch val.(type) {
	case string, bool, int, int64, float64:
		return true
	default:
		return false
	}
}
