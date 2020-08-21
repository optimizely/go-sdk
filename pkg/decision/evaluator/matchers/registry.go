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
	"github.com/optimizely/go-sdk/pkg/logging"
)

// Matcher type is used to evaluate audience conditional primitives
type Matcher func(entities.Condition, entities.UserContext, logging.OptimizelyLogProducer) (bool, error)

const (
	// ExactMatchType name for the "exact" matcher
	ExactMatchType = "exact"
	// ExistsMatchType name for the "exists" matcher
	ExistsMatchType = "exists"
	// LtMatchType name for the "lt" matcher
	LtMatchType = "lt"
	// LeMatchType name for the "le" matcher
	LeMatchType = "le"
	// GtMatchType name for the "gt" matcher
	GtMatchType = "gt"
	// GeMatchType name for the "ge" matcher
	GeMatchType = "ge"
	// SubstringMatchType name for the "substring" matcher
	SubstringMatchType = "substring"
	// SemverEqMatchType name for the semver_eq matcher
	SemverEqMatchType = "semver_eq"
	// SemverLtMatchType name for the semver_eq matcher
	SemverLtMatchType = "semver_lt"
	// SemverLeMatchType name for the semver_eq matcher
	SemverLeMatchType = "semver_le"
	// SemverGtMatchType name for the semver_eq matcher
	SemverGtMatchType = "semver_gt"
	// SemverGeMatchType name for the semver_eq matcher
	SemverGeMatchType = "semver_ge"
)

var registry = map[string]Matcher{
	ExactMatchType:     ExactMatcher,
	ExistsMatchType:    ExistsMatcher,
	LtMatchType:        LtMatcher,
	LeMatchType:        LeMatcher,
	GtMatchType:        GtMatcher,
	GeMatchType:        GeMatcher,
	SubstringMatchType: SubstringMatcher,
	SemverEqMatchType:  SemverEqMatcher,
	SemverLtMatchType:  SemverLtMatcher,
	SemverLeMatchType:  SemverLeMatcher,
	SemverGtMatchType:  SemverGtMatcher,
	SemverGeMatchType:  SemverGeMatcher,
}

var lock = sync.RWMutex{}

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
