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

package models

import "time"

// DispatcherType - represents event-dispatcher type
type DispatcherType string

const (
	// ProxyEventDispatcher - the event-dispatcher type is proxy
	ProxyEventDispatcher DispatcherType = "ProxyEventDispatcher"
	// NoOpEventDispatcher - the event-dispatcher type is no-op
	NoOpEventDispatcher DispatcherType = "NoopEventDispatcher"
)

// EventProcessorDefaultBatchSize - The default value for event processor batch size
const EventProcessorDefaultBatchSize = 1

// EventProcessorDefaultQueueSize - The default value for event processor queue size
const EventProcessorDefaultQueueSize = 1

// EventProcessorDefaultFlushInterval - The default value for event processor flush interval
const EventProcessorDefaultFlushInterval = 250 * time.Millisecond

// APIType - represents api type
type APIType string

const (
	// IsFeatureEnabled - the api type is IsFeatureEnabled
	IsFeatureEnabled APIType = "is_feature_enabled"
	// GetFeatureVariable - the api type is GetFeatureVariable
	GetFeatureVariable APIType = "get_feature_variable"
	// GetFeatureVariableInteger - the api type is GetFeatureVariableInteger
	GetFeatureVariableInteger APIType = "get_feature_variable_integer"
	// GetFeatureVariableDouble - the api type is GetFeatureVariableDouble
	GetFeatureVariableDouble APIType = "get_feature_variable_double"
	// GetFeatureVariableBoolean - the api type is GetFeatureVariableBoolean
	GetFeatureVariableBoolean APIType = "get_feature_variable_boolean"
	// GetFeatureVariableString - the api type is GetFeatureVariableString
	GetFeatureVariableString APIType = "get_feature_variable_string"
	// GetEnabledFeatures - the api type is GetEnabledFeatures
	GetEnabledFeatures APIType = "get_enabled_features"
)

// KeyListenerCalled - Key for listener called
const KeyListenerCalled = "listener_called"
