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

package matchers

import (
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/entities"
)

// ExactMatcher matches against the "exact" match type
type ExactMatcher struct{}

// Match returns true if the user's attribute match the condition's string value
func (m ExactMatcher) Match(condition entities.Condition, user entities.UserContext) (bool, error) {
	if stringValue, ok := condition.Value.(string); ok {
		attributeValue, err := user.Attributes.GetString(condition.Name)
		if err != nil {
			return false, err
		}
		return stringValue == attributeValue, nil
	}

	if boolValue, ok := condition.Value.(bool); ok {
		attributeValue, err := user.Attributes.GetBool(condition.Name)
		if err != nil {
			return false, err
		}
		return boolValue == attributeValue, nil
	}

	return false, fmt.Errorf("audience condition %s evaluated to UNKNOWN because the condition value type is not supported", condition.Name)
}
