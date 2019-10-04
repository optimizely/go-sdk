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

// Package mappers ...
package mappers

import (
	"errors"
	"reflect"

	"github.com/json-iterator/go"
	"github.com/optimizely/go-sdk/pkg/entities"
)

var errEmptyTree = errors.New("empty tree")
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Takes the conditions array from the audience in the datafile and turns it into a condition tree
func buildConditionTree(conditions interface{}) (conditionTree *entities.TreeNode, retErr error) {

	var parsedConditions interface{}
	switch v := conditions.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsedConditions); err != nil {
			retErr = err
			return
		}
	default:
		parsedConditions = conditions
	}

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
					jsonBody, err := json.Marshal(typedV)
					if err != nil {
						retErr = err
						return
					}
					condition := entities.Condition{}
					if err := json.Unmarshal(jsonBody, &condition); err != nil {
						retErr = err
						return
					}
					n.Item = condition
				}

				root.Nodes = append(root.Nodes, n)

				populateConditions(v.Index(i), n)
			}
		}
	}

	populateConditions(value, conditionTree)
	if conditionTree.Nodes == nil && conditionTree.Operator == "" {
		retErr = errEmptyTree
		conditionTree = nil
	}
	return conditionTree, retErr
}

// Takes the conditions array from the audience in the datafile and turns it into a condition tree
func buildAudienceConditionTree(conditions interface{}) (conditionTree *entities.TreeNode, err error) {

	var operators = []string{"or", "and", "not"} // any other operators?
	value := reflect.ValueOf(conditions)
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
				if value, ok := typedV.(string); ok {
					if stringInSlice(value, operators) {
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

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
