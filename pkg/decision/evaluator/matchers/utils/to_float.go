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

// Package utils //
package utils

import "github.com/goccy/go-reflect"

// ToFloat attempts to convert the given value to a float
func ToFloat(value interface{}) (float64, bool) {

	if value == nil {
		return 0, false
	}
	var floatType = reflect.TypeOf(float64(0))
	v := reflect.ValueNoEscapeOf(value)
	v = reflect.Indirect(v)

	if v.Type().String() == "float64" || v.Type().ConvertibleTo(floatType) {
		floatValue := v.Convert(floatType).Float()
		return floatValue, true

	}
	return 0, false
}
