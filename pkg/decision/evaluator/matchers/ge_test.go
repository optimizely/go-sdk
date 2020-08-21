/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

package matchers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/pkg/entities"
)

var geMatcher, _ = Get(GeMatchType)

func TestGeMatcherInt(t *testing.T) {
	condition := entities.Condition{
		Match: "ge",
		Value: 42,
		Name:  "int_42",
	}

	// Test match - same type
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 43,
		},
	}
	result, err := gtMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match int to float
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.9999,
		},
	}

	result, err = geMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match: 42.9999 converts to 42
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.00000,
		},
	}

	result, err = geMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 41,
		},
	}

	result, err = geMatcher(condition, user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": 42,
		},
	}

	_, err = geMatcher(condition, user)
	assert.Error(t, err)

	// Test wrong int
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": "some_string",
		},
	}

	_, err = geMatcher(condition, user)
	assert.Error(t, err)

	// Test bad condition
	condition = entities.Condition{
		Match: "ge",
		Value: "123",
		Name:  "int_42",
	}
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": "some_string",
		},
	}

	_, err = gtMatcher(condition, user)
	assert.Error(t, err)
}

func TestGeMatcherFloat(t *testing.T) {
	condition := entities.Condition{
		Match: "ge",
		Value: 4.2,
		Name:  "float_4_2",
	}

	// Test match float to int
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 5,
		},
	}
	result, err := geMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.29999,
		},
	}
	result, err = geMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
		},
	}

	result, err = geMatcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_3": 4.2,
		},
	}

	_, err = geMatcher(condition, user)
	assert.Error(t, err)
}
