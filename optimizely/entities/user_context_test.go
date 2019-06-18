package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAttributesGetString(t *testing.T) {
	userAttributes := UserAttributes{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}

	// Test happy path
	stringAttribute, _ := userAttributes.GetString("string_foo")
	assert.Equal(t, "foo", stringAttribute)

	// Test non-existent attr name
	_, err := userAttributes.GetString("string_bar")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No string attribute named "string_bar"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	// Test non-string attribute
	_, err = userAttributes.GetString("bool_true")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No string attribute named "bool_true"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}

func TestUserAttributesGetBool(t *testing.T) {
	userAttributes := UserAttributes{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}

	// Test happy path
	boolAttribute, _ := userAttributes.GetBool("bool_true")
	assert.Equal(t, true, boolAttribute)

	// Test non-existent attr name
	_, err := userAttributes.GetBool("bool_false")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No bool attribute named "bool_false"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	_, err = userAttributes.GetBool("string_foo")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No bool attribute named "string_foo"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}

func TestUserAttributesGetFloat(t *testing.T) {
	userAttributes := UserAttributes{
		Attributes: map[string]interface{}{
			"int_42":    42,
			"float_4_2": 42.0,
		},
	}

	// Test happy path
	floatAttribute1, _ := userAttributes.GetFloat("int_42")
	floatAttribute2, _ := userAttributes.GetFloat("float_4_2")
	assert.Equal(t, 42.0, floatAttribute1)
	assert.Equal(t, 42.0, floatAttribute2)

	// Test non-existent attr name
	_, err := userAttributes.GetFloat("bool_false")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No float attribute named "bool_false"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	_, err = userAttributes.GetFloat("string_foo")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `No float attribute named "string_foo"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}
