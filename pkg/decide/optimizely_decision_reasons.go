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

// Package decide //
package decide

import "fmt"

// DecisionReasons defines the reasons for which the decision was made.
type DecisionReasons struct {
	errors, logs   []string
	includeReasons bool
}

// NewDecisionReasons returns a new instance of OptimizelyDecisionReasons.
func NewDecisionReasons(options []Options) DecisionReasons {
	includeReasons := false
	for option := range options {
		if option == int(IncludeReasons) {
			includeReasons = true
			break
		}
	}
	return DecisionReasons{
		errors:         []string{},
		logs:           []string{},
		includeReasons: includeReasons,
	}
}

// AddError appends given message to the error list.
func (o *DecisionReasons) AddError(message string) {
	o.errors = append(o.errors, message)
}

// AddInfof appends given info message to the info list.
func (o *DecisionReasons) AddInfof(format string, arguments ...interface{}) string {
	message := fmt.Sprintf(format, arguments...)
	o.logs = append(o.errors, message)
	return message
}

// ToReport returns reasons to be reported.
func (o *DecisionReasons) ToReport() []string {
	reasons := o.errors
	if o.includeReasons {
		reasons = append(reasons, o.logs...)
	}
	return reasons
}
