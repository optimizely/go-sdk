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

import (
	"fmt"
	"reflect"
)

var floatType = reflect.TypeOf(float64(0))
var intType = reflect.TypeOf(int64(0))

// GetBoolValue will attempt to convert the given value to a bool
func GetBoolValue(value interface{}) (bool, error) {
	if value != nil {
		v := reflect.ValueOf(value)
		if v.Type().String() == "bool" {
			return v.Bool(), nil
		}
	}

	return false, fmt.Errorf(`value "%v" could not be converted to bool`, value)
}

// GetFloatValue will attempt to convert the given value to a float64
func GetFloatValue(value interface{}) (float64, error) {
	if value != nil {
		v := reflect.ValueOf(value)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		}
	}

	return 0, fmt.Errorf(`value "%v" could not be converted to float`, value)
}

// GetIntValue will attempt to convert the given value to an int64
func GetIntValue(value interface{}) (int64, error) {
	if value != nil {
		v := reflect.ValueOf(value)
		if v.Type().String() == "int64" || v.Type().ConvertibleTo(intType) {
			intValue := v.Convert(intType).Int()
			return intValue, nil
		}
	}

	return 0, fmt.Errorf(`value "%v" could not be converted to int`, value)
}

// GetStringValue will attempt to convert the given value to a string
func GetStringValue(value interface{}) (string, error) {
	if value != nil {
		v := reflect.ValueOf(value)
		if v.Type().String() == "string" {
			return v.String(), nil
		}
	}

	return "", fmt.Errorf(`value "%v" could not be converted to string`, value)
}
