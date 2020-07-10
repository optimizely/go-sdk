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
	"github.com/optimizely/go-sdk/pkg/logging"
)

// SubstringMatcher matches against the "substring" match type
type SubstringMatcher struct {
	Condition entities.Condition
	Logger    logging.OptimizelyLogProducer
}

// Match returns true if the user's attribute is a substring of the condition's string value
func (m SubstringMatcher) Match(user entities.UserContext) (bool, error) {

	if !user.CheckAttributeExists(m.Condition.Name) {
		m.Logger.Debug(fmt.Sprintf(`Audience condition %s evaluated to UNKNOWN because a null value was passed for user attribute "%s".`, m.Condition.StringRepresentation, m.Condition))
		return false, fmt.Errorf(`no attribute named "%s"`, m.Condition.Name)
	}

	if stringValue, ok := m.Condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(m.Condition.Name)
		if err != nil {
			m.Logger.Warning(fmt.Sprintf(`Audience condition "%s" evaluated to UNKNOWN because a value of type "%s" was passed for user attribute "%s".`, m.Condition.StringRepresentation, m.Condition.Value, m.Condition.Name))
			return false, err
		}
		return strings.Contains(attributeValue, stringValue), nil
	}

	m.Logger.Warning(fmt.Sprintf(`Audience condition "%s" has an unsupported condition value. You may need to upgrade to a newer release of the Optimizely SDK.`, m.Condition.StringRepresentation))
	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", m.Condition.Name)
}
