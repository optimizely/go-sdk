/****************************************************************************
 * Copyright 2025, Optimizely, Inc. and contributors                        *
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

package featuretoggle

import "testing"

func TestHoldoutEnabled_DefaultDisabled(t *testing.T) {
	// Reset any test overrides
	HoldoutEnabledForTesting = nil
	
	// Should be disabled by default
	if HoldoutEnabled() {
		t.Error("Expected holdouts to be disabled by default")
	}
}

func TestHoldoutEnabled_CanBeOverriddenForTesting(t *testing.T) {
	// Test enabling
	enabled := true
	HoldoutEnabledForTesting = &enabled
	defer func() { HoldoutEnabledForTesting = nil }()
	
	if !HoldoutEnabled() {
		t.Error("Expected holdouts to be enabled when override is set to true")
	}
	
	// Test disabling
	disabled := false
	HoldoutEnabledForTesting = &disabled
	
	if HoldoutEnabled() {
		t.Error("Expected holdouts to be disabled when override is set to false")
	}
}