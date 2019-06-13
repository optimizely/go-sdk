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

package mappers

import (
	"encoding/json"
	"reflect"

	datafileEntities "github.com/optimizely/go-sdk/optimizely/config/datafileProjectConfig/entities"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// MapAudiences maps the raw datafile audience entities to SDK Audience entities
func MapAudiences(audiences []datafileEntities.Audience) map[string]entities.Audience {

	audienceMap := make(map[string]entities.Audience)
	for _, audience := range audiences {
		conditionTree, err := buildConditionTree(audience.Conditions)
		if err != nil {
			// @TODO: handle error
		}
		audienceMap[audience.ID] = entities.Audience{
			ID:            audience.ID,
			Name:          audience.Name,
			ConditionTree: conditionTree,
		}
	}
	return audienceMap
}

// Takes the conditions array from the audience in the datafile and turns it into a condition tree
func buildConditionTree(conditions interface{}) (*entities.ConditionTreeNode, error) {

	value := reflect.ValueOf(conditions)
	visited := make(map[interface{}]bool)
	var retErr error

	conditionTree := &entities.ConditionTreeNode{}
	var populateConditions func(v reflect.Value, root *entities.ConditionTreeNode)
	populateConditions = func(v reflect.Value, root *entities.ConditionTreeNode) {

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
				n := &entities.ConditionTreeNode{}
				typedV := v.Index(i).Interface()
				switch typedV.(type) {
				case string:
					n.Operator = typedV.(string)
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
					n.Condition = condition
				}

				root.Nodes = append(root.Nodes, n)

				populateConditions(v.Index(i), n)
			}
		}
	}

	populateConditions(value, conditionTree)
	return conditionTree, retErr
}
