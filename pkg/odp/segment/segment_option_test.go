/****************************************************************************
 * Copyright 2023, Optimizely, Inc. and contributors                        *
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

package segment

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
	options = []string{"IGNORE_CACHE"}
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 1)
	assert.Equal(t, IgnoreCache, translatedOptions[0])

	// Checking after appending further options
	options = append(options, "RESET_CACHE")
	translatedOptions, err = TranslateOptions(options)
	assert.NoError(t, err)
	assert.Len(t, translatedOptions, 2)
	assert.Equal(t, ResetCache, translatedOptions[1])
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
