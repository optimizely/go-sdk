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

// Package matchers //
package matchers

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// SemverGtMatcher checks if the semver string in attributes is greater than the semver string in condition
type SemverGtMatcher struct {
	Condition entities.Condition
}

// Match returns true if the user's attribute match the condition's string value
func (m SemverGtMatcher) Match(user entities.UserContext) (bool, error) {

	comparison, err := SemverEvaluator(&m.Condition, &user)
	if err != nil {
		return false, err
	}
	return comparison > 0, nil
}
