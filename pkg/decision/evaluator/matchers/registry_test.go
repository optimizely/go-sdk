package matchers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimizely/go-sdk/pkg/entities"
)

func TestRegister(t *testing.T) {
	expected := func(condition entities.Condition, user entities.UserContext) (bool, error) {
		return false, nil
	}
	Register("test", expected)
	actual := assertMatcher(t, "test")
	matches, err := actual(entities.Condition{}, entities.UserContext{})
	assert.False(t, matches)
	assert.NoError(t, err)
}

func TestInit(t *testing.T) {
	assertMatcher(t, ExactMatchType)
	assertMatcher(t, ExistsMatchType)
	assertMatcher(t, LtMatchType)
	assertMatcher(t, GtMatchType)
	assertMatcher(t, SubstringMatchType)
}

func assertMatcher(t *testing.T, name string) Matcher {
	actual, ok := Get(name)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	return actual
}
