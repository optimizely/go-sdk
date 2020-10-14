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

package client

import (
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
)

// OptimizelyDecision defines a struct that the SDK makes for a flag key and a user context
type OptimizelyDecision struct {
	// The variation key of the decision. This value will be empty when decision making fails.
	VariationKey string
	// The boolean value indicating if the flag is enabled or not.
	Enabled bool
	// The collection of variables assocaited with the decision.
	Variables optimizelyjson.OptimizelyJSON
	// The rule key of the decision.
	RuleKey string
	// The flag key for which the decision has been made for.
	FlagKey string
	// The user context for which the decision has been made for.
	UserContext *OptimizelyUserContext
	// An array of error/info/debug messages describing why the decision has been made.
	Reasons []string
}

// ErrorDecision defines a decision with error
func ErrorDecision(key string, user *OptimizelyUserContext, err error) OptimizelyDecision {
	return OptimizelyDecision{
		FlagKey:     key,
		UserContext: user,
		Reasons:     []string{err.Error()},
	}
}
