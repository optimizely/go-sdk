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

import (
	"errors"

	"golang.org/x/mod/semver"
)

// SemanticVersion represents semantic version
type SemanticVersion string

// CompareVersion compares two semantic versions
// result will be 0 if v == w, -1 if v < w, or +1 if v > w.
func (s SemanticVersion) CompareVersion(targetedVersion SemanticVersion) (int, error) {

	sVersion := string(s)
	sTargetedVersion := string(targetedVersion)

	if sTargetedVersion == "" {
		// Any version.
		return 0, nil
	}

	// Up to the precision of targetedVersion, expect version to match exactly.
	if semver.IsValid(sVersion) && semver.IsValid(sTargetedVersion) {
		return semver.Compare(sVersion, sTargetedVersion), nil
	}
	return 0, errors.New("provided versions are in an invalid format")
}
