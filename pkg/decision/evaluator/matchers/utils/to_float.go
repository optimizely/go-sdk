// Package utils //
package utils

import "reflect"

// ToFloat attempts to convert the given value to a float
func ToFloat(value interface{}) (float64, bool) {

	if value == nil {
		return 0, false
	}
	var floatType = reflect.TypeOf(float64(0))
	v := reflect.ValueOf(value)
	v = reflect.Indirect(v)

	if v.Type().String() == "float64" || v.Type().ConvertibleTo(floatType) {
		floatValue := v.Convert(floatType).Float()
		return floatValue, true

	}
	return 0, false
}
