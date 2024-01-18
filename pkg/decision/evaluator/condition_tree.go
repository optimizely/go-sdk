/****************************************************************************
 * Copyright 2019-2022, Optimizely, Inc. and contributors                   *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// String constant representing custom attribute condition type.
const customAttributeType = "custom_attribute"

// String constant representing a third-party condition type.
const thirdPartyDimension = "third_party_dimension"

// Valid types allowed for validation
var validTypes = [...]string{customAttributeType, thirdPartyDimension}

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
	Evaluate(*entities.TreeNode, *entities.TreeParameters, *decide.Options) (evalResult, isValid bool, reasons decide.DecisionReasons)
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
func (c MixedTreeEvaluator) Evaluate(node *entities.TreeNode, condTreeParams *entities.TreeParameters, options *decide.Options) (evalResult, isValid bool, reasons decide.DecisionReasons) {
	reasons = decide.NewDecisionReasons(options)
	operator := node.Operator
	if operator != "" {
		switch operator {
		case andOperator:
			return c.evaluateAnd(node.Nodes, condTreeParams, options)
		case notOperator:
			return c.evaluateNot(node.Nodes, condTreeParams, options)
		default: // orOperator
			return c.evaluateOr(node.Nodes, condTreeParams, options)
		}
	}

	var result bool
	var err error
	var decisionReasons decide.DecisionReasons
	switch v := node.Item.(type) {
	case entities.Condition:
		evaluator := NewCustomAttributeConditionEvaluator(c.logger)
		result, decisionReasons, err = evaluator.Evaluate(node.Item.(entities.Condition), condTreeParams, options)
		reasons.Append(decisionReasons)
	case string:
		evaluator := NewAudienceConditionEvaluator(c.logger)
		result, decisionReasons, err = evaluator.Evaluate(node.Item.(string), condTreeParams, options)
		reasons.Append(decisionReasons)
	default:
		fmt.Printf("I don't know about type %T!\n", v)
		return false, false, reasons
	}

	if err != nil {
		// Result is invalid
		return false, false, reasons
	}
	return result, true, reasons
}

func (c MixedTreeEvaluator) evaluateAnd(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters, options *decide.Options) (evalResult, isValid bool, reasons decide.DecisionReasons) {
	finalReasons := decide.NewDecisionReasons(options)
	sawInvalid := false
	for _, node := range nodes {
		result, isValid, decisionReasons := c.Evaluate(node, condTreeParams, options)
		finalReasons.Append(decisionReasons)
		if !isValid {
			return false, isValid, finalReasons
		} else if !result {
			return result, isValid, finalReasons
		}
	}

	if sawInvalid {
		// bubble up the invalid result
		return false, false, finalReasons
	}

	return true, true, finalReasons
}

func (c MixedTreeEvaluator) evaluateNot(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters, options *decide.Options) (evalResult, isValid bool, reasons decide.DecisionReasons) {
	finalReasons := decide.NewDecisionReasons(options)
	if len(nodes) > 0 {
		result, isValid, decisionReasons := c.Evaluate(nodes[0], condTreeParams, options)
		finalReasons.Append(decisionReasons)
		if !isValid {
			return false, false, finalReasons
		}
		return !result, isValid, finalReasons
	}
	return false, false, finalReasons
}

func (c MixedTreeEvaluator) evaluateOr(nodes []*entities.TreeNode, condTreeParams *entities.TreeParameters, options *decide.Options) (evalResult, isValid bool, reasons decide.DecisionReasons) {
	finalReasons := decide.NewDecisionReasons(options)
	sawInvalid := false
	for _, node := range nodes {
		result, isValid, decisionReasons := c.Evaluate(node, condTreeParams, options)
		finalReasons.Append(decisionReasons)
		if !isValid {
			sawInvalid = true
		} else if result {
			return result, isValid, finalReasons
		}
	}

	if sawInvalid {
		// bubble up the invalid result
		return false, false, finalReasons
	}

	return false, true, finalReasons
}
