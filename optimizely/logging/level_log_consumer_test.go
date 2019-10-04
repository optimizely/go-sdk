package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdoutFilteredLevelLogConsumer(t *testing.T) {
	newLogger := NewStdoutFilteredLevelLogConsumer(LogLevelInfo)

	assert.Equal(t, newLogger.level, LogLevel(2))
	assert.NotNil(t, newLogger.logger)

	newLogger.SetLogLevel(3)
	assert.Equal(t, newLogger.level, LogLevel(3))

	newLogger.Log(1, "this is hidden")
	newLogger.Log(4, "this is visible")
}
