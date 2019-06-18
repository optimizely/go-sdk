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
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// conditionEvalResult is the result of evaluating a Condition, which can be true/false or null if the condition could not be evaluated
type conditionEvalResult string

const customAttributeType = "custom_attribute"

const (
	// TRUE means the condition passes
	TRUE conditionEvalResult = "TRUE"
	// FALSE means the condition does not pass
	FALSE conditionEvalResult = "FALSE"
	// NULL means the condition could not be evaluated
	NULL conditionEvalResult = "NULL"
)

// ConditionTreeEvaluator evaluates a condition tree
type ConditionTreeEvaluator struct {
	conditionEvaluatorMap map[string]ConditionEvaluator
}

// NewConditionTreeEvaluator creates a condition tree evaluator with the out-of-the-box condition evaluators
func NewConditionTreeEvaluator() *ConditionTreeEvaluator {
	// For now, only one evaluator per attribute type
	conditionEvaluatorMap := make(map[string]ConditionEvaluator)
	conditionEvaluatorMap[customAttributeType] = CustomAttributeConditionEvaluator{}
	return &ConditionTreeEvaluator{
		conditionEvaluatorMap: conditionEvaluatorMap,
	}
}

// Evaluate returns true if the userAttributes satisfy the given condition tree
func (c ConditionTreeEvaluator) Evaluate(node *entities.ConditionTreeNode, user entities.UserContext) bool {
	// This wrapper method converts the conditionEvalResult to a boolean
	result := c.evaluate(node, user)
	return result == TRUE
}

// Helper method to recursively evaluate a condition tree
func (c ConditionTreeEvaluator) evaluate(node *entities.ConditionTreeNode, user entities.UserContext) conditionEvalResult {
	// @TODO(mng): Implement tree evaluator with and/or/not operators
	return TRUE
}
