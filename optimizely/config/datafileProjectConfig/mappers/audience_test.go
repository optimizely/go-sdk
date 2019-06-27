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
	"testing"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
)

func TestBuildConditionTreeSimpleAudienceCondition(t *testing.T) {
	conditionString := "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" } ] ] ]"
	var conditions interface{}
	json.Unmarshal([]byte(conditionString), &conditions)
	conditionTree, err := buildConditionTree(conditions)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	expectedConditionTree := &entities.ConditionTreeNode{
		Operator: "and",
		Nodes: []*entities.ConditionTreeNode{
			&entities.ConditionTreeNode{
				Operator: "or",
				Nodes: []*entities.ConditionTreeNode{
					&entities.ConditionTreeNode{
						Operator: "or",
						Nodes: []*entities.ConditionTreeNode{
							&entities.ConditionTreeNode{
								Condition: entities.Condition{
									Name:  "s_foo",
									Match: "exact",
									Type:  "custom_attribute",
									Value: "foo",
								},
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expectedConditionTree, conditionTree)
}
