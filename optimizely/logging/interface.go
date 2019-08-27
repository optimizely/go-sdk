package logging

// OptimizelyLogConsumer consumes log messages produced by the log producers
type OptimizelyLogConsumer interface {
	Log(level LogLevel, message string)
	SetLogLevel(logLevel LogLevel)
}

// OptimizelyLogProducer produces log messages to be consumed by the log consumer
type OptimizelyLogProducer interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string, err interface{})
}
