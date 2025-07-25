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

// Package client //
package client

import (
	"github.com/optimizely/go-sdk/v2/pkg/optimizelyjson"
)

// OptimizelyDecision defines the decision returned by decide api.
type OptimizelyDecision struct {
	VariationKey string                         `json:"variationKey"`
	Enabled      bool                           `json:"enabled"`
	Variables    *optimizelyjson.OptimizelyJSON `json:"-"`
	RuleKey      string                         `json:"ruleKey"`
	FlagKey      string                         `json:"flagKey"`
	UserContext  OptimizelyUserContext          `json:"userContext"`
	Reasons      []string                       `json:"reasons"`
	CmabUUID     *string                        `json:"cmabUUID,omitempty"`
}

// NewOptimizelyDecision creates and returns a new instance of OptimizelyDecision
func NewOptimizelyDecision(
	variationKey, ruleKey, flagKey string,
	enabled bool,
	variables *optimizelyjson.OptimizelyJSON,
	userContext OptimizelyUserContext,
	reasons []string,
	cmabUUID *string,
) OptimizelyDecision {
	return OptimizelyDecision{
		VariationKey: variationKey,
		Enabled:      enabled,
		Variables:    variables,
		RuleKey:      ruleKey,
		FlagKey:      flagKey,
		UserContext:  userContext,
		Reasons:      reasons,
		CmabUUID:     cmabUUID,
	}
}

// NewErrorDecision returns a decision with error
func NewErrorDecision(key string, user OptimizelyUserContext, err error) OptimizelyDecision {
	return OptimizelyDecision{
		FlagKey:     key,
		UserContext: user,
		Variables:   optimizelyjson.NewOptimizelyJSONfromMap(map[string]interface{}{}),
		Reasons:     []string{err.Error()},
		CmabUUID:    nil, // CmabUUID is optional and defaults to nil
	}
}
