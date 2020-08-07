/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package matchers //
package matchers

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/entities"
)

// Matcher type is used to evaluate audience conditional primitives
type Matcher func(entities.Condition, entities.UserContext) (bool, error)

const (
	// ExactMatchType name for the "exact" matcher
	ExactMatchType = "exact"
	// ExistsMatchType name for the "exists" matcher
	ExistsMatchType = "exists"
	// LtMatchType name for the "lt" matcher
	LtMatchType = "lt"
	// GtMatchType name for the "gt" matcher
	GtMatchType = "gt"
	// SubstringMatchType name for the "substring" matcher
	SubstringMatchType = "substring"
)

var registry = make(map[string]Matcher)
var lock = sync.RWMutex{}

func init() {
	Register(ExactMatchType, ExactMatcher)
	Register(ExistsMatchType, ExistsMatcher)
	Register(LtMatchType, LtMatcher)
	Register(GtMatchType, GtMatcher)
	Register(SubstringMatchType, SubstringMatcher)
}

// Register new matchers by providing a name and a Matcher implementation
func Register(name string, matcher Matcher) {
	lock.Lock()
	defer lock.Unlock()

	registry[name] = matcher
}

// Get an implementation of a Matcher function by its registered name
func Get(name string) (Matcher, bool) {
	lock.RLock()
	defer lock.RUnlock()

	matcher, ok := registry[name]
	return matcher, ok
}
