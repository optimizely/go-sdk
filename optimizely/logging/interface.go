package logging

// OptimizelyLogger is the interface for an Optimizely logger
type OptimizelyLogger interface {
	Log(level int, message string)
	SetLogLevel(logLevel int)
}
