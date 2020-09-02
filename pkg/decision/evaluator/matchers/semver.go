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

// Package matchers //
package matchers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/optimizely/go-sdk/pkg/logging"

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/pkg/errors"
)

const (
	buildSeperator      = "+"
	preReleaseSeperator = "-"
	whiteSpace          = " "
)

var digitCheck = regexp.MustCompile(`^[0-9]+$`)

// SemanticVersion defines the class
type SemanticVersion struct {
	Condition string // condition is always a string here
}

func (sv SemanticVersion) compareVersion(attribute string) (int, error) {

	targetedVersionParts, err := sv.splitSemanticVersion(sv.Condition)
	if err != nil {
		return 0, err
	}
	versionParts, e := sv.splitSemanticVersion(attribute)
	if e != nil {
		return 0, e
	}

	// Up to the precision of targetedVersion, expect version to match exactly.
	for idx := range targetedVersionParts {

		switch {
		case len(versionParts) <= idx:
			if sv.isPreRelease(attribute) {
				return 1, nil
			}
			return -1, nil
		case !sv.isNumber(versionParts[idx]):
			// Compare strings
			if result := sv.compareVersionStrings(versionParts[idx], targetedVersionParts[idx], attribute); result != 0 {
				return result, nil
			}
		case sv.isNumber(targetedVersionParts[idx]): // both targetedVersionParts and versionParts are digits
			if result := sv.compareVersionNumbers(versionParts[idx], targetedVersionParts[idx]); result != 0 {
				return result, nil
			}
		default:
			return -1, nil
		}
	}

	if sv.isPreRelease(attribute) && !sv.isPreRelease(sv.Condition) {
		return -1, nil
	}

	return 0, nil
}

func (sv SemanticVersion) splitSemanticVersion(targetedVersion string) ([]string, error) {

	if sv.hasWhiteSpace(targetedVersion) {
		return []string{}, errors.New(string(reasons.AttributeFormatInvalid))
	}

	targetPrefix := targetedVersion
	var targetSuffix []string

	if sv.isPreRelease(targetedVersion) || sv.isBuild(targetedVersion) {
		sep := buildSeperator
		if sv.isPreRelease(targetedVersion) {
			sep = preReleaseSeperator
		}
		// this is going to slit with the first occurrence.
		targetParts := strings.SplitN(targetedVersion, sep, 2)
		// in the case it is neither a build or pre release it will return the
		// original string as the first element in the array
		if len(targetParts) == 0 {
			return []string{}, errors.New(string(reasons.AttributeFormatInvalid))
		}

		targetPrefix = targetParts[0]
		targetSuffix = targetParts[1:]

	}

	// Expect a version string of the form x.y.z
	// split all dots with SplitAfter
	targetedVersionParts := strings.Split(targetPrefix, ".")

	if len(targetedVersionParts) > 3 {
		return []string{}, errors.New(string(reasons.AttributeFormatInvalid))
	}

	if len(targetedVersionParts) == 0 {
		return []string{}, errors.New(string(reasons.AttributeFormatInvalid))
	}

	for i := 0; i < len(targetedVersionParts); i++ {
		if !sv.isNumber(targetedVersionParts[i]) {
			return []string{}, errors.New(string(reasons.AttributeFormatInvalid))
		}
	}

	targetedVersionParts = append(targetedVersionParts, targetSuffix...)
	return targetedVersionParts, nil
}

// returns 1 if versionPart > targetedVersionPart, -1 if targetedVersionPart > versionPart, 0 otherwise
func (sv SemanticVersion) compareVersionStrings(versionPart, targetedVersionPart, attribute string) int {
	if versionPart < targetedVersionPart {
		if sv.isPreRelease(sv.Condition) && !sv.isPreRelease(attribute) {
			return 1
		}
		return -1
	} else if versionPart > targetedVersionPart {
		if !sv.isPreRelease(sv.Condition) && sv.isPreRelease(attribute) {
			return -1
		}
		return 1
	}
	return 0
}

// returns 1 if versionPart > targetedVersionPart, -1 if targetedVersionPart > versionPart, 0 otherwise
func (sv SemanticVersion) compareVersionNumbers(versionPart, targetedVersionPart string) int {
	if sv.toInt(versionPart) < sv.toInt(targetedVersionPart) {
		return -1
	} else if sv.toInt(versionPart) > sv.toInt(targetedVersionPart) {
		return 1
	}
	return 0
}

func (sv SemanticVersion) isNumber(str string) bool {
	return digitCheck.MatchString(str)
}

func (sv SemanticVersion) toInt(str string) int {
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return i
}

func (sv SemanticVersion) isPreRelease(str string) bool {
	if prIndex := strings.Index(str, preReleaseSeperator); prIndex > 0 {
		if buildIndex := strings.Index(str, buildSeperator); buildIndex > 0 {
			return prIndex < buildIndex
		}
		return true
	}
	return false
}

func (sv SemanticVersion) isBuild(str string) bool {
	if buildIndex := strings.Index(str, buildSeperator); buildIndex > 0 {
		if prIndex := strings.Index(str, preReleaseSeperator); prIndex > 0 {
			return buildIndex < prIndex
		}
		return true
	}
	return false
}

func (sv SemanticVersion) hasWhiteSpace(str string) bool {
	return str == "" || strings.Contains(str, whiteSpace)
}

// SemverEvaluator is a help function to wrap a common evaluation code
func SemverEvaluator(cond entities.Condition, user entities.UserContext) (int, error) {

	if stringValue, ok := cond.Value.(string); ok {
		attributeValue, err := user.GetStringAttribute(cond.Name)
		if err != nil {
			return 0, err
		}
		semVer := SemanticVersion{stringValue}
		comparison, e := semVer.compareVersion(attributeValue)
		if e != nil {
			return 0, e
		}
		return comparison, nil
	}
	return 0, fmt.Errorf("audience condition %s evaluated to NULL because the condition value type is not supported", cond.Name)
}

// SemverEqMatcher returns true if the user's semver attribute is equal to the semver condition value
func SemverEqMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, error) {
	comparison, err := SemverEvaluator(condition, user)
	if err != nil {
		return false, err
	}
	return comparison == 0, nil
}

// SemverGeMatcher returns true if the user's semver attribute is greater or equal to the semver condition value
func SemverGeMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, error) {
	comparison, err := SemverEvaluator(condition, user)
	if err != nil {
		return false, err
	}
	return comparison >= 0, nil
}

// SemverGtMatcher returns true if the user's semver attribute is greater than the semver condition value
func SemverGtMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, error) {
	comparison, err := SemverEvaluator(condition, user)
	if err != nil {
		return false, err
	}
	return comparison > 0, nil
}

// SemverLeMatcher returns true if the user's semver attribute is less than or equal to the semver condition value
func SemverLeMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, error) {
	comparison, err := SemverEvaluator(condition, user)
	if err != nil {
		return false, err
	}
	return comparison <= 0, nil
}

// SemverLtMatcher returns true if the user's semver attribute is less than the semver condition value
func SemverLtMatcher(condition entities.Condition, user entities.UserContext, logger logging.OptimizelyLogProducer) (bool, error) {
	comparison, err := SemverEvaluator(condition, user)
	if err != nil {
		return false, err
	}
	return comparison < 0, nil
}
