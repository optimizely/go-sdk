/****************************************************************************
 * Copyright 2026, Optimizely, Inc. and contributors                       *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package openfeature

import (
	of "github.com/open-feature/go-sdk/openfeature"
)

// mapReason maps an Optimizely decision state to an OpenFeature Reason.
func mapReason(variationKey string, enabled bool, hasError bool) of.Reason {
	if hasError {
		return of.ErrorReason
	}
	if variationKey == "" {
		return of.DefaultReason
	}
	if enabled {
		return of.TargetingMatchReason
	}
	return of.DisabledReason
}

// makeResolutionError creates an OpenFeature ResolutionError for the given
// error code and message using the appropriate typed constructor.
func makeResolutionError(code of.ErrorCode, msg string) of.ResolutionError {
	switch code {
	case of.FlagNotFoundCode:
		return of.NewFlagNotFoundResolutionError(msg)
	case of.TargetingKeyMissingCode:
		return of.NewTargetingKeyMissingResolutionError(msg)
	case of.TypeMismatchCode:
		return of.NewTypeMismatchResolutionError(msg)
	case of.ParseErrorCode:
		return of.NewParseErrorResolutionError(msg)
	case of.ProviderNotReadyCode:
		return of.NewProviderNotReadyResolutionError(msg)
	default:
		return of.NewGeneralResolutionError(msg)
	}
}
