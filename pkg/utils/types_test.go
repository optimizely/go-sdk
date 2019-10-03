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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var trueBool = true
var falseBool = false
var float32bit float32 = math.MaxFloat32
var float64bit = math.MaxFloat64
var int8bit int8 = math.MaxInt8
var int16bit int16 = math.MaxInt16
var int32bit int32 = math.MaxInt32
var int64bit int64 = math.MaxInt64
var intu8bit uint8 = math.MaxUint8
var intu16bit uint16 = math.MaxUint16
var intu32bit uint32 = math.MaxUint32
var intu64bit uint64 = math.MaxUint64
var stringType = "Test string"
var byteType byte = 'Z'

func TestGetBoolValue(t *testing.T) {
	val1, err1 := GetBoolValue(trueBool)
	assert.Equal(t, val1, trueBool)
	assert.Nil(t, err1)
	val2, err2 := GetBoolValue(falseBool)
	assert.Equal(t, val2, falseBool)
	assert.Nil(t, err2)

	val3, err3 := GetBoolValue(float32bit)
	assert.NotNil(t, err3)
	assert.Equal(t, val3, false)
	val4, err4 := GetBoolValue(float64bit)
	assert.NotNil(t, err4)
	assert.Equal(t, val4, false)
	val5, err5 := GetBoolValue(int8bit)
	assert.NotNil(t, err5)
	assert.Equal(t, val5, false)
	val6, err6 := GetBoolValue(int16bit)
	assert.NotNil(t, err6)
	assert.Equal(t, val6, false)
	val7, err7 := GetBoolValue(int32bit)
	assert.NotNil(t, err7)
	assert.Equal(t, val7, false)
	val8, err8 := GetBoolValue(int64bit)
	assert.NotNil(t, err8)
	assert.Equal(t, val8, false)
	val9, err9 := GetBoolValue(intu8bit)
	assert.NotNil(t, err9)
	assert.Equal(t, val9, false)
	val10, err10 := GetBoolValue(intu16bit)
	assert.NotNil(t, err10)
	assert.Equal(t, val10, false)
	val11, err11 := GetBoolValue(intu32bit)
	assert.NotNil(t, err11)
	assert.Equal(t, val11, false)
	val12, err12 := GetBoolValue(intu64bit)
	assert.NotNil(t, err12)
	assert.Equal(t, val12, false)
	val13, err13 := GetBoolValue(stringType)
	assert.NotNil(t, err13)
	assert.Equal(t, val13, false)
	val14, err14 := GetBoolValue(byteType)
	assert.NotNil(t, err14)
	assert.Equal(t, val14, false)
	val15, err15 := GetBoolValue(nil)
	assert.NotNil(t, err15)
	assert.Equal(t, val15, false)
}

func TestGetFloatValue(t *testing.T) {
	val1, err1 := GetFloatValue(trueBool)
	assert.Equal(t, val1, float64(0))
	assert.NotNil(t, err1)
	val2, err2 := GetFloatValue(falseBool)
	assert.Equal(t, val2, float64(0))
	assert.NotNil(t, err2)
	val3, err3 := GetFloatValue(stringType)
	assert.Equal(t, val3, float64(0))
	assert.NotNil(t, err3)
	val4, err4 := GetFloatValue(nil)
	assert.Equal(t, val4, float64(0))
	assert.NotNil(t, err4)

	val5, err5 := GetFloatValue(int8bit)
	assert.NotNil(t, val5)
	assert.Nil(t, err5)
	val6, err6 := GetFloatValue(int16bit)
	assert.NotNil(t, val6)
	assert.Nil(t, err6)
	val7, err7 := GetFloatValue(int32bit)
	assert.NotNil(t, val7)
	assert.Nil(t, err7)
	val8, err8 := GetFloatValue(int64bit)
	assert.NotNil(t, val8)
	assert.Nil(t, err8)
	val9, err9 := GetFloatValue(intu8bit)
	assert.NotNil(t, val9)
	assert.Nil(t, err9)
	val10, err10 := GetFloatValue(intu16bit)
	assert.NotNil(t, val10)
	assert.Nil(t, err10)
	val11, err11 := GetFloatValue(intu32bit)
	assert.NotNil(t, val11)
	assert.Nil(t, err11)
	val12, err12 := GetFloatValue(intu64bit)
	assert.NotNil(t, val12)
	assert.Nil(t, err12)
	val13, err13 := GetFloatValue(byteType)
	assert.NotNil(t, val13)
	assert.Nil(t, err13)
	val14, err14 := GetFloatValue(float32bit)
	assert.NotNil(t, val14)
	assert.Nil(t, err14)
	val15, err15 := GetFloatValue(float64bit)
	assert.NotNil(t, val15)
	assert.Nil(t, err15)
}

func TestGetIntValue(t *testing.T) {
	val1, err1 := GetIntValue(trueBool)
	assert.Equal(t, val1, int64(0))
	assert.NotNil(t, err1)
	val2, err2 := GetIntValue(falseBool)
	assert.Equal(t, val2, int64(0))
	assert.NotNil(t, err2)
	val3, err3 := GetIntValue(stringType)
	assert.Equal(t, val3, int64(0))
	assert.NotNil(t, err3)
	val4, err4 := GetIntValue(nil)
	assert.Equal(t, val4, int64(0))
	assert.NotNil(t, err4)

	val5, err5 := GetIntValue(int8bit)
	assert.NotNil(t, val5)
	assert.Nil(t, err5)
	val6, err6 := GetIntValue(int16bit)
	assert.NotNil(t, val6)
	assert.Nil(t, err6)
	val7, err7 := GetIntValue(int32bit)
	assert.NotNil(t, val7)
	assert.Nil(t, err7)
	val8, err8 := GetIntValue(int64bit)
	assert.NotNil(t, val8)
	assert.Nil(t, err8)
	val9, err9 := GetIntValue(intu8bit)
	assert.NotNil(t, val9)
	assert.Nil(t, err9)
	val10, err10 := GetIntValue(intu16bit)
	assert.NotNil(t, val10)
	assert.Nil(t, err10)
	val11, err11 := GetIntValue(intu32bit)
	assert.NotNil(t, val11)
	assert.Nil(t, err11)
	val12, err12 := GetIntValue(intu64bit)
	assert.NotNil(t, val12)
	assert.Nil(t, err12)
	val13, err13 := GetIntValue(byteType)
	assert.NotNil(t, val13)
	assert.Nil(t, err13)
	val14, err14 := GetIntValue(float32bit)
	assert.NotNil(t, val14)
	assert.Nil(t, err14)
	val15, err15 := GetIntValue(float64bit)
	assert.NotNil(t, val15)
	assert.Nil(t, err15)
}

func TestGetStringValue(t *testing.T) {
	val1, err1 := GetStringValue(stringType)
	assert.Equal(t, val1, stringType)
	assert.Nil(t, err1)

	val2, err2 := GetStringValue(float32bit)
	assert.NotNil(t, err2)
	assert.Equal(t, val2, "")
	val3, err3 := GetStringValue(float64bit)
	assert.NotNil(t, err3)
	assert.Equal(t, val3, "")
	val4, err4 := GetStringValue(int8bit)
	assert.NotNil(t, err4)
	assert.Equal(t, val4, "")
	val5, err5 := GetStringValue(int16bit)
	assert.NotNil(t, err5)
	assert.Equal(t, val5, "")
	val6, err6 := GetStringValue(int32bit)
	assert.NotNil(t, err6)
	assert.Equal(t, val6, "")
	val7, err7 := GetStringValue(int64bit)
	assert.NotNil(t, err7)
	assert.Equal(t, val7, "")
	val8, err8 := GetStringValue(intu8bit)
	assert.NotNil(t, err8)
	assert.Equal(t, val8, "")
	val9, err9 := GetStringValue(intu16bit)
	assert.NotNil(t, err9)
	assert.Equal(t, val9, "")
	val10, err10 := GetStringValue(intu32bit)
	assert.NotNil(t, err10)
	assert.Equal(t, val10, "")
	val11, err11 := GetStringValue(intu64bit)
	assert.NotNil(t, err11)
	assert.Equal(t, val11, "")
	val12, err12 := GetStringValue(byteType)
	assert.NotNil(t, err12)
	assert.Equal(t, val12, "")
	val13, err13 := GetStringValue(trueBool)
	assert.NotNil(t, err13)
	assert.Equal(t, val13, "")
	val14, err14 := GetStringValue(falseBool)
	assert.NotNil(t, err14)
	assert.Equal(t, val14, "")
	val15, err15 := GetStringValue(nil)
	assert.NotNil(t, err15)
	assert.Equal(t, val15, "")
}
