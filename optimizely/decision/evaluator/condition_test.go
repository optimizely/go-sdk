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

package evaluator

import (
	"testing"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
)

func TestCustomAttributeConditionEvaluator(t *testing.T) {
	conditionEvaluator := CustomAttributeConditionEvaluator{}
	condition := entities.Condition{
		Match: "exact",
		Value: "foo",
		Name:  "string_foo",
		Type:  "custom_attribute",
	}

	// Test condition passes
	user := entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
		},
	}

	condTreeParams := entities.NewTreeParameters(&user, map[string]entities.Audience{})
	result, _ := conditionEvaluator.Evaluate(condition, condTreeParams)
	assert.Equal(t, result, true)

	// Test condition fails
	user = entities.UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "not_foo",
		},
	}
	result, _ = conditionEvaluator.Evaluate(condition, condTreeParams)
	assert.Equal(t, result, false)
}
