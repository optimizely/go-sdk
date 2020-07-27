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

// Package entities //
package entities

// OptimizelyDecideOption represents Optimizely decide api options
type OptimizelyDecideOption int

const (
	// DisableTracking when set, will not send decision (impression) event
	DisableTracking OptimizelyDecideOption = iota
	// EnabledOnly when set, will return decisions for enabled-flags only.
	EnabledOnly
	// BypassUPS when set, will bypass UPS (both lookup and save) for decision
	BypassUPS
	// ForExperiment when set, will specify that the key parameter(s) of decide and decideAll APIs should be for experiments (not flags)
	ForExperiment
	// IncludeReasons when set, will return decision debugging messages in reasons
	IncludeReasons
)
