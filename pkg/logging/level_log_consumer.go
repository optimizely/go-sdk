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
	"io"
	"log"
	"sort"
	"strings"
)

// FilteredLevelLogConsumer is an implementation of the OptimizelyLogConsumer that filters by log level
type FilteredLevelLogConsumer struct {
	level  LogLevel
	logger *log.Logger
}

// Log logs the message if it's log level is higher than or equal to the logger's set level
func (l *FilteredLevelLogConsumer) Log(level LogLevel, message string, fields map[string]interface{}) {
	if l.level <= level {
		// prepends the name and log level to the message
		messBuilder := strings.Builder{}
		_, _ = messBuilder.WriteString("[")

		_, _ = messBuilder.WriteString(level.String())
		_, _ = messBuilder.WriteString("]")
		keys := make([]string, len(fields))
		for k,_ := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s, ok := fields[k].(string);ok {
				if s == "" {
					continue
				}
				_, _ = messBuilder.WriteString("[")
				_, _ = messBuilder.WriteString(s)
				_, _ = messBuilder.WriteString("]")
			}
		}
		_, _ = messBuilder.WriteString(" ")
		_, _ = messBuilder.WriteString(message)
		l.logger.Println(messBuilder.String())
	}
}

// SetLogLevel changes the log level to the given level
func (l *FilteredLevelLogConsumer) SetLogLevel(level LogLevel) {
	l.level = level
}

// NewFilteredLevelLogConsumer returns a new logger that logs to stdout
func NewFilteredLevelLogConsumer(level LogLevel, out io.Writer) *FilteredLevelLogConsumer {
	return &FilteredLevelLogConsumer{
		level:  level,
		logger: log.New(out, "[Optimizely]", log.LstdFlags),
	}
}
