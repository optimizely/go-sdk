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

package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTargetString(t *testing.T) {
	target := SemanticVersion("2.0")
	version := SemanticVersion("2.0.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 0, result)
}

func TestTargetFullStringTargetLess(t *testing.T) {
	target := SemanticVersion("2.0.0")
	version := SemanticVersion("2.0.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestTargetFullStringTargetMore(t *testing.T) {
	target := SemanticVersion("2.0.1")
	version := SemanticVersion("2.0.0")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestTargetFullStringTargetEq(t *testing.T) {
	target := SemanticVersion("2.0.0")
	version := SemanticVersion("2.0.0")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 0, result)
}

func TestTargetMajorPartGreater(t *testing.T) {
	target := SemanticVersion("3.0")
	version := SemanticVersion("2.0.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestTargetMajorPartLess(t *testing.T) {
	target := SemanticVersion("2.0")
	version := SemanticVersion("3.0.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestTargetMinorPartGreater(t *testing.T) {
	target := SemanticVersion("2.3")
	version := SemanticVersion("2.0.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestTargetMinorPartLess(t *testing.T) {
	target := SemanticVersion("2.0")
	version := SemanticVersion("2.9.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestTargetMinorPartEqual(t *testing.T) {
	target := SemanticVersion("2.9")
	version := SemanticVersion("2.9.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 0, result)
}

func TestTargetPatchGreater(t *testing.T) {
	target := SemanticVersion("2.3.5")
	version := SemanticVersion("2.3.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestTargetPatchLess(t *testing.T) {
	target := SemanticVersion("2.9.0")
	version := SemanticVersion("2.9.1")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestTargetPatchEqual(t *testing.T) {
	target := SemanticVersion("2.9.9")
	version := SemanticVersion("2.9.9")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 0, result)
}

func TestTargetPatchWithBetaTagEqual(t *testing.T) {
	target := SemanticVersion("2.9.9-beta")
	version := SemanticVersion("2.9.9-beta")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 0, result)
}

func TestPartialVersionEqual(t *testing.T) {
	target := SemanticVersion("2.9.8")
	version := SemanticVersion("2.9")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestBetaTagGreater(t *testing.T) {
	target := SemanticVersion("2.1.2")
	version := SemanticVersion("2.1.3-beta")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestBetaToRelease(t *testing.T) {
	target := SemanticVersion("2.1.2-release")
	version := SemanticVersion("2.1.2-beta")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestReleaseToBeta(t *testing.T) {
	target := SemanticVersion("2.1.2-beta")
	version := SemanticVersion("2.1.2-release")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestTargetWithVersionBetaLess(t *testing.T) {
	target := SemanticVersion("2.1.3")
	version := SemanticVersion("2.1.3-beta")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, -1, result)
}

func TestTargetBetaLess(t *testing.T) {
	target := SemanticVersion("2.1.3-beta")
	version := SemanticVersion("2.1.3")
	result, err := version.CompareVersion(target)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)
}

func TestOtherTests(t *testing.T) {

	targets := []string{"2.1", "2.1", "2", "2"}
	versions := []string{"2.1.0", "2.1.215", "2.12", "2.785.13"}

	for idx, target := range targets {
		result, err := SemanticVersion(versions[idx]).CompareVersion(SemanticVersion(target))
		assert.Nil(t, err)
		assert.Equal(t, 0, result)
	}
}

func TestInvalidAttributes(t *testing.T) {

	target := "2.1.0"
	versions := []string{"-", ".", "..", "+", "+test", " ", "2 .3. 0", "2.", ".2.2", "3.7.2.2"}
	for _, version := range versions {
		_, err := SemanticVersion(version).CompareVersion(SemanticVersion(target))
		assert.Error(t, err)
	}
}
