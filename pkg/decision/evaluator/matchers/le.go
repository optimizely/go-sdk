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
	"github.com/optimizely/go-sdk/pkg/entities"
)

// LeMatcher matches against the "le" match type
type LeMatcher struct {
	Condition entities.Condition
}

// Match returns true if the user's attribute is less than or equal to the condition's string value
func (m LeMatcher) Match(user entities.UserContext) (bool, error) {

	var result bool
	var err error

	gtMatcher := GtMatcher(m)

	if result, err = gtMatcher.Match(user); err == nil && result {
		return true, nil
	}

	exactMatcher := ExactMatcher(m)

	if result, err = exactMatcher.Match(user); err == nil && result {
		return true, nil
	}

	return false, err
}
