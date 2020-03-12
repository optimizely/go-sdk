/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package logging //
package logging

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// SdkKey key used to get and set the sdk key on the context passed in for creation.
const SdkKey string = "OptimizelySdkKey"

// GetSdkKey get the sdk key from the "activation context"
func GetSdkKey(ctx context.Context) string {
	if val := ctx.Value(SdkKey); val != nil {
		if v, ok := val.(string);ok {
			return v
		}
	}
	return ""
}

// LogLevel represents the level of the log (i.e. Debug, Info, Warning, Error)
type LogLevel int

var loggersCache = sync.Map{}

func setLogger(sdkKey string, logger OptimizelyLogConsumer) {
	loggersCache.Store(sdkKey, logger)
}

func getLogger(sdkKey string) OptimizelyLogConsumer {
	if val, ok := loggersCache.Load(sdkKey);ok && val != nil {
		if lp, ok := val.(OptimizelyLogConsumer); ok {
			return lp
		}
	}
	return nil
}

func (l LogLevel) String() string {
	return [...]string{"", "Debug", "Info", "Warning", "Error"}[l]
}

const (
	// LogLevelDebug log level
	LogLevelDebug LogLevel = iota + 1

	// LogLevelInfo log level
	LogLevelInfo

	// LogLevelWarning log level
	LogLevelWarning

	// LogLevelError log level
	LogLevelError
)

// SetLogger replaces the default logger with the given logger
func SetLogger(ctx context.Context, logger OptimizelyLogConsumer) {
		setLogger(GetSdkKey(ctx), logger)
}

// SetLogLevel sets the log level to the given level
func SetLogLevel(ctx context.Context, logLevel LogLevel) {
	if key := GetSdkKey(ctx); key != "" {
		if logger := getLogger(key); logger != nil {
			logger.SetLogLevel(logLevel)
		}
	}
}

// GetLogger returns a log producer with the given name
func GetLogger(ctx context.Context, name string) OptimizelyLogProducer {
	return NamedLogProducer{
		ctx: ctx,
		fields: map[string]interface{}{"name": name},
	}
}

// NamedLogProducer produces logs prefixed with its name
type NamedLogProducer struct {
	ctx context.Context
	fields map[string]interface{}
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
		message = fmt.Sprintf("%s: %v", message, err)
	}
	p.log(LogLevelError, message)
}

var defaultLogConsumer = NewFilteredLevelLogConsumer(LogLevelInfo, os.Stdout)
var loggerLock = &sync.Mutex{}

func (p NamedLogProducer) log(logLevel LogLevel, message string) {
	if logger := getLogger(GetSdkKey(p.ctx)); logger != nil {
		loggerLock.Lock()
		logger.Log(logLevel, message, p.fields)
		loggerLock.Unlock()
	} else {
		loggerLock.Lock()
		defaultLogConsumer.Log(logLevel, message, p.fields)
		loggerLock.Unlock()
	}
}
