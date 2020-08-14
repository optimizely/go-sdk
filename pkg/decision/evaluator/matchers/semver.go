/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
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

	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/optimizely/go-sdk/pkg/entities"

	"github.com/pkg/errors"
)

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
			return -1, nil
		case !sv.isNumber(versionParts[idx]):
			// Compare strings
			if versionParts[idx] < targetedVersionParts[idx] {
				return -1, nil
			} else if versionParts[idx] > targetedVersionParts[idx] {
				return 1, nil
			}
		case sv.isNumber(targetedVersionParts[idx]): // both targetedVersionParts and versionParts are digits
			if sv.toInt(versionParts[idx]) < sv.toInt(targetedVersionParts[idx]) {
				return -1, nil
			} else if sv.toInt(versionParts[idx]) > sv.toInt(targetedVersionParts[idx]) {
				return 1, nil
			}
		default:
			return -1, nil
		}
	}

	return 0, nil
}

func (sv SemanticVersion) splitSemanticVersion(targetedVersion string) (parts []string, err error) {

	splitBy := ""
	if sv.isBuild(targetedVersion) {
		splitBy = sv.buildSeperator()
	} else if sv.isPreRelease(targetedVersion) {
		splitBy = sv.preReleaseSeperator()
	}
	targetParts := strings.Split(targetedVersion, splitBy)
	if len(targetParts) == 0 {
		return parts, errors.New(string(reasons.AttributeFormatInvalid))
	}

	targetPrefix := targetParts[0]
	targetSuffix := targetParts[1:]

	// Expect a version string of the form x.y.z
	targetedVersionParts := strings.Split(targetPrefix, ".")

	if len(targetedVersionParts) == 0 {
		return parts, errors.New(string(reasons.AttributeFormatInvalid))
	}

	targetedVersionParts = append(targetedVersionParts, targetSuffix...)
	return targetedVersionParts, nil
}

func (sv SemanticVersion) isNumber(str string) bool {
	var digitCheck = regexp.MustCompile(`^[0-9]+$`)
	return (digitCheck.MatchString(str))
}

func (sv SemanticVersion) toInt(str string) int {
	i, e := strconv.Atoi(str)
	if e != nil {
		return 0
	}
	return i
}

func (sv SemanticVersion) isPreRelease(str string) bool {
	return strings.Contains(str, "-")
}

func (sv SemanticVersion) isBuild(str string) bool {
	return strings.Contains(str, "+")
}

func (sv SemanticVersion) buildSeperator() string {
	return "+"
}

func (sv SemanticVersion) preReleaseSeperator() string {
	return "-"
}

// SemverEvaluator is a help function to wrap a common evaluation code
func SemverEvaluator(cond *entities.Condition, user *entities.UserContext) (int, error) {

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
