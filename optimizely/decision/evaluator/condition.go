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

const (
	exactMatchType  = "exact"
	existsMatchType = "exists"
	ltMatchType     = "lt"
	gtMatchType     = "gt"
)

// ConditionEvaluator evaluates a condition against the given user's attributes
type ConditionEvaluator interface {
	Evaluate(entities.Condition, *entities.ConditionTreeParameters) (bool, error)
}

// CustomAttributeConditionEvaluator evaluates conditions with custom attributes
type CustomAttributeConditionEvaluator struct{}

// Evaluate returns true if the given user's attributes match the condition
func (c CustomAttributeConditionEvaluator) Evaluate(condition entities.Condition, condTreeParams *entities.ConditionTreeParameters) (bool, error) {
	// We should only be evaluating custom attributes
	if condition.Type != customAttributeType {
		return false, fmt.Errorf(`Unable to evaluator condition of type "%s"`, condition.Type)
	}

	var matcher matchers.Matcher
	matchType := condition.Match
	switch matchType {
	case exactMatchType:
		matcher = matchers.ExactMatcher{
			Condition: condition,
		}
	case existsMatchType:
		matcher = matchers.ExistsMatcher{
			Condition: condition,
		}
	case ltMatchType:
		matcher = matchers.LtMatcher{
			Condition: condition,
		}
	case gtMatchType:
		matcher = matchers.GtMatcher{
			Condition: condition,
		}
	default:
		return false, fmt.Errorf(`Invalid Condition matcher "%s"`, condition.Match)
	}

	user := *condTreeParams.User
	result, err := matcher.Match(user)
	return result, err
}

// AudienceConditionEvaluator evaluates conditions with audience condition
type AudienceConditionEvaluator struct{}

// Evaluate returns true if the given user's attributes match the condition
func (c AudienceConditionEvaluator) Evaluate(condition entities.Condition, condTreeParams *entities.ConditionTreeParameters) (bool, error) {
	// We should only be evaluating custom attributes
	if condition.Type != audienceCondition {
		return false, fmt.Errorf(`Unable to evaluator condition of type "%s"`, condition.Type)
	}
	var ok bool
	var audienceID string

	if audienceID, ok = condition.Value.(string); ok {
		if audience, good := condTreeParams.AudienceMap[audienceID]; good {
			condTree := audience.ConditionTree
			conditionTreeEvaluator := NewConditionTreeEvaluator()
			retValue := conditionTreeEvaluator.Evaluate(condTree, condTreeParams)
			return retValue, nil
		}
	}

	return false, fmt.Errorf(`Unable to evaluate nested tree for audience ID "%s"`, audienceID)
}
