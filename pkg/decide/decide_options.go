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

// Package decide has option definitions for decide api
package decide

// OptimizelyDecideOptions controlling flag decisions.
type OptimizelyDecideOptions int

const (
	// DisableDecisionEvent when set, disables decision event tracking.
	DisableDecisionEvent OptimizelyDecideOptions = 1 + iota
	// EnabledFlagsOnly when set, returns decisions only for flags which are enabled.
	EnabledFlagsOnly
	// IgnoreUserProfileService when set, skips user profile service for decision.
	IgnoreUserProfileService
	// IncludeReasons when set, includes info and debug messages in the decision reasons.
	IncludeReasons
	// ExcludeVariables when set, excludes variable values from the decision result.
	ExcludeVariables
)

// Options defines options for controlling flag decisions.
type Options struct {
	DisableDecisionEvent     bool
	EnabledFlagsOnly         bool
	IgnoreUserProfileService bool
	IncludeReasons           bool
	ExcludeVariables         bool
}
