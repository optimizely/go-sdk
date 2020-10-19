/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                        *
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

// Package client //
package client

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptimizelyUserContext defines user contexts that the SDK will use to make decisions for.
type OptimizelyUserContext struct {
	optimizely  *OptimizelyClient
	userContext entities.UserContext
	mutex       sync.RWMutex
}

// NewOptimizelyUserContext returns an instance of the optimizely user context.
func NewOptimizelyUserContext(optimizely *OptimizelyClient, userContext entities.UserContext) *OptimizelyUserContext {
	// store a copy of the provided usercontext so it isn't affected by changes made afterwards.
	userContextCopy := copyUserContext(userContext)
	optimizelyUserContext := OptimizelyUserContext{
		optimizely:  optimizely,
		userContext: userContextCopy,
	}
	return &optimizelyUserContext
}

// GetOptimizely returns optimizely client instance for Optimizely user context
func (u *OptimizelyUserContext) GetOptimizely() *OptimizelyClient {
	return u.optimizely
}

// GetUserContext returns user settings for Optimizely user context
func (u *OptimizelyUserContext) GetUserContext() entities.UserContext {
	u.mutex.RLock()
	u.mutex.RUnlock()
	return copyUserContext(u.userContext)
}

// SetAttribute sets an attribute for a given key.
func (u *OptimizelyUserContext) SetAttribute(key string, value interface{}) {
	u.mutex.Lock()
	u.userContext.Attributes[key] = value
	u.mutex.Unlock()
}

// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (u *OptimizelyUserContext) Decide(key string) OptimizelyDecision {
	return CreateErrorDecision(key, nil, decide.GetDecideError(decide.SDKNotReady))
}

// DecideWithOptions returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (u *OptimizelyUserContext) DecideWithOptions(key string, options []decide.Options) OptimizelyDecision {
	return CreateErrorDecision(key, nil, decide.GetDecideError(decide.SDKNotReady))
}

// DecideAll returns a key-map of decision results for all active flag keys.
func (u *OptimizelyUserContext) DecideAll() map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// DecideAllWithOptions returns a key-map of decision results for all active flag keys with options.
func (u *OptimizelyUserContext) DecideAllWithOptions(options []decide.Options) map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// DecideForKeys returns a key-map of decision results for multiple flag keys.
func (u *OptimizelyUserContext) DecideForKeys(keys []string) map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// DecideForKeysWithOptions returns a key-map of decision results for multiple flag keys and options.
func (u *OptimizelyUserContext) DecideForKeysWithOptions(keys []string, options []decide.Options) map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// TrackEvent generates a conversion event with the given event key if it exists and queues it up to be sent to the Optimizely
// log endpoint for results processing.
func (u *OptimizelyUserContext) TrackEvent(eventKey string, eventTags map[string]interface{}) (err error) {
	return u.optimizely.Track(eventKey, u.userContext, eventTags)
}

func copyUserContext(userContext entities.UserContext) entities.UserContext {
	userContextCopy := userContext
	userContextCopy.Attributes = make(map[string]interface{})
	if userContext.Attributes != nil {
		for k, v := range userContext.Attributes {
			userContextCopy.Attributes[k] = v
		}
	}
	return userContextCopy
}
