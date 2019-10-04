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

func TestSubstringMatcher(t *testing.T) {
	matcher := SubstringMatcher{
		Condition: entities.Condition{
			Match: "substring",
			Value: "foo",
			Name:  "string_foo",
		},
	}

	// Test match
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foobar",
		},
	}

	result, err := matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test no match
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "bar",
		},
	}

	result, err = matcher.Match(user)
	assert.NoError(t, err)
	assert.False(t, result)

	// Test error case
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"not_string_foo": "foo",
		},
	}

	_, err = matcher.Match(user)
	assert.Error(t, err)
}
