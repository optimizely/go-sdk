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

// SemVerLtMatcher matches against the "semver_lt" match type
type SemVerLtMatcher struct {
	Condition entities.Condition
}

// Match returns true if condition value for semantic versions is less than target semantic version
func (m SemVerLtMatcher) Match(user entities.UserContext) (bool, error) {

	if stringValue, ok := m.Condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(m.Condition.Name)
		if err != nil {
			return false, err
		}
		result, err := entities.SemanticVersion(attributeValue).CompareVersion(entities.SemanticVersion(stringValue))
		if err == nil {
			return result < 0, nil
		}
		return false, err
	}

	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", m.Condition.Name)
}
