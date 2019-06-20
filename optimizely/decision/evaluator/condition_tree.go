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
type conditionEvalResult int

const customAttributeType = "custom_attribute"

const (
	// TRUE means the condition passes
	TRUE conditionEvalResult = iota
	// FALSE means the condition does not pass
	FALSE
	// NULL means the condition could not be evaluated
	NULL
)

const (
	// "and" operator returns true if all conditions evaluate to true
	andOperator = "and"
	// "not" operator negates the result of the given condition
	notOperator = "not"
	// "or" operator returns true if any of the conditions evaluate to true
	orOperator = "or"
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
	operator := node.Operator
	if operator != "" {
		switch operator {
		case andOperator:
			return c.evaluateAnd(node.Nodes, user)
		case notOperator:
			return c.evaluateNot(node.Nodes, user)
		case orOperator:
			fallthrough
		default:
			return c.evaluateOr(node.Nodes, user)
		}
	}

	conditionEvaluator, ok := c.conditionEvaluatorMap[node.Condition.Type]
	if !ok {
		// TODO(mng): log error
		return NULL
	}
	result, err := conditionEvaluator.Evaluate(node.Condition, user)
	if err != nil {
		// TODO(mng): log error
		return NULL
	}
	if result == true {
		return TRUE
	}
	return FALSE
}

func (c ConditionTreeEvaluator) evaluateAnd(nodes []*entities.ConditionTreeNode, user entities.UserContext) conditionEvalResult {
	sawNull := false
	for _, node := range nodes {
		result := c.evaluate(node, user)
		if result == FALSE {
			return FALSE
		} else if result == NULL {
			sawNull = true
		}
	}

	if sawNull {
		return NULL
	}

	return TRUE
}

func (c ConditionTreeEvaluator) evaluateNot(nodes []*entities.ConditionTreeNode, user entities.UserContext) conditionEvalResult {
	if len(nodes) > 0 {
		result := c.evaluate(nodes[0], user)
		if result == NULL {
			return NULL
		} else if result == TRUE {
			return FALSE
		} else if result == FALSE {
			return TRUE
		}
	}
	return NULL
}

func (c ConditionTreeEvaluator) evaluateOr(nodes []*entities.ConditionTreeNode, user entities.UserContext) conditionEvalResult {
	sawNull := false
	for _, node := range nodes {
		result := c.evaluate(node, user)
		if result == TRUE {
			return TRUE
		} else if result == NULL {
			sawNull = true
		}
	}

	if sawNull {
		return NULL
	}

	return FALSE
}
