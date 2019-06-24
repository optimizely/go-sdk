package logging

var loggerInstance OptimizelyLogger

const (
	_ = iota

	// LogLevelError log level
	LogLevelError

	// LogLevelWarning log level
	LogLevelWarning

	// LogLevelInfo log level
	LogLevelInfo

	// LogLevelDebug log level
	LogLevelDebug
)

func init() {
	loggerInstance = NewStdoutFilteredLevelLogger(LogLevelInfo)
}

// SetLogger replaces the default logger with the given logger
func SetLogger(logger OptimizelyLogger) {
	loggerInstance = logger
}

// SetLogLevel sets the log level to the given level
func SetLogLevel(logLevel int) {
	loggerInstance.SetLogLevel(logLevel)
}

// Info logs the given message with a INFO level
func Info(message string) {
	loggerInstance.Log(LogLevelInfo, message)
}

// Debug logs the given message with a DEBUG level
func Debug(message string) {
	loggerInstance.Log(LogLevelDebug, message)
}

// Error logs the given message with a ERROR level
func Error(message string) {
	loggerInstance.Log(LogLevelError, message)
}
