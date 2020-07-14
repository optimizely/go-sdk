/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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

// Package evaluator //
package evaluator

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

const customAttributeType = "custom_attribute"

const (
	// "and" operator returns true if all conditions evaluate to true
	andOperator = "and"
	// "not" operator negates the result of the given condition
	notOperator = "not"
	// "or" operator returns true if any of the conditions evaluate to true
	// orOperator = "or"
)

// TreeEvaluator evaluates a tree
type TreeEvaluator interface {
	Evaluate(*entities.TreeNode, *entities.TreeParameters) (evalResult, isValid bool)
}

// MixedTreeEvaluator evaluates a tree of mixed node types (condition node or audience nodes)
type MixedTreeEvaluator struct {
	logger logging.OptimizelyLogProducer
}

// NewMixedTreeEvaluator creates a condition tree evaluator with the out-of-the-box condition evaluators
func NewMixedTreeEvaluator(logger logging.OptimizelyLogProducer) *MixedTreeEvaluator {
	return &MixedTreeEvaluator{logger: logger}
}

// Evaluate returns whether the userAttributes satisfy the given condition tree and the evaluation of the condition is valid or not (to handle null bubbling)
func (c MixedTreeEvaluator) Evaluate(node *entities.TreeNode, condTreeParams *entities.TreeParameters) (evalResult, isValid bool) {
	operator := node.Operator
	if operator != "" {
		switch operator {
		case andOperator:
			return c.evaluateAnd(node.Nodes, condTreeParams)
		case notOperator:
			return c.evaluateNot(node.Nodes, condTreeParams)
		default: // orOperator
			return c.evaluateOr(node.Nodes, condTreeParams)
		}
	}

	var result bool
	var err error
	switch v := node.Item.(type) {
	case entities.Condition:
		evaluator := NewCustomAttributeConditionEvaluator(c.logger)
		result, err = evaluator.Evaluate(node.Item.(entities.Condition), condTreeParams)
	case string:
		evaluator := NewAudienceConditionEvaluator(c.logger)
		result, err = evaluator.Evaluate(node.Item.(string), condTreeParams)
	default:
		fmt.Printf("I don't know about type %T!\n", v)
		return false, false
	}

	if err != nil {
		// Result is invalid
		return false, false
	}
	return result, true
}

func (c MixedTreeEvaluator) evaluateAnd(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters) (evalResult, isValid bool) {
	sawInvalid := false
	for _, node := range nodes {
		result, isValid := c.Evaluate(node, condTreeParams)
		if !isValid {
			return false, isValid
		} else if !result {
			return result, isValid
		}
	}

	if sawInvalid {
		// bubble up the invalid result
		return false, false
	}

	return true, true
}

func (c MixedTreeEvaluator) evaluateNot(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters) (evalResult, isValid bool) {
	if len(nodes) > 0 {
		result, isValid := c.Evaluate(nodes[0], condTreeParams)
		if !isValid {
			return false, false
		}
		return !result, isValid
	}
	return false, false
}

func (c MixedTreeEvaluator) evaluateOr(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters) (evalResult, isValid bool) {
	sawInvalid := false
	for _, node := range nodes {
		result, isValid := c.Evaluate(node, condTreeParams)
		if !isValid {
			sawInvalid = true
		} else if result {
			return result, isValid
		}
	}

	if sawInvalid {
		// bubble up the invalid result
		return false, false
	}

	return false, true
}
