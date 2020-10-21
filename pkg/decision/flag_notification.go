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

// Package decision //
package decision

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// FlagNotification constructs default flag notification
func FlagNotification(flagKey, variationKey, ruleKey string, enabled, decisionEventDispatched bool, userContext *entities.UserContext, variables map[string]interface{}, reasons []string) *notification.DecisionNotification {

	if flagKey == "" {
		return nil
	}

	decisionInfo := map[string]interface{}{
		"flagKey":                 flagKey,
		"enabled":                 enabled,
		"variables":               variables,
		"variationKey":            variationKey,
		"ruleKey":                 ruleKey,
		"reasons":                 reasons,
		"decisionEventDispatched": decisionEventDispatched,
	}

	decisionInfo["flagKey"] = flagKey
	decisionInfo["enabled"] = enabled
	decisionInfo["variables"] = variables
	decisionInfo["variationKey"] = variationKey
	decisionInfo["ruleKey"] = ruleKey
	decisionInfo["reasons"] = reasons
	decisionInfo["decisionEventDispatched"] = decisionEventDispatched

	decisionNotification := &notification.DecisionNotification{
		DecisionInfo: decisionInfo,
		Type:         notification.Flag,
		UserContext:  *userContext,
	}
	return decisionNotification
}
