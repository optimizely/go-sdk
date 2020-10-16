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

// Package client //
package client

import (
	"github.com/optimizely/go-sdk/pkg/optimizelyjson"
)

// OptimizelyDecision defines a struct that the SDK makes for a flag key and a user context
type OptimizelyDecision struct {
	variationKey string
	enabled      bool
	variables    *optimizelyjson.OptimizelyJSON
	ruleKey      string
	flagKey      string
	userContext  *OptimizelyUserContext
	reasons      []string
}

// NewOptimizelyDecision creates and returns a new instance of OptimizelyDecision
func NewOptimizelyDecision(variationKey, ruleKey, flagKey string, enabled bool, variables *optimizelyjson.OptimizelyJSON, userContext *OptimizelyUserContext, reasons []string) OptimizelyDecision {
	return OptimizelyDecision{
		variationKey: variationKey,
		enabled:      enabled,
		variables:    variables,
		ruleKey:      ruleKey,
		flagKey:      flagKey,
		userContext:  userContext,
		reasons:      reasons,
	}
}

// CreateErrorDecision returns a decision with error
func CreateErrorDecision(key string, user *OptimizelyUserContext, err error) OptimizelyDecision {
	return OptimizelyDecision{
		flagKey:     key,
		userContext: user,
		reasons:     []string{err.Error()},
	}
}

// GetVariationKey returns variation key for optimizely decision.
func (o OptimizelyDecision) GetVariationKey() string {
	return o.variationKey
}

// GetEnabled return the boolean value indicating if the flag is enabled or not.
func (o OptimizelyDecision) GetEnabled() bool {
	return o.enabled
}

// GetVariables returns the collection of variables associated with the decision.
func (o OptimizelyDecision) GetVariables() *optimizelyjson.OptimizelyJSON {
	return o.variables
}

// GetRuleKey returns the rule key of the decision.
func (o OptimizelyDecision) GetRuleKey() string {
	return o.ruleKey
}

// GetFlagKey returns the flag key for which the decision was made.
func (o OptimizelyDecision) GetFlagKey() string {
	return o.flagKey
}

// GetUserContext returns the user context for which the  decision was made.
func (o OptimizelyDecision) GetUserContext() *OptimizelyUserContext {
	return o.userContext
}

// GetReasons returns an array of error/info/debug messages describing why the decision has been made.
func (o OptimizelyDecision) GetReasons() []string {
	return o.reasons
}

// HasFailed returns if variation was not found.
func (o OptimizelyDecision) HasFailed() bool {
	return o.variationKey == ""
}
