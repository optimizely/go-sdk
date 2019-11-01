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
	"fmt"
	"strings"

	"github.com/optimizely/go-sdk/pkg/entities"
)

// SubstringMatcher matches against the "substring" match type
type SubstringMatcher struct {
	Condition entities.Condition
}

// Match returns true if the user's attribute is a substring of the condition's string value
func (m SubstringMatcher) Match(user entities.UserContext) (bool, error) {

	if stringValue, ok := m.Condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(m.Condition.Name)
		if err != nil {
			return false, nil
		}
		return strings.Contains(attributeValue, stringValue), nil
	}

	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", m.Condition.Name)
}
