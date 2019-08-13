package logging

import "fmt"

// LogLevel represents the level of the log (i.e. Debug, Info, Warning, Error)
type LogLevel int

func (l LogLevel) String() string {
	return [...]string{"Debug", "Info", "Warning", "Error"}[l]
}

var defaultLogConsumer OptimizelyLogConsumer

const (
	// LogLevelDebug log level
	LogLevelDebug LogLevel = iota

	// LogLevelInfo log level
	LogLevelInfo

	// LogLevelWarning log level
	LogLevelWarning

	// LogLevelError log level
	LogLevelError
)

func init() {
	defaultLogConsumer = NewStdoutFilteredLevelLogConsumer(LogLevelInfo)
}

// SetLogger replaces the default logger with the given logger
func SetLogger(logger OptimizelyLogConsumer) {
	defaultLogConsumer = logger
}

// SetLogLevel sets the log level to the given level
func SetLogLevel(logLevel LogLevel) {
	defaultLogConsumer.SetLogLevel(logLevel)
}

// GetLogger returns a log producer with the given name
func GetLogger(name string) OptimizelyLogProducer {
	return NamedLogProducer{
		name: name,
	}
}

// NamedLogProducer produces logs prefixed with its name
type NamedLogProducer struct {
	name string
}

// Debug logs the given message with a DEBUG level
func (p NamedLogProducer) Debug(message string) {
	p.log(LogLevelDebug, message)
}

// Info logs the given message with a INFO level
func (p NamedLogProducer) Info(message string) {
	p.log(LogLevelInfo, message)
}

// Warning logs the given message with a WARNING level
func (p NamedLogProducer) Warning(message string) {
	p.log(LogLevelWarning, message)
}

// Error logs the given message with a ERROR level
func (p NamedLogProducer) Error(message string, err interface{}) {
	if err != nil {
		message = fmt.Sprintf("%s %v", message, err)
	}
	p.log(LogLevelError, message)
}

func (p NamedLogProducer) log(logLevel LogLevel, message string) {

	// prepends the name and log level to the message
	message = fmt.Sprintf("[%s][%s] %s", p.name, logLevel, message)
	defaultLogConsumer.Log(logLevel, message)
}
