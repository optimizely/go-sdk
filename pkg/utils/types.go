/****************************************************************************
 * Copyright 2019,2022  Optimizely, Inc. and contributors                   *
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
	"math"
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

// Validates if the type of provided value is numeric.
func isNumericType(value interface{}) (float64, error) {
	switch i := value.(type) {
	case int:
		return float64(i), nil
	case int8:
		return float64(i), nil
	case int16:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case uint8:
		return float64(i), nil
	case uint16:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uintptr:
		return float64(i), nil
	case float32:
		return float64(i), nil
	case float64:
		return i, nil
	default:
		v := reflect.ValueOf(value)
		v = reflect.Indirect(v)
		return math.NaN(), fmt.Errorf("can't convert %v to float64", v.Type())
	}
}

// Validates if the provided value is a valid numeric value.
func isValidNumericValue(value interface{}) bool {
	if floatValue, err := isNumericType(value); err == nil {
		if math.IsNaN(floatValue) {
			return false
		}
		if math.IsInf(floatValue, 1) {
			return false
		}
		if math.IsInf(floatValue, -1) {
			return false
		}
		if math.IsInf(floatValue, 0) {
			return false
		}
		if math.Abs(floatValue) > math.Pow(2, 53) {
			return false
		}
		return true
	}
	return false
}

// IsValidAttribute check if attribute value is valid
func IsValidAttribute(value interface{}) bool {
	if value == nil {
		return false
	}

	switch value.(type) {
	// https://go.dev/tour/basics/11
	case bool, string:
		return true
	default:
		return isValidNumericValue(value)
	}
}
