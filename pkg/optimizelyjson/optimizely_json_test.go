package optimizelyjson

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type OptimizelyJsonTestSuite struct {
	suite.Suite
	data           map[string]interface{}
	dynamicList    []interface{}
	optimizelyJson *OptimizelyJSON
	payload        string
}

func (suite *OptimizelyJsonTestSuite) SetupTest() {

	suite.dynamicList = []interface{}{"1", "2", 3.01, 4.23, true}
	suite.payload = `{"field1":1,"field2":2.5,"field3":"three","field4":{"inner_field1":3,"inner_field2":["1","2",3.01,4.23,true]},"field5":true,"field6":null}`

	suite.data = map[string]interface{}{
		"field1": 1.0,
		"field2": 2.5,
		"field3": "three",
		"field4": map[string]interface{}{"inner_field1": 3.0, "inner_field2": suite.dynamicList},
		"field5": true,
		"field6": nil,
	}
	suite.optimizelyJson = NewOptimizelyJSONfromString(suite.payload)
}

func (suite *OptimizelyJsonTestSuite) TestConstructors() {

	object1 := NewOptimizelyJSONfromString(suite.payload)
	object2 := NewOptimizelyJSONfromMap(suite.data)

	object1ToString, err1Str := object1.ToString()
	object2ToString, err2Str := object2.ToString()
	suite.NoError(err1Str)
	suite.NoError(err2Str)
	suite.Equal(object1ToString, object2ToString)

	object1ToMap, err1Map := object1.ToMap()
	object2ToMap, err2Map := object2.ToMap()
	suite.NoError(err1Map)
	suite.NoError(err2Map)
	suite.Equal(object1ToMap, object2ToMap)

}

func (suite *OptimizelyJsonTestSuite) TestToMap() {

	returnValue, err := suite.optimizelyJson.ToMap()
	suite.NoError(err)
	suite.Equal(suite.data, returnValue)
}

func (suite *OptimizelyJsonTestSuite) TestToString() {

	returnValue, err := suite.optimizelyJson.ToString()
	suite.NoError(err)
	expected := `{"field1":1,"field2":2.5,"field3":"three","field4":{"inner_field1":3,"inner_field2":["1","2",3.01,4.23,true]},"field5":true,"field6":null}`
	suite.Equal(expected, returnValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueInvalidJsonKeyEmptySchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("some_key", &emptyStruct)
	suite.Error(err)
	suite.Equal(`json key "some_key" not found`, err.Error())
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueInvalidJsonMultipleKeyEmptySchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field3.some_key", &emptyStruct)
	suite.Error(err)
	suite.Equal(`json key "some_key" not found`, err.Error())
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyEmptySchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field4", &emptyStruct)
	suite.NoError(err)
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonMultipleKeyWrongSchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field4.inner_field1", &emptyStruct)
	suite.Error(err) // cannot unmarshal number into a struct
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonMultipleKeyValidSchema() {
	var intValue int
	err := suite.optimizelyJson.GetValue("field4.inner_field1", &intValue)
	suite.NoError(err)
	suite.Equal(3, intValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonMultipleKeyValidGenericSchema() {
	var value interface{}
	err := suite.optimizelyJson.GetValue("field4.inner_field2", &value)
	suite.NoError(err)
	suite.Equal(suite.dynamicList, value)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyIntValue() {
	var intValue int
	err := suite.optimizelyJson.GetValue("field1", &intValue)
	suite.NoError(err)
	suite.Equal(1, intValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyDoubleValue() {
	var doubleValue float64
	err := suite.optimizelyJson.GetValue("field2", &doubleValue)
	suite.NoError(err)
	suite.Equal(2.5, doubleValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyStringValue() {
	var stringValue string
	err := suite.optimizelyJson.GetValue("field3", &stringValue)
	suite.NoError(err)
	suite.Equal("three", stringValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyBoolValue() {
	var boolValue bool
	err := suite.optimizelyJson.GetValue("field5", &boolValue)
	suite.NoError(err)
	suite.Equal(true, boolValue)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyNullValue() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field6", &emptyStruct)
	suite.NoError(err)
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueInValidJsonKey() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field4.", &emptyStruct)
	suite.Error(err)
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueEmptyJsonKeyEmptySchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("", &emptyStruct)
	suite.NoError(err)
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueEmptyJsonMultipleKeyEmptySchema() {
	emptyStruct := struct{}{}
	err := suite.optimizelyJson.GetValue("field4..some_field", &emptyStruct)
	suite.Error(err)
	suite.Equal(struct{}{}, emptyStruct)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueEmptyJsonKeyWholeSchema() {

	type field4Struct struct {
		InnerField1 int           `json:"inner_field1"`
		InnerField2 []interface{} `json:"inner_field2"`
	}

	type schema struct {
		Field1 int
		Field2 float64
		Field3 string
		Field4 field4Struct
		Field5 bool
		Field6 interface{}
	}
	sc := schema{}
	err := suite.optimizelyJson.GetValue("", &sc)
	suite.NoError(err)

	expected := schema{
		Field1: 1,
		Field2: 2.5,
		Field3: "three",
		Field4: field4Struct{InnerField1: 3, InnerField2: suite.dynamicList},
		Field5: true,
		Field6: nil,
	}
	suite.Equal(expected, sc)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyPartialSchema() {

	type schema struct {
		InnerField1 int           `json:"inner_field1"`
		InnerField2 []interface{} `json:"inner_field2"`
	}

	sc := schema{}
	err := suite.optimizelyJson.GetValue("field4", &sc)
	suite.NoError(err)

	expected := schema{
		InnerField1: 3,
		InnerField2: suite.dynamicList,
	}
	suite.Equal(expected, sc)

	// check if it does not destroy original object
	err = suite.optimizelyJson.GetValue("field4", &sc)
	suite.NoError(err)
	suite.Equal(expected, sc)
}

func (suite *OptimizelyJsonTestSuite) TestGetValueValidJsonKeyArraySchema() {

	var array []interface{}

	err := suite.optimizelyJson.GetValue("field4.inner_field2", &array)
	suite.NoError(err)

	suite.Equal("1", array[0])
	suite.Equal("2", array[1])
	suite.Equal(3.01, array[2])
	suite.Equal(4.23, array[3])

}

func TestOptimizelyJsonTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyJsonTestSuite))
}
