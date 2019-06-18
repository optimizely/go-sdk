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
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/decision/evaluator/matchers"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ConditionEvaluator evaluates a condition against the given user's attributes
type ConditionEvaluator interface {
	Evaluate(entities.Condition, entities.UserContext) (bool, error)
}

// CustomAttributeConditionEvaluator evaluates conditions with custom attributes
type CustomAttributeConditionEvaluator struct{}

// Evaluate returns true if the given user's attributes match the condition
func (c CustomAttributeConditionEvaluator) Evaluate(condition entities.Condition, user entities.UserContext) (bool, error) {
	// We should only be evaluating custom attributes
	if condition.Type != customAttributeType {
		return false, fmt.Errorf(`Unable to evaluator condition of type "%s"`, condition.Type)
	}

	var matcher matchers.Matcher
	matchType := condition.Match
	switch matchType {
	case "exact":
		matcher = matchers.ExactMatcher{
			Condition: condition,
		}
	}

	result, err := matcher.Match(user)
	return result, err
}
