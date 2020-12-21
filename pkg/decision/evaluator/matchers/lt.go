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
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// LtMatcher matches against the "lt" match type
func LtMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, decide.DecisionReasons, error) {
	reasons := decide.NewDecisionReasons(nil)
	res, decideReasons, err := compare(condition, user, logger)
	reasons.Append(decideReasons)
	if err != nil {
		return false, reasons, err
	}
	return res < 0, reasons, nil
}
