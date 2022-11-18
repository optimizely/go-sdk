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

		// create a map of string -> int
		diff := make(map[string]int, len(a))
		for _, _x := range a {
			diff[_x]++
		}
		for _, _y := range b {
			// check if both arrays contain the same elements
			if _, ok := diff[_y]; !ok {
				return false
			}
			// remove count for already checked values
			diff[_y]--
			// delete key from map if count is zero
			if diff[_y] == 0 {
				delete(diff, _y)
			}
		}
		return len(diff) == 0
	}
	return isFirstValueNil && isSecondValueNil
}

// IsValidOdpData validates if data has all valid types only (string, integer, float, boolean, and nil),
func IsValidOdpData(data map[string]interface{}) bool {
	for _, v := range data {
		if v != nil && !IsValidAttribute(v) {
			return false
		}
	}
	return true
}
