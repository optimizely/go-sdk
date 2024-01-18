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

// Package mappers ...
package mappers

import (
	"errors"
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"github.com/optimizely/go-sdk/v2/pkg/decision/evaluator/matchers"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
)

var errEmptyTree = errors.New("empty tree")
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// OperatorType defines logical operator for conditions
type OperatorType string

// Default conditional operators
const (
	And OperatorType = "and"
	Or  OperatorType = "or"
	Not OperatorType = "not"
)

// Takes the conditions array from the audience in the datafile and turns it into a condition tree
func buildConditionTree(conditions interface{}) (conditionTree *entities.TreeNode, odpSegments []string, retErr error) {

	parsedConditions, retErr := parseConditions(conditions)
	if retErr != nil {
		return nil, nil, retErr
	}
	odpSegments = []string{}
	value := reflect.ValueOf(parsedConditions)
	visited := make(map[interface{}]bool)

	conditionTree = &entities.TreeNode{}

	var populateConditions func(v reflect.Value, root *entities.TreeNode)
	populateConditions = func(v reflect.Value, root *entities.TreeNode) {

		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if v.Kind() == reflect.Ptr {
				// Check for recursive data
				if visited[v.Interface()] {
					return
				}
				visited[v.Interface()] = true
			}
			v = v.Elem()
		}

		switch v.Kind() {

		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				n := &entities.TreeNode{}
				typedV := v.Index(i).Interface()
				switch value := typedV.(type) {
				case string:
					n.Operator = value
					root.Operator = n.Operator
					continue

				case map[string]interface{}:
					if err := createLeafCondition(value, n); err != nil {
						retErr = err
						return
					}
					// Extract odp segment from leaf node if applicable
					extractSegment(&odpSegments, n)
				}

				root.Nodes = append(root.Nodes, n)

				populateConditions(v.Index(i), n)
			}
		}
	}

	// Check for leaf conditions
	if value.Kind() == reflect.Map {
		typedV := value.Interface()
		if v, ok := typedV.(map[string]interface{}); ok {
			n := &entities.TreeNode{}
			if err := createLeafCondition(v, n); err != nil {
				retErr = err
				return nil, nil, retErr
			}
			// Extract odp segment from leaf node if applicable
			extractSegment(&odpSegments, n)
			conditionTree.Operator = "or"
			conditionTree.Nodes = append(conditionTree.Nodes, n)
		}
	} else {
		populateConditions(value, conditionTree)
	}

	if conditionTree.Nodes == nil && conditionTree.Operator == "" {
		retErr = errEmptyTree
		conditionTree = nil
	}
	return conditionTree, odpSegments, retErr
}

// Parses conditions for audience in the datafile
func parseConditions(conditions interface{}) (parsedConditions interface{}, retErr error) {
	switch v := conditions.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsedConditions); err != nil {
			return nil, err
		}
	default:
		parsedConditions = conditions
	}
	return parsedConditions, nil
}

// Creates condition for the leaf node in the condition tree
func createLeafCondition(typedV map[string]interface{}, node *entities.TreeNode) error {
	jsonBody, err := json.Marshal(typedV)
	if err != nil {
		return err
	}
	condition := entities.Condition{}
	if err := json.Unmarshal(jsonBody, &condition); err != nil {
		return err
	}
	condition.StringRepresentation = string(jsonBody)
	node.Item = condition
	return nil
}

// Takes the conditions array from the audience in the datafile and turns it into a condition tree
func buildAudienceConditionTree(conditions interface{}) (conditionTree *entities.TreeNode, err error) {

	value := reflect.ValueOf(conditions)
	visited := make(map[interface{}]bool)

	conditionTree = &entities.TreeNode{Operator: "or"}
	var populateConditions func(v reflect.Value, root *entities.TreeNode)
	populateConditions = func(v reflect.Value, root *entities.TreeNode) {

		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if v.Kind() == reflect.Ptr {
				// Check for recursive data
				if visited[v.Interface()] {
					return
				}
				visited[v.Interface()] = true
			}
			v = v.Elem()
		}

		switch v.Kind() {

		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				n := &entities.TreeNode{}
				typedV := v.Index(i).Interface()
				if value, ok := typedV.(string); ok {
					if isValidOperator(value) {
						n.Operator = typedV.(string)
						root.Operator = n.Operator
						continue
					} else {
						n.Item = value

					}
				}

				root.Nodes = append(root.Nodes, n)

				populateConditions(v.Index(i), n)
			}
		}
	}

	populateConditions(value, conditionTree)

	if conditionTree.Nodes == nil && conditionTree.Operator == "" {
		err = errEmptyTree
		conditionTree = nil
	}

	return conditionTree, err
}

func isValidOperator(op string) bool {
	operator := OperatorType(op)
	switch operator {
	case And, Or, Not:
		return true
	}
	return false
}

// Extracts odpSegment from leaf node and adds it to odpSegments array
func extractSegment(odpSegments *[]string, node *entities.TreeNode) {
	condition, ok := node.Item.(entities.Condition)
	if !ok {
		return
	}
	// Add segment to list only if match type is qualified and value is a non-empty string
	if condition.Match == matchers.QualifiedMatchType {
		if segment, ok := condition.Value.(string); ok && segment != "" {
			*odpSegments = append(*odpSegments, segment)
		}
	}
}
