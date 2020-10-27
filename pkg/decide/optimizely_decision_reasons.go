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
	"sync"
)

// DecisionReasons defines the reasons for which the decision was made.
type DecisionReasons struct {
	errors, logs   []string
	includeReasons bool
	mutex          *sync.RWMutex
}

// NewDecisionReasons returns a new instance of DecisionReasons.
func NewDecisionReasons(options OptimizelyDecideOptions) DecisionReasons {
	return DecisionReasons{
		errors:         []string{},
		logs:           []string{},
		includeReasons: options.IncludeReasons,
		mutex:          new(sync.RWMutex),
	}
}

// AddError appends given message to the error list.
func (o *DecisionReasons) AddError(message string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.errors = append(o.errors, message)
}

// AddInfof appends given info message to the info list after formatting.
func (o *DecisionReasons) AddInfof(format string, arguments ...interface{}) string {
	message := fmt.Sprintf(format, arguments...)
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.logs = append(o.logs, message)
	return message
}

// ToReport returns reasons to be reported.
func (o DecisionReasons) ToReport() []string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	reasons := o.errors
	if o.includeReasons {
		reasons = append(reasons, o.logs...)
	}
	return reasons
}
