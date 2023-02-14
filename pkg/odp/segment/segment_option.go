/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package segment //
package segment

import "errors"

// OptimizelySegmentOption represents options controlling audience segments.
type OptimizelySegmentOption string

const (
	// IgnoreCache ignores cache (save/lookup)
	IgnoreCache OptimizelySegmentOption = "IGNORE_CACHE"
	// ResetCache resets cache
	ResetCache OptimizelySegmentOption = "RESET_CACHE"
)

// Options defines options for controlling audience segments.
type Options struct {
	IgnoreCache bool
	ResetCache  bool
}

// TranslateOptions converts string options array to array of OptimizelySegmentOptions
func TranslateOptions(options []string) ([]OptimizelySegmentOption, error) {
	segmentOptions := []OptimizelySegmentOption{}
	for _, val := range options {
		switch OptimizelySegmentOption(val) {
		case IgnoreCache:
			segmentOptions = append(segmentOptions, IgnoreCache)
		case ResetCache:
			segmentOptions = append(segmentOptions, ResetCache)
		default:
			return []OptimizelySegmentOption{}, errors.New("invalid option: " + val)
		}
	}
	return segmentOptions, nil
}
