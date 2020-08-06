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

// ExistsMatcher matches against the "exists" match type
type ExistsMatcher struct {
}

// Match returns true if the user's attribute is in the condition
func (m ExistsMatcher) Match(condition entities.Condition, user entities.UserContext) (bool, error) {
	return user.CheckAttributeExists(condition.Name), nil
}
