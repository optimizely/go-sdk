package matchers

import (
	"github.com/optimizely/go-sdk/pkg/decision/reasons"
	"github.com/pkg/errors"

	"regexp"
	"strconv"
	"strings"
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
		if len(versionParts) <= idx {
			return -1, nil
		} else if !sv.isNumber(versionParts[idx]) {
			//Compare strings
			if versionParts[idx] < targetedVersionParts[idx] {
				return -1, nil
			} else if versionParts[idx] > targetedVersionParts[idx] {
				return 1, nil
			}
		} else if sv.toInt(versionParts[idx]) < sv.toInt(targetedVersionParts[idx]) {
			return -1, nil
		} else if sv.toInt(versionParts[idx]) > sv.toInt(targetedVersionParts[idx]) {
			return 1, nil
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
