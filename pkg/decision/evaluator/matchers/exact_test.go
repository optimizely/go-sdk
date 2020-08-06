/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

func TestExactMatcherString(t *testing.T) {
	matcher, _ := Get(ExactMatchType)
	condition := entities.Condition{
		Match: "exact",
		Value: "foo",
		Name:  "string_foo",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}
	result, err := matcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not_foo",
		},
	}

	result, err = matcher(condition, user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_not_foo": "foo",
		},
	}

	_, err = matcher(condition, user)
	assert.Error(t, err)
}

func TestExactMatcherBool(t *testing.T) {
	matcher, _ := Get(ExactMatchType)
	condition := entities.Condition{
		Match: "exact",
		Value: true,
		Name:  "bool_true",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": true,
		},
	}
	result, err := matcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"bool_true": false,
		},
	}

	result, err = matcher(condition, user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"not_bool_true": true,
		},
	}

	_, err = matcher(condition, user)
	assert.Error(t, err)
}

func TestExactMatcherInt(t *testing.T) {
	matcher, _ := Get(ExactMatchType)
	condition := entities.Condition{
		Match: "exact",
		Value: 42,
		Name:  "int_42",
	}

	// Test match - same type
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42,
		},
	}
	result, err := matcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match int to float
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.0,
		},
	}

	result, err = matcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 43,
		},
	}

	result, err = matcher(condition, user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": 42,
		},
	}

	_, err = matcher(condition, user)
	assert.Error(t, err)
}

func TestExactMatcherFloat(t *testing.T) {
	matcher, _ := Get(ExactMatchType)
	condition := entities.Condition{
		Match: "exact",
		Value: 4.2,
		Name:  "float_4_2",
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
		},
	}
	result, err := matcher(condition, user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.3,
		},
	}

	result, err = matcher(condition, user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test attribute not found
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_3": 4.2,
		},
	}

	_, err = matcher(condition, user)
	assert.Error(t, err)
}
