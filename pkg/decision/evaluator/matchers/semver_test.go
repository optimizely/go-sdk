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

package matchers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/pkg/entities"
)

type test struct {
	name      string
	version   interface{}
	attribute string

	validBool bool
	result    bool
}

func TestSemverEqMatcher(t *testing.T) {

	condition := entities.Condition{
		Match: "semver_eq",
		Value: "2.0",
		Name:  "version",
	}

	testSuite := []test{
		{name: "equalValues", attribute: "version", version: "2.0.0", result: true, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "2.9", result: false, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "1.9", result: false, validBool: true},
		{name: "attributeNotFound", attribute: "version1", version: "2.0", result: false, validBool: false},
		{name: "invalidType", attribute: "version", version: true, result: false, validBool: false},
		{name: "invalidType", attribute: "version", version: 37, result: false, validBool: false},
	}

	for _, ts := range testSuite {
		user := entities.UserContext{
			Attributes: map[string]interface{}{
				ts.attribute: ts.version,
			},
		}
		result, err := SemverEqMatcher(condition, user)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)
		} else {
			assert.Error(t, err)
		}
	}

}

func TestSemverGeMatcher(t *testing.T) {
	condition := entities.Condition{
		Match: "semver_ge",
		Value: "2.0",
		Name:  "version",
	}

	testSuite := []test{
		{name: "equalValues", attribute: "version", version: "2.0.0", result: true, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "2.9", result: true, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "1.9", result: false, validBool: true},
		{name: "attributeNotFound", attribute: "version1", version: "2.0", result: false, validBool: false},
	}

	for _, ts := range testSuite {
		user := entities.UserContext{
			Attributes: map[string]interface{}{
				ts.attribute: ts.version,
			},
		}
		result, err := SemverGeMatcher(condition, user)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)
		} else {
			assert.Error(t, err)
		}
	}

}

func TestSemverGtMatcher(t *testing.T) {
	condition := entities.Condition{
		Match: "semver_gt",
		Value: "2.0",
		Name:  "version",
	}

	testSuite := []test{
		{name: "equalValues", attribute: "version", version: "2.0.0", result: false, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "2.9", result: true, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "1.9", result: false, validBool: true},
		{name: "attributeNotFound", attribute: "version1", version: "2.0", result: false, validBool: false},
	}

	for _, ts := range testSuite {
		user := entities.UserContext{
			Attributes: map[string]interface{}{
				ts.attribute: ts.version,
			},
		}
		result, err := SemverGtMatcher(condition, user)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestSemverLeMatcher(t *testing.T) {
	condition := entities.Condition{
		Match: "semver_le",
		Value: "2.0",
		Name:  "version",
	}

	testSuite := []test{
		{name: "equalValues", attribute: "version", version: "2.0.0", result: true, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "2.9", result: false, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "1.9", result: true, validBool: true},
		{name: "attributeNotFound", attribute: "version1", version: "2.0", result: false, validBool: false},
	}

	for _, ts := range testSuite {
		user := entities.UserContext{
			Attributes: map[string]interface{}{
				ts.attribute: ts.version,
			},
		}
		result, err := SemverLeMatcher(condition, user)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestSemverLtMatcher(t *testing.T) {
	condition := entities.Condition{
		Match: "semver_lt",
		Value: "2.0",
		Name:  "version",
	}

	testSuite := []test{
		{name: "equalValues", attribute: "version", version: "2.0.0", result: false, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "2.9", result: false, validBool: true},
		{name: "notEqualValues", attribute: "version", version: "1.9", result: true, validBool: true},
		{name: "attributeNotFound", attribute: "version1", version: "2.0", result: false, validBool: false},
	}

	for _, ts := range testSuite {
		user := entities.UserContext{
			Attributes: map[string]interface{}{
				ts.attribute: ts.version,
			},
		}
		result, err := SemverLtMatcher(condition, user)
		if ts.validBool {
			assert.NoError(t, err)
			assert.Equal(t, ts.result, result)
		} else {
			assert.Error(t, err)
		}
	}
}
