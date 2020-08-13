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

func TestSemverLtMatcherLess(t *testing.T) {
	matcher := SemverLtMatcher{
		Condition: entities.Condition{
			Match: "lt",
			Value: "2.2",
			Name:  "is_semantic_less",
		},
	}

	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"is_semantic_less": "2.1.6",
		},
	}
	result, err := matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestSemverLtMatcherFullLess(t *testing.T) {
	matcher := SemverLtMatcher{
		Condition: entities.Condition{
			Match: "lt",
			Value: "2.1.9",
			Name:  "is_semantic_full_less",
		},
	}

	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"is_semantic_full_less": "2.1.6",
		},
	}
	result, err := matcher.Match(user)
	assert.NoError(t, err)
	assert.True(t, result)
}
