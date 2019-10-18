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

// OptimizelyLogConsumer consumes log messages produced by the log producers
type OptimizelyLogConsumer interface {
	Log(level LogLevel, message string, fields map[string]interface{})
	SetLogLevel(logLevel LogLevel)
}

// OptimizelyLogProducer produces log messages to be consumed by the log consumer
type OptimizelyLogProducer interface {
	Debug(message string)
	Info(message string)
	Warning(message string)
	Error(message string, err interface{})
}
