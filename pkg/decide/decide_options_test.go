/****************************************************************************
 * Copyright 2021-2025, Optimizely, Inc. and contributors                   *
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

package decide

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateOptionsValidCases(t *testing.T) {
	// Checking nil options
	translatedOptions, err := TranslateOptions(nil)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 0)

	// Checking empty options
	options := []string{}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 0)

	// Checking correct options
	options = []string{"DISABLE_DECISION_EVENT", "ENABLED_FLAGS_ONLY"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 2)
	assert.Equal(t, DisableDecisionEvent, translatedOptions[0])
	assert.Equal(t, EnabledFlagsOnly, translatedOptions[1])

	// Checking after appending further options
	options = append(options, "IGNORE_USER_PROFILE_SERVICE", "EXCLUDE_VARIABLES", "INCLUDE_REASONS")
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 5)
	assert.Equal(t, IgnoreUserProfileService, translatedOptions[2])
	assert.Equal(t, ExcludeVariables, translatedOptions[3])
	assert.Equal(t, IncludeReasons, translatedOptions[4])
}

func TestTranslateOptionsInvalidCases(t *testing.T) {
	// Checking empty value as option
	options := []string{""}
	translatedOptions, err := TranslateOptions(options)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("invalid option: %v", options[0]), err)
	assert.Len(t, translatedOptions, 0)

	// Checking invalid value as option
	options[0] = "INVALID"
	translatedOptions, err = TranslateOptions(options)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("invalid option: %v", options[0]), err)
	assert.Len(t, translatedOptions, 0)
}

// TestTranslateOptionsCMABOptions tests the new CMAB-related options
func TestTranslateOptionsCMABOptions(t *testing.T) {
	// Test IGNORE_CMAB_CACHE option
	options := []string{"IGNORE_CMAB_CACHE"}
	translatedOptions, err := TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 1)
	assert.Equal(t, IgnoreCMABCache, translatedOptions[0])

	// Test RESET_CMAB_CACHE option
	options = []string{"RESET_CMAB_CACHE"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 1)
	assert.Equal(t, ResetCMABCache, translatedOptions[0])

	// Test INVALIDATE_USER_CMAB_CACHE option
	options = []string{"INVALIDATE_USER_CMAB_CACHE"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 1)
	assert.Equal(t, InvalidateUserCMABCache, translatedOptions[0])

	// Test all CMAB options together
	options = []string{"IGNORE_CMAB_CACHE", "RESET_CMAB_CACHE", "INVALIDATE_USER_CMAB_CACHE"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 3)
	assert.Equal(t, IgnoreCMABCache, translatedOptions[0])
	assert.Equal(t, ResetCMABCache, translatedOptions[1])
	assert.Equal(t, InvalidateUserCMABCache, translatedOptions[2])

	// Test CMAB options with other options
	options = []string{"DISABLE_DECISION_EVENT", "IGNORE_CMAB_CACHE", "ENABLED_FLAGS_ONLY", "RESET_CMAB_CACHE", "INVALIDATE_USER_CMAB_CACHE"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 5)
	assert.Equal(t, DisableDecisionEvent, translatedOptions[0])
	assert.Equal(t, IgnoreCMABCache, translatedOptions[1])
	assert.Equal(t, EnabledFlagsOnly, translatedOptions[2])
	assert.Equal(t, ResetCMABCache, translatedOptions[3])
	assert.Equal(t, InvalidateUserCMABCache, translatedOptions[4])
}
