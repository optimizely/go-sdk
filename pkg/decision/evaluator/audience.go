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

// Package evaluator //
package evaluator

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// AudienceEvaluator evaluates an audience against the given user's attributes
type AudienceEvaluator interface {
	Evaluate(audience entities.Audience, condTreeParams *entities.TreeParameters) bool
}

// TypedAudienceEvaluator evaluates typed audiences
type TypedAudienceEvaluator struct {
	conditionTreeEvaluator TreeEvaluator
}

// NewTypedAudienceEvaluator creates a new instance of the TypedAudienceEvaluator
func NewTypedAudienceEvaluator() *TypedAudienceEvaluator {
	conditionTreeEvaluator := NewMixedTreeEvaluator()
	return &TypedAudienceEvaluator{
		conditionTreeEvaluator: *conditionTreeEvaluator,
	}
}

// Evaluate evaluates the typed audience against the given user's attributes
func (a TypedAudienceEvaluator) Evaluate(audience entities.Audience, condTreeParams *entities.TreeParameters) bool {
	return a.conditionTreeEvaluator.Evaluate(audience.ConditionTree, condTreeParams)
}
