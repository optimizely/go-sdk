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

	"github.com/optimizely/go-sdk/pkg/decision/evaluator/matchers/utils"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// ExactMatcher matches against the "exact" match type
type ExactMatcher struct {
}

// Match returns true if the user's attribute match the condition's string value
func (m ExactMatcher) Match(condition entities.Condition, user entities.UserContext) (bool, error) {
	if stringValue, ok := condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(condition.Name)
		if err != nil {
			return false, err
		}
		return stringValue == attributeValue, nil
	}

	if boolValue, ok := condition.Value.(bool); ok {
		attributeValue, err := user.GetBoolAttribute(condition.Name)
		if err != nil {
			return false, err
		}
		return boolValue == attributeValue, nil
	}

	if floatValue, ok := utils.ToFloat(condition.Value); ok {
		attributeValue, err := user.GetFloatAttribute(condition.Name)
		if err != nil {
			return false, err
		}
		return floatValue == attributeValue, nil
	}

	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", condition.Name)
}
