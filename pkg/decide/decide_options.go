/****************************************************************************
 * Copyright 2020-2025, Optimizely, Inc. and contributors                   *
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

// Package decide has option definitions and helper methods for decide api
package decide

import "errors"

// OptimizelyDecideOptions controlling flag decisions.
type OptimizelyDecideOptions string

const (
	// DisableDecisionEvent when set, disables decision event tracking.
	DisableDecisionEvent OptimizelyDecideOptions = "DISABLE_DECISION_EVENT"
	// EnabledFlagsOnly when set, returns decisions only for flags which are enabled.
	EnabledFlagsOnly OptimizelyDecideOptions = "ENABLED_FLAGS_ONLY"
	// IgnoreUserProfileService when set, skips user profile service for decision.
	IgnoreUserProfileService OptimizelyDecideOptions = "IGNORE_USER_PROFILE_SERVICE"
	// IncludeReasons when set, includes info and debug messages in the decision reasons.
	IncludeReasons OptimizelyDecideOptions = "INCLUDE_REASONS"
	// ExcludeVariables when set, excludes variable values from the decision result.
	ExcludeVariables OptimizelyDecideOptions = "EXCLUDE_VARIABLES"
	// IgnoreCMABCache instructs the SDK to ignore the CMAB cache and make a fresh request
	IgnoreCMABCache OptimizelyDecideOptions = "IGNORE_CMAB_CACHE"
	// ResetCMABCache instructs the SDK to reset the entire CMAB cache
	ResetCMABCache OptimizelyDecideOptions = "RESET_CMAB_CACHE"
	// InvalidateUserCMABCache instructs the SDK to invalidate CMAB cache entries for the current user
	InvalidateUserCMABCache OptimizelyDecideOptions = "INVALIDATE_USER_CMAB_CACHE"
)

// Options defines options for controlling flag decisions.
type Options struct {
	DisableDecisionEvent     bool
	EnabledFlagsOnly         bool
	IgnoreUserProfileService bool
	IncludeReasons           bool
	ExcludeVariables         bool
	IgnoreCMABCache          bool
	ResetCMABCache           bool
	InvalidateUserCMABCache  bool
}

// TranslateOptions converts string options array to array of OptimizelyDecideOptions
func TranslateOptions(options []string) ([]OptimizelyDecideOptions, error) {
	decideOptions := []OptimizelyDecideOptions{}
	for _, val := range options {
		switch OptimizelyDecideOptions(val) {
		case DisableDecisionEvent:
			decideOptions = append(decideOptions, DisableDecisionEvent)
		case EnabledFlagsOnly:
			decideOptions = append(decideOptions, EnabledFlagsOnly)
		case IgnoreUserProfileService:
			decideOptions = append(decideOptions, IgnoreUserProfileService)
		case ExcludeVariables:
			decideOptions = append(decideOptions, ExcludeVariables)
		case IncludeReasons:
			decideOptions = append(decideOptions, IncludeReasons)
		case IgnoreCMABCache:
			decideOptions = append(decideOptions, IgnoreCMABCache)
		case ResetCMABCache:
			decideOptions = append(decideOptions, ResetCMABCache)
		case InvalidateUserCMABCache:
			decideOptions = append(decideOptions, InvalidateUserCMABCache)
		default:
			return []OptimizelyDecideOptions{}, errors.New("invalid option: " + val)
		}
	}
	return decideOptions, nil
}
