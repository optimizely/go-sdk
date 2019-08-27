package logging

import (
	"log"
	"os"
)

// FilteredLevelLogConsumer is an implementation of the OptimizelyLogConsumer that filters by log level
type FilteredLevelLogConsumer struct {
	level  LogLevel
	logger *log.Logger
}

// Log logs the message if it's log level is higher than or equal to the logger's set level
func (l *FilteredLevelLogConsumer) Log(level LogLevel, message string) {
	if l.level <= level {
		l.logger.Println(message)
	}
}

// SetLogLevel changes the log level to the given level
func (l *FilteredLevelLogConsumer) SetLogLevel(level LogLevel) {
	l.level = level
}

// NewStdoutFilteredLevelLogConsumer returns a new logger that logs to stdout
func NewStdoutFilteredLevelLogConsumer(level LogLevel) *FilteredLevelLogConsumer {
	return &FilteredLevelLogConsumer{
		level:  level,
		logger: log.New(os.Stdout, "[Optimizely]", log.LstdFlags),
	}
}
