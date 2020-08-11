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
	"fmt"

	"github.com/optimizely/go-sdk/pkg/entities"
)

const (
	// SemverEqMatchType represents match type for semantic version equality comparator
	SemverEqMatchType = "semver_eq"
	// SemverLtMatchType represents match type for semantic version less than comparator
	SemverLtMatchType = "semver_lt"
	// SemverLeMatchType represents match type for semantic version less than or equal to comparator
	SemverLeMatchType = "semver_le"
	// SemverGtMatchType represents match type for semantic version greater than comparator
	SemverGtMatchType = "semver_gt"
	// SemverGeMatchType represents match type for semantic version greater than or equal to comparator
	SemverGeMatchType = "semver_ge"
)

// SemVerEqMatcher matches against the "semver_eq" match type
type SemVerEqMatcher struct {
	Condition entities.Condition
}

// Match returns true if condition value for semantic versions is equal to the target semantic version
func (m SemVerEqMatcher) Match(user entities.UserContext) (bool, error) {
	return match(m.Condition, user)
}

func match(condition entities.Condition, user entities.UserContext) (bool, error) {
	if stringValue, ok := condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(condition.Name)
		if err != nil {
			return false, err
		}
		result, err := entities.SemanticVersion(attributeValue).CompareVersion(entities.SemanticVersion(stringValue))
		if err == nil {
			switch condition.Match {
			case SemverEqMatchType:
				return result == 0, nil
			case SemverLtMatchType:
				return result < 0, nil
			case SemverLeMatchType:
				return result <= 0, nil
			case SemverGtMatchType:
				return result > 0, nil
			case SemverGeMatchType:
				return result >= 0, nil
			}
		}
		return false, err
	}
	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", condition.Name)
}
