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

package logging

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockOptimizelyLogger struct {
	mock.Mock
	loggedMessages []string
}

func (m *MockOptimizelyLogger) Log(level LogLevel, message string, fields map[string]interface{}) {
	m.Called(level, message, fields["name"])
	m.loggedMessages = append(m.loggedMessages, message)
}

func (m *MockOptimizelyLogger) SetLogLevel(level LogLevel) {
	m.Called(level)
}

func TestNamedLoggerDebug(t *testing.T) {
	testLogMessage := "Test debug message"
	testLogName := "test-debug"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelDebug, testLogMessage, testLogName)

	SetLogger(testLogger)

	logProducer := GetLogger("", testLogName)
	logProducer.Debug(testLogMessage)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{testLogMessage}, testLogger.loggedMessages)
}

func TestNamedLoggerInfo(t *testing.T) {
	testLogMessage := "Test info message"
	testLogName := "test-info"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelInfo, testLogMessage, testLogName)

	SetLogger(testLogger)

	logProducer := GetLogger("testSdkKey", testLogName)
	logProducer.Info(testLogMessage)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{testLogMessage}, testLogger.loggedMessages)
}

func TestNamedLoggerWarning(t *testing.T) {
	testLogMessage := "Test warn message"
	testLogName := "test-warn"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelWarning, testLogMessage, testLogName)

	SetLogger(testLogger)

	logProducer := GetLogger("testSdkKey", testLogName)
	logProducer.Warning(testLogMessage)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{testLogMessage}, testLogger.loggedMessages)
}

func TestNamedLoggerError(t *testing.T) {
	testLogMessage := "Test error message"
	testLogName := "test-error"
	expectedLogMessage := "Test error message: I am an error object"
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("Log", LogLevelError, expectedLogMessage, testLogName)
	SetLogger(testLogger)

	err := errors.New("I am an error object")
	logProducer := GetLogger("", testLogName)
	logProducer.Error(testLogMessage, err)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{expectedLogMessage}, testLogger.loggedMessages)
}

func TestNamedLoggerFields(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelDebug, out)

	SetLogger(newLogger)

	logger := GetLogger("TestNamedLoggerFields-sdkKey", "TestNamedLoggerFields")
	logger.Debug("test message")

	key := GetSdkKeyLogMapping("TestNamedLoggerFields-sdkKey")
	assert.Contains(t, out.String(), "test message")
	assert.Contains(t, out.String(), "[Debug]")
	assert.Contains(t, out.String(), "[TestNamedLoggerFields]")
	assert.Contains(t, out.String(), key)
	assert.NotContains(t, out.String(), "TestNamedLoggerFields-sdkKey")
	assert.Contains(t, out.String(), "[Optimizely]")

}

func TestLogSdkKeyOverride(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelDebug, out)

	SetLogger(newLogger)

	key := "test-test-test"
	SetSdkKeyLogMapping("TestLogOverride-sdkKey", key)

	logger := GetLogger("TestLogOverride-sdkKey", "TestLogSdkKeyOverride")
	logger.Debug("test message")

	assert.Contains(t, out.String(), "test message")
	assert.Contains(t, out.String(), "[Debug]")
	assert.Contains(t, out.String(), "[TestLogSdkKeyOverride]")
	assert.Contains(t, out.String(), key)
	assert.NotContains(t, out.String(), "TestNamedLoggerFields-sdkKey")
	assert.Contains(t, out.String(), "[Optimizely]")
}

func TestLogSdkKey(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelDebug, out)

	SetLogger(newLogger)

	key := "TestLogSdkKey-sdkKey"

	UseSdkKeyForLogging(key)

	logger := GetLogger(key, "TestLogSdkKeyOverride")
	logger.Debug("test message")

	assert.Contains(t, out.String(), "test message")
	assert.Contains(t, out.String(), "[Debug]")
	assert.Contains(t, out.String(), "[TestLogSdkKeyOverride]")
	assert.Contains(t, out.String(), key)
	assert.Contains(t, out.String(), "[Optimizely]")
}

func TestLoggingOrder(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelDebug, out)

	SetLogger(newLogger)

	key := "TestLoggingOrder-sdkKey"

	logger := GetLogger(key, "TestLoggingOrder")
	logger.Debug("test message")

	key = GetSdkKeyLogMapping(key)

	response := out.String()

	assert.Contains(t, response, "test message")
	assert.Contains(t, response, "[Debug][" + key +"][TestLoggingOrder] test message" )
	assert.True(t, strings.HasPrefix(response, "[Optimizely]"))

}

func TestLoggingOrderEmpty(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelDebug, out)

	SetLogger(newLogger)

	key := ""

	logger := GetLogger(key, "TestLoggingOrder")
	logger.Debug("test message")

	key = GetSdkKeyLogMapping(key)

	response := out.String()

	assert.Contains(t, response, "test message")
	assert.Contains(t, response, "[Debug][TestLoggingOrder] test message" )
	assert.True(t, strings.HasPrefix(response, "[Optimizely]"))

}

func TestSetLogLevel(t *testing.T) {
	testLogger := new(MockOptimizelyLogger)
	testLogger.On("SetLogLevel", LogLevelError)

	SetLogger(testLogger)
	SetLogLevel(LogLevelError)

	testLogger.AssertExpectations(t)
}
