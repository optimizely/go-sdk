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

import (
	"fmt"
)

// DefaultDecisionReasons provides the default implementation of DecisionReasons.
type DefaultDecisionReasons struct {
	errors, logs   []string
	includeReasons bool
}

// NewDecisionReasons returns a new instance of DecisionReasons.
func NewDecisionReasons(options *Options) *DefaultDecisionReasons {
	includeReasons := false
	if options != nil {
		includeReasons = options.IncludeReasons
	}
	return &DefaultDecisionReasons{
		errors:         []string{},
		logs:           []string{},
		includeReasons: includeReasons,
	}
}

// AddError appends given message to the error list.
func (o *DefaultDecisionReasons) AddError(format string, arguments ...interface{}) {
	o.errors = append(o.errors, fmt.Sprintf(format, arguments...))
}

// AddInfo appends given info message to the info list after formatting.
func (o *DefaultDecisionReasons) AddInfo(format string, arguments ...interface{}) string {
	message := fmt.Sprintf(format, arguments...)
	if !o.includeReasons {
		return message
	}
	o.logs = append(o.logs, message)
	return message
}

// Append appends given reasons.
func (o *DefaultDecisionReasons) Append(reasons DecisionReasons) {
	if decisionReasons, ok := reasons.(*DefaultDecisionReasons); ok {
		o.errors = append(o.errors, decisionReasons.errors...)
		if o.includeReasons {
			o.logs = append(o.logs, decisionReasons.logs...)
		}
	}
}

// ToReport returns reasons to be reported.
func (o *DefaultDecisionReasons) ToReport() []string {
	reasons := o.errors
	if !o.includeReasons {
		return reasons
	}
	return append(reasons, o.logs...)
}
