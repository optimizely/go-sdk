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
	"fmt"
	"os"
	"sync"
)

// LogLevel represents the level of the log (i.e. Debug, Info, Warning, Error)
type LogLevel int

func (l LogLevel) String() string {
	return [...]string{"", "Debug", "Info", "Warning", "Error"}[l]
}

var defaultLogConsumer OptimizelyLogConsumer
var mutex = &sync.Mutex{}

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

func init() {
	mutex.Lock()
	defaultLogConsumer = NewFilteredLevelLogConsumer(LogLevelInfo, os.Stdout)
	mutex.Unlock()
}

// SetLogger replaces the default logger with the given logger
func SetLogger(logger OptimizelyLogConsumer) {
	mutex.Lock()
	defaultLogConsumer = logger
	mutex.Unlock()
}

// SetLogLevel sets the log level to the given level
func SetLogLevel(logLevel LogLevel) {
	mutex.Lock()
	defaultLogConsumer.SetLogLevel(logLevel)
	mutex.Unlock()
}

// GetLogger returns a log producer with the given name
func GetLogger(sdkKey string, name string) OptimizelyLogProducer {
	return NamedLogProducer{
		fields: map[string]interface{}{"name": name, "sdkKey":sdkKey},
	}
}

// NamedLogProducer produces logs prefixed with its name
type NamedLogProducer struct {
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

func (p NamedLogProducer) log(logLevel LogLevel, message string) {
	mutex.Lock()
	defaultLogConsumer.Log(logLevel, message, p.fields)
	mutex.Unlock()
}
