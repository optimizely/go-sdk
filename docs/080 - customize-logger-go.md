---
title: "Customize logger"
slug: "customize-logger-go"
hidden: false
createdAt: "2019-08-21T21:14:40.407Z"
updatedAt: "2019-08-21T21:15:15.382Z"
---
The **logger** logs information about your experiments to help you with debugging. You can customize where log information is sent and what kind of information is tracked.

The default logger in the Go SDK logs to STDOUT. If you want to disable or customize the logger, you can provide an implementation of the **OptimizelyLogConsumer** interface.
[block:api-header]
{
  "title": "Custom logger implementation in the SDK"
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "import \"github.com/optimizely/go-sdk/optimizely/logging\"\n\n\ntype CustomLogger struct {\n}\n\n\nfunc (l *CustomLogger) Log(level LogLevel, message string) {\n\n}\n\n\nfunc (l *CustomLogger) SetLogLevel(level LogLevel) {\n\n}\n\ncustomLogger := New(CustomLogger)\n\nlogging.setLogger(customLogger)\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Setting the log level"
}
[/block]
You can also change the default log level from INFO to any of the other log levels.
[block:code]
{
  "codes": [
    {
      "code": "import \"github.com/optimizely/go-sdk/optimizely/logging\"\n\n// Set log level to Debug\nlogging.SetLogLevel(logging.LogLevelDebug)\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Log levels"
}
[/block]
The table below lists the log levels for the Go SDK.
[block:parameters]
{
  "data": {
    "h-0": "Log Level",
    "h-1": "Explanation",
    "0-0": "**LogLevelError**",
    "0-1": "Events that prevent feature flags from functioning correctly (for example, invalid datafile in initialization and invalid feature keys) are logged. The user can take action to correct.",
    "1-0": "**LogLevelWarning**",
    "1-1": "Events that don't prevent feature flags from functioning correctly, but can have unexpected outcomes (for example, future API deprecation, logger or error handler are not set properly, and nil values from getters) are logged.",
    "2-0": "**LogLevelInfo**",
    "2-1": "Events of significance (for example, activate started, activate succeeded, tracking started, and tracking succeeded) are logged. This is helpful in showing the lifecycle of an API call.",
    "3-0": "**LogLevelDebug**",
    "3-1": "Any information related to errors that can help us debug the issue (for example, the feature flag is not running, user is not included in the rollout) are logged."
  },
  "cols": 2,
  "rows": 4
}
[/block]