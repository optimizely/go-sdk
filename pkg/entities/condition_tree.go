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

import "strings"

// TreeNode in a condition tree
type TreeNode struct {
	Item     interface{} // can be a condition or a string
	Operator string

	Nodes []*TreeNode
}

func _buildString(tn *TreeNode, in string) string {
	if len(tn.Nodes) == 1 {
		return _buildString(tn.Nodes[0], in)
	}
	for _, v := range tn.Nodes {
		if in == "" {
			in += `"` + v.Item.(string) + `"`
		} else {
			in += " " + strings.ToUpper(tn.Operator) + ` "` + v.Item.(string) + `"`
		}
	}
	return in
}

// GetAudienceString returns audience string
func (tn *TreeNode) GetAudienceString() string {
	return _buildString(tn, "")
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
