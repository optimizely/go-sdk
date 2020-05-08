---
title: "Customize logger"
excerpt: ""
slug: "customize-logger-go"
hidden: false
createdAt: "2019-08-21T21:14:40.407Z"
updatedAt: "2019-08-21T21:15:15.382Z"
---
The **logger** logs information about your experiments to help you with debugging. You can customize where log information is sent and what kind of information is tracked.

The default logger in the Go SDK logs to STDOUT. If you want to disable or customize the logger, you can provide an implementation of the **OptimizelyLogConsumer** interface.

### Custom logger implementation in the SDK

```go
import "github.com/optimizely/go-sdk/pkg/logging"


type CustomLogger struct {
}


func (l *CustomLogger) Log(level logging.LogLevel, message string, fields map[string]interface{}) {

}


func (l *CustomLogger) SetLogLevel(level logging.LogLevel) {

}

customLogger := New(CustomLogger)

logging.SetLogger(customLogger)


```

### Setting the log level

You can also change the default log level from INFO to any of the other log levels.

```go
import "github.com/optimizely/go-sdk/pkg/logging"

// Set log level to Debug
logging.SetLogLevel(logging.LogLevelDebug)
```

### Log levels

The table below lists the log levels for the Go SDK.
| Log Level           | Explanation                                                                                                                                                                                                                       |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **LogLevelError**   | Events that prevent feature flags from functioning correctly (for example, invalid datafile in initialization and invalid feature keys) are logged. The user can take action to correct.                                          |
| **LogLevelWarning** | Events that don't prevent feature flags from functioning correctly, but can have unexpected outcomes (for example, future API deprecation, logger or error handler are not set properly, and nil values from getters) are logged. |
| **LogLevelInfo**    | Events of significance (for example, activate started, activate succeeded, tracking started, and tracking succeeded) are logged. This is helpful in showing the lifecycle of an API call.                                         |
| **LogLevelDebug**   | Any information related to errors that can help us debug the issue (for example, the feature flag is not running, user is not included in the rollout) are logged.                                                                |