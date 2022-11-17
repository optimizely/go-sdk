/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareSlices(t *testing.T) {
	assert.True(t, CompareSlices(nil, nil))
	assert.True(t, CompareSlices([]string{}, []string{}))
	// ordered
	assert.True(t, CompareSlices([]string{"a", "b"}, []string{"a", "b"}))
	// unordered
	assert.True(t, CompareSlices([]string{"a", "c", "b"}, []string{"b", "c", "a"}))
	// unordered with repetition
	assert.True(t, CompareSlices([]string{"a", "c", "c", "b", "a"}, []string{"b", "c", "a", "a", "c"}))
	// unordered with repetition and not equal
	assert.False(t, CompareSlices([]string{"a", "c", "c", "b", "a"}, []string{"b", "c", "a", "a", "c", "b"}))
	// not equal
	assert.False(t, CompareSlices([]string{"a"}, []string{}))
	assert.False(t, CompareSlices([]string{}, []string{"a"}))
	assert.False(t, CompareSlices([]string{"a", "b"}, []string{"a"}))
	// array and nil
	assert.False(t, CompareSlices([]string{}, nil))
	assert.False(t, CompareSlices(nil, []string{}))
}

func TestIsValidODPData(t *testing.T) {

	validData := map[string]interface{}{
		"key11": "value-1",
		"key12": true,
		"key13": 3.5,
		"key14": nil,
		"key15": 1,
	}

	invalidData1 := map[string]interface{}{
		"key11": []string{},
	}
	invalidData2 := map[string]interface{}{
		"key11": map[string]interface{}{},
	}

	assert.True(t, IsValidODPData(validData))
	assert.False(t, IsValidODPData(invalidData1))
	assert.False(t, IsValidODPData(invalidData2))
}
