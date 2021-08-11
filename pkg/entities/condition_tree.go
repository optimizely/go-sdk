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

// Package entities //
package entities

// import "strings"
import "fmt"

// TreeNode in a condition tree
type TreeNode struct {
	Item     interface{} // can be a condition or a string
	Operator string

	Nodes []*TreeNode
}

func (t *TreeNode) String() string {
	return fmt.Sprintf("type(%T) %+v %+v %+v\n", t.Item, t.Item, t.Operator, len(t.Nodes))
}

func _buildString(tn *TreeNode, in string) string {
	if len(tn.Nodes) == 1 {
		in += `["`
		in += tn.Operator
		in += `", `
		in = _buildString(tn.Nodes[0], in)
		in += `]"`
	} else if x, ok := tn.Item.(Condition); ok {
		in += x.StringRepresentation
	}
	return in
}

// GetAudienceString returns audience string
func (t *TreeNode) GetAudienceString() string {
	rt := _buildString(t, "")
	return rt
}

// TreeParameters represents parameters of a tree
type TreeParameters struct {
	User        *UserContext
	AudienceMap map[string]Audience
}

// NewTreeParameters returns TreeParameters object
func NewTreeParameters(user *UserContext, audience map[string]Audience) *TreeParameters {
	return &TreeParameters{User: user, AudienceMap: audience}
}
