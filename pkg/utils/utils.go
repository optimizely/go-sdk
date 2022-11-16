/****************************************************************************
 * Copyright 2022  Optimizely, Inc. and contributors                        *
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

// Package utils //
package utils

// CompareSlices determines if two string slices are equal
func CompareSlices(a, b []string) bool {
	// Required in case a nil value and an empty array is given
	isFirstValueNil := a == nil
	isSecondValueNil := b == nil

	if !isFirstValueNil && !isSecondValueNil {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}
	return isFirstValueNil && isSecondValueNil
}

// IsValidODPData validates if data has all valid types only (string, integer, float, boolean, and nil),
func IsValidODPData(data map[string]interface{}) bool {
	for _, v := range data {
		if v != nil && !IsValidAttribute(v) {
			return false
		}
	}
	return true
}
