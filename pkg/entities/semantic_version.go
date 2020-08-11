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
	"strconv"
	"strings"
)

// SemanticVersion represents semantic version
type SemanticVersion string

// CompareVersion compares two semantic versions
func (s SemanticVersion) CompareVersion(targetedVersion SemanticVersion) (val int, err error) {

	if string(targetedVersion) == "" {
		// Any version.
		return 0, nil
	}

	targetedVersionParts, err := targetedVersion.splitSemanticVersion()
	if err != nil {
		return val, err
	}

	versionParts, err := s.splitSemanticVersion()
	if err != nil {
		return val, err
	}

	// Up to the precision of targetedVersion, expect version to match exactly.
	for idx := range targetedVersionParts {
		if len(versionParts) <= idx {
			// even if they are equal at this point. if the target is a prerelease then it must be greater than the pre release.
			if targetedVersion.isPreRelease() {
				return 1, nil
			}
			return -1, nil
		}

		if !SemanticVersion(versionParts[idx]).isNumber() {
			// Compare strings
			if versionParts[idx] < targetedVersionParts[idx] {
				return -1, nil
			} else if versionParts[idx] > targetedVersionParts[idx] {
				return 1, nil
			}
		} else if part, err := strconv.Atoi(versionParts[idx]); err == nil {
			if target, err := strconv.Atoi(targetedVersionParts[idx]); err == nil {
				if part < target {
					return -1, nil
				} else if part > target {
					return 1, nil
				}
			}
		} else {
			return -1, nil
		}
	}
	if s.isPreRelease() && !targetedVersion.isPreRelease() {
		return -1, nil
	}
	return 0, nil
}

func (s SemanticVersion) splitSemanticVersion() (targetedVersionParts []string, err error) {

	var targetParts []string
	var targetPrefix = string(s)
	var targetSuffix []string
	invalidAttributesError := errors.New("provided attributes are in an invalid format")

	if s.hasWhiteSpace() {
		return targetedVersionParts, invalidAttributesError
	}

	if s.isPreRelease() || s.isBuild() {
		if s.isPreRelease() {
			targetParts = strings.Split(targetPrefix, s.preReleaseSeperator())
		} else {
			targetParts = strings.Split(targetPrefix, s.buildSeperator())
		}
		targetParts = s.deleteEmpty(targetParts)
		if len(targetParts) <= 1 {
			return targetedVersionParts, invalidAttributesError
		}
		targetPrefix = targetParts[0]
		targetSuffix = targetParts[1:]
	}

	// Expect a version string of the form x.y.z
	dotCount := strings.Count(targetPrefix, ".")
	if dotCount > 2 {
		return targetedVersionParts, invalidAttributesError
	}
	targetedVersionParts = strings.Split(targetPrefix, ".")
	targetedVersionParts = s.deleteEmpty(targetedVersionParts)

	if len(targetedVersionParts) != dotCount+1 {
		return []string{}, invalidAttributesError
	}
	targetedVersionParts = append(targetedVersionParts, targetSuffix...)

	return targetedVersionParts, nil
}

func (s SemanticVersion) hasWhiteSpace() bool {
	return strings.Contains(string(s), " ")
}

func (s SemanticVersion) isNumber() bool {
	_, err := strconv.Atoi(string(s))
	return err == nil
}

func (s SemanticVersion) isPreRelease() bool {
	return strings.Contains(string(s), s.preReleaseSeperator())
}

func (s SemanticVersion) isBuild() bool {
	return strings.Contains(string(s), s.buildSeperator())
}

func (s SemanticVersion) buildSeperator() string {
	return "+"
}

func (s SemanticVersion) preReleaseSeperator() string {
	return "-"
}

func (s SemanticVersion) deleteEmpty(arr []string) []string {
	var finalArray []string
	for _, str := range arr {
		if str != "" {
			finalArray = append(finalArray, str)
		}
	}
	return finalArray
}
