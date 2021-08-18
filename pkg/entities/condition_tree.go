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

// import "strings"
import "fmt"

// TreeNode in a condition tree
type TreeNode struct {
	Item     interface{} // can be a condition or a string
	Operator string

	Nodes []*TreeNode
}

func (t *TreeNode) String() string {
	return fmt.Sprintf("type(%T)/ %+v/ %+v/ %+v\n", t.Item, t.Item, t.Operator, len(t.Nodes))
}

func (t *TreeNode) isLeaf() bool {
	return len(t.Nodes) == 0
}

func _mapAudience(s string, m map[string]string) string {
	if m[s] != "" {
		s = m[s]
	}
	return `"` + s + `"`
}

func _buildStringFromStringTree(tn *TreeNode, s string, m map[string]string) string {
	fmt.Println(tn)
	op := tn.Operator
	if len(tn.Nodes) == 0 {
		return s
	}
	if op == "not" {
		s += "NOT "
		if tn.Nodes[0].isLeaf() {
			id := tn.Nodes[0].Item.(string)
			s += _mapAudience(id, m)
		} else {
			s += "("
			s = _buildStringFromStringTree(tn.Nodes[0], s, m)
			s += ")"
		}
		return s
	}
	if !tn.Nodes[0].isLeaf() {
		s += "("
		s = _buildStringFromStringTree(tn.Nodes[0], s, m)
		s += ")"
	} else {
		id := tn.Nodes[0].Item.(string)
		s += _mapAudience(id, m)
	}
	for _, v := range tn.Nodes[1:] {
		if !v.isLeaf() {
			s += " " + strings.ToUpper(op) + " ("
			s = _buildStringFromStringTree(v, s, m)
			s += ")"
			continue
		}
		id := _mapAudience(v.Item.(string), m)
		s += fmt.Sprintf(` %v %v`, strings.ToUpper(op), id)
	}
	return s
}
func _buildString(tn *TreeNode, in string, m map[string]string) string {
	if tn == nil {
		return ""
	}
	if len(tn.Nodes) == 1 {
		in += `["`
		in += tn.Operator
		in += `", `
		in = _buildString(tn.Nodes[0], in, m)
		in += `]`
	} else if len(tn.Nodes) > 1 {
		op := tn.Operator
		id := tn.Nodes[0].Item.(string)
		if m[id] != "" {
			id = `"` + m[id] + `"`
		}
		in += id
		for _, v := range tn.Nodes[1:] {
			id := v.Item.(string)
			if m[id] != "" {
				id = m[id]
			}

			in += fmt.Sprintf(` %v "%v"`, op, id)
		}
	} else if x, ok := tn.Item.(Condition); ok {
		in += x.StringRepresentation
	}
	return in
}

// GetAudienceString returns audience string
func (t *TreeNode) GetAudienceString(m map[string]string, isString bool) string {
	var rt string
	if isString {
		rt = _buildStringFromStringTree(t, "", m)
	} else {
		rt = _buildString(t, "", make(map[string]string))
	}
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
