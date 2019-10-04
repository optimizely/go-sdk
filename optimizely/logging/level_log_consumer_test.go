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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStdoutFilteredLevelLogConsumer(t *testing.T) {
	newLogger := NewStdoutFilteredLevelLogConsumer(LogLevelInfo)

	assert.Equal(t, newLogger.level, LogLevel(2))
	assert.NotNil(t, newLogger.logger)

	newLogger.SetLogLevel(3)
	assert.Equal(t, newLogger.level, LogLevel(3))

	newLogger.Log(1, "this is hidden")
	newLogger.Log(4, "this is visible")
}
