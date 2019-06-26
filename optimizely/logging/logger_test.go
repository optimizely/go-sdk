package logging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

type MockOptimizelyLogger struct {
	mock.Mock
	loggedMessages []string
}

func (m *MockOptimizelyLogger) Log(level int, message string) {
	m.Called(level, message)
	m.loggedMessages = append(m.loggedMessages, message)
}

func (m *MockOptimizelyLogger) SetLogLevel(level int) {
	m.Called(level)
}

func TestNamedLoggerInfo(t *testing.T) {
	testLogMessage := "Test info message"
	expectedLogMessage := "[test-info] Test info message"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelInfo, expectedLogMessage)

	SetLogger(testLogger)

	logProducer := GetLogger("test-info")
	logProducer.Info(testLogMessage)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{expectedLogMessage}, testLogger.loggedMessages)
}

func TestNamedLoggerError(t *testing.T) {
	testLogMessage := "Test error message"
	expectedLogMessage := "[test-error] Test error message I am an error object"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelError, expectedLogMessage)
	SetLogger(testLogger)

	err := errors.New("I am an error object")
	logProducer := GetLogger("test-error")
	logProducer.Error(testLogMessage, err)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{expectedLogMessage}, testLogger.loggedMessages)
}
