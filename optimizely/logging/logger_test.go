package logging

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockOptimizelyLogger struct {
	mock.Mock
}

func (m *MockOptimizelyLogger) Log(level int, message string) {
	m.Called(level, message)
}

func (m *MockOptimizelyLogger) SetLogLevel(level int) {
	m.Called(level)
}

func TestLoggerInfo(t *testing.T) {
	testLogMessage := "Test info message"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelInfo, testLogMessage)

	SetLogger(testLogger)
	Info(testLogMessage)
	testLogger.AssertExpectations(t)
}
