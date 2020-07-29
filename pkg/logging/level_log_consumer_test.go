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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilteredLogging(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelInfo, out)

	assert.Equal(t, newLogger.level, LogLevel(2))
	assert.NotNil(t, newLogger.logger)

	newLogger.SetLogLevel(3)
	assert.Equal(t, newLogger.level, LogLevel(3))

	newLogger.Log(1, "this is hidden", map[string]interface{}{})
	assert.Equal(t, "", out.String())
	out.Reset()

	newLogger.Log(4, "this is visible", map[string]interface{}{})
	assert.Contains(t, out.String(), "this is visible")
	out.Reset()
}

func TestLogFormatting(t *testing.T) {
	out := &bytes.Buffer{}
	newLogger := NewFilteredLevelLogConsumer(LogLevelInfo, out)

	newLogger.Log(LogLevelInfo, "test message", map[string]interface{}{"name": "test-name", "sdkKey": "testLogFormatting-sdkKey"})
	assert.Contains(t, out.String(), "test message")
	assert.Contains(t, out.String(), "[Info]")
	assert.Contains(t, out.String(), "[test-name]")
	assert.Contains(t, out.String(), "[testLogFormatting-sdkKey]")
	assert.Contains(t, out.String(), "[Optimizely]")
}

func BenchmarkLogger(b *testing.B) {
	b.Run("FilteredLevelLogConsumer-Stdout", func(b *testing.B) {
		logger := NewFilteredLevelLogConsumer(LogLevelInfo, os.Stdout)
		benchmarkLogger(b, logger)
	})
	b.Run("FilteredLevelLogConsumer-ByteBuffer", func(b *testing.B) {
		out := &bytes.Buffer{}
		logger := NewFilteredLevelLogConsumer(LogLevelInfo, out)
		benchmarkLogger(b, logger)
	})
	b.Run("FilteredLevelLogConsumer-Discard", func(b *testing.B) {
		logger := NewFilteredLevelLogConsumer(LogLevelInfo, ioutil.Discard)
		benchmarkLogger(b, logger)
	})
	b.Run("ZeroLogConsumer", func(b *testing.B) {
		logger := NewZeroLogConsumer(LogLevelInfo)
		benchmarkLogger(b, logger)
	})
}

func benchmarkLogger(b *testing.B, logger OptimizelyLogConsumer) {
	fields := map[string]interface{}{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}

	for i := 0; i < b.N; i++ {
		logger.Log(LogLevelInfo, "test message", fields)
	}
}
