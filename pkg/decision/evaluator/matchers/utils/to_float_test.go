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

package utils

import (
	"testing"

	"math"

	"github.com/stretchr/testify/assert"
)

func TestToFloat(t *testing.T) {

	// Invalid values
	zeroValue := float64(0)

	result, success := ToFloat("abc")
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat("1212")
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat(math.Inf)
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat(nil)
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat(true)
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat([]string{})
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	result, success = ToFloat(map[string]string{})
	assert.Equal(t, zeroValue, result)
	assert.Equal(t, false, success)

	// Valid values
	result, success = ToFloat(121.0)
	assert.Equal(t, float64(121), result)
	assert.Equal(t, true, success)

	result, success = ToFloat(5000)
	assert.Equal(t, float64(5000), result)
	assert.Equal(t, true, success)
}
