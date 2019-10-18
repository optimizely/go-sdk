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
	"errors"
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

	logProducer := GetLogger(testLogName)
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

	logProducer := GetLogger(testLogName)
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

	logProducer := GetLogger(testLogName)
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
	logProducer := GetLogger(testLogName)
	logProducer.Error(testLogMessage, err)
	testLogger.AssertExpectations(t)
	assert.Equal(t, []string{expectedLogMessage}, testLogger.loggedMessages)
}
