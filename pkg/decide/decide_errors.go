/****************************************************************************
 * Copyright 2020-2021, Optimizely, Inc. and contributors                   *
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

// Package decide has error definitions for decide api
package decide

import (
	"fmt"
)

type decideMessage string

const (
	// SDKNotReady when sdk is not ready
	SDKNotReady decideMessage = "Optimizely SDK not configured properly yet."
	// FlagKeyInvalid when invalid flag key is provided
	FlagKeyInvalid decideMessage = `No flag was found for key "%s".`
	// VariableValueInvalid when invalid variable value is provided
	VariableValueInvalid decideMessage = `Variable value for key "%s" is invalid or wrong type.`
)

// GetDecideMessage returns message for decide type
func GetDecideMessage(messageType decideMessage, arguments ...interface{}) string {
	return fmt.Sprintf(string(messageType), arguments...)
}

// GetDecideError returns error for decide type
func GetDecideError(messageType decideMessage, arguments ...interface{}) error {
	return fmt.Errorf(string(messageType), arguments...)
}
