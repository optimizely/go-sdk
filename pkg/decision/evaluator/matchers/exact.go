/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                   *
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
	"github.com/optimizely/go-sdk/pkg/logging"
)

// ExactMatcher matches against the "exact" match type
type ExactMatcher struct {
	Condition entities.Condition
	Logger    logging.OptimizelyLogProducer
}

// Match returns true if the user's attribute match the condition's string value
func (m ExactMatcher) Match(user entities.UserContext) (bool, error) {

	if !user.CheckAttributeExists(m.Condition.Name) {
		m.Logger.Debug(fmt.Sprintf(logging.NullUserAttribute.String(), m.Condition.StringRepresentation, m.Condition.Name))
		return false, fmt.Errorf(`no attribute named "%s"`, m.Condition.Name)
	}

	if stringValue, ok := m.Condition.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(m.Condition.Name)
		if err != nil {
			val, _ := user.GetAttribute(m.Condition.Name)
			m.Logger.Warning(fmt.Sprintf(logging.InvalidAttributeValueType.String(), m.Condition.StringRepresentation, val, m.Condition.Name))
			return false, err
		}
		return stringValue == attributeValue, nil
	}

	if boolValue, ok := m.Condition.Value.(bool); ok {
		attributeValue, err := user.GetBoolAttribute(m.Condition.Name)
		if err != nil {
			val, _ := user.GetAttribute(m.Condition.Name)
			m.Logger.Warning(fmt.Sprintf(logging.InvalidAttributeValueType.String(), m.Condition.StringRepresentation, val, m.Condition.Name))
			return false, err
		}
		return boolValue == attributeValue, nil
	}

	if floatValue, ok := utils.ToFloat(m.Condition.Value); ok {
		attributeValue, err := user.GetFloatAttribute(m.Condition.Name)
		if err != nil {
			val, _ := user.GetAttribute(m.Condition.Name)
			m.Logger.Warning(fmt.Sprintf(logging.InvalidAttributeValueType.String(), m.Condition.StringRepresentation, val, m.Condition.Name))
			return false, err
		}
		return floatValue == attributeValue, nil
	}

	m.Logger.Warning(fmt.Sprintf(logging.UnsupportedConditionValue.String(), m.Condition.StringRepresentation))
	return false, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", m.Condition.Name)
}
