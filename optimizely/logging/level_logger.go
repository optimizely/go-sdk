package logging

import (
	"log"
	"os"
)

// FilteredLevelLogger is an implementation of the OptimizelyLogger that filters by log level
type FilteredLevelLogger struct {
	level  int
	logger log.Logger
}

// Log logs the message if it's log level is lower than the logger's set level
func (l *FilteredLevelLogger) Log(level int, message string) {
	if l.level >= level {
		l.logger.Println(message)
	}
}

// SetLogLevel changes the log level to the given level
func (l *FilteredLevelLogger) SetLogLevel(level int) {
	l.level = level
}

// NewStdoutFilteredLevelLogger returns a new logger that logs to stdout
func NewStdoutFilteredLevelLogger(level int) *FilteredLevelLogger {
	return &FilteredLevelLogger{
		level:  level,
		logger: *log.New(os.Stdout, "[Optimizely]", log.LstdFlags),
	}
}
