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

	"github.com/optimizely/go-sdk/optimizely/entities"
)

func TestLtMatcherInt(t *testing.T) {
	matcher := LtMatcher{
		Condition: entities.Condition{
			Match: "lt",
			Value: 42,
			Name:  "int_42",
		},
	}

	// Test match - same type
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 41,
		},
	}
	result, err := matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match int to float
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 41.9999,
		},
	}

	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 42.00000,
		},
	}

	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_42": 43,
		},
	}

	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test error case
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"int_43": 42,
		},
	}

	_, err = matcher.Match(user)
	assert.Error(t, err)
}

func TestLtMatcherFloat(t *testing.T) {
	matcher := LtMatcher{
		Condition: entities.Condition{
			Match: "lt",
			Value: 4.2,
			Name:  "float_4_2",
		},
	}

	// Test match float to int
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4,
		},
	}
	result, err := matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.19999,
		},
	}
	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_2": 4.2,
		},
	}

	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test error case
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"float_4_3": 4.2,
		},
	}

	_, err = matcher.Match(user)
	assert.Error(t, err)
}
