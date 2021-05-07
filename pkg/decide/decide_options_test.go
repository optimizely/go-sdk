/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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
