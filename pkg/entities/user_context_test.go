package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAttributesGetStringAttribute(t *testing.T) {
	userContext := UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}

	// Test happy path
	stringAttribute, _ := userContext.GetStringAttribute("string_foo")
	assert.Equal(t, "foo", stringAttribute)

	// Test non-existent attr name
	_, err := userContext.GetStringAttribute("string_bar")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no string attribute named "string_bar"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	// Test non-string attribute
	_, err = userContext.GetStringAttribute("bool_true")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no string attribute named "bool_true"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}

func TestUserAttributesGetBoolAttribute(t *testing.T) {
	userContext := UserContext{
		Attributes: map[string]interface{}{
			"string_foo": "foo",
			"bool_true":  true,
		},
	}

	// Test happy path
	boolAttribute, _ := userContext.GetBoolAttribute("bool_true")
	assert.Equal(t, true, boolAttribute)

	// Test non-existent attr name
	_, err := userContext.GetBoolAttribute("bool_false")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no bool attribute named "bool_false"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	_, err = userContext.GetBoolAttribute("string_foo")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no bool attribute named "string_foo"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}

func TestUserAttributesGetFloatAttribute(t *testing.T) {
	userContext := UserContext{
		Attributes: map[string]interface{}{
			"int_42":    42,
			"float_4_2": 42.0,
		},
	}

	// Test happy path
	floatAttribute1, _ := userContext.GetFloatAttribute("int_42")
	floatAttribute2, _ := userContext.GetFloatAttribute("float_4_2")
	assert.Equal(t, 42.0, floatAttribute1)
	assert.Equal(t, 42.0, floatAttribute2)

	// Test non-existent attr name
	_, err := userContext.GetFloatAttribute("bool_false")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no float attribute named "bool_false"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}

	_, err = userContext.GetFloatAttribute("string_foo")
	if assert.Error(t, err) {
		assert.Equal(t, err.Error(), `no float attribute named "string_foo"`)
	} else {
		assert.Fail(t, "Error should have been thrown")
	}
}
