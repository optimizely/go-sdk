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
	mutex       *sync.RWMutex
}

// returns an instance of the optimizely user context.
func newOptimizelyUserContext(optimizely *OptimizelyClient, userContext entities.UserContext) OptimizelyUserContext {
	// store a copy of the provided usercontext so it isn't affected by changes made afterwards.
	userContextCopy := copyUserContext(userContext)
	return OptimizelyUserContext{
		optimizely:  optimizely,
		userContext: userContextCopy,
		mutex:       new(sync.RWMutex),
	}
}

// GetOptimizely returns optimizely client instance for Optimizely user context
func (o *OptimizelyUserContext) GetOptimizely() *OptimizelyClient {
	return o.optimizely
}

// GetUserContext returns user settings for Optimizely user context
func (o *OptimizelyUserContext) GetUserContext() entities.UserContext {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return copyUserContext(o.userContext)
}

// SetAttribute sets an attribute for a given key.
func (o *OptimizelyUserContext) SetAttribute(key string, value interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.userContext.Attributes[key] = value
}

// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (o *OptimizelyUserContext) Decide(key string, options []decide.Options) OptimizelyDecision {
	if options == nil {
		return o.optimizely.decide(*o, key, []decide.Options{})
	}
	return o.optimizely.decide(*o, key, options)
}

// DecideAll returns a key-map of decision results for all active flag keys with options.
func (o *OptimizelyUserContext) DecideAll(options []decide.Options) map[string]OptimizelyDecision {
	if options == nil {
		return o.optimizely.decideAll(*o, []decide.Options{})
	}
	return o.optimizely.decideAll(*o, options)
}

// DecideForKeys returns a key-map of decision results for multiple flag keys and options.
func (o *OptimizelyUserContext) DecideForKeys(keys []string, options []decide.Options) map[string]OptimizelyDecision {
	if options == nil {
		return o.optimizely.decideForKeys(*o, keys, []decide.Options{})
	}
	return o.optimizely.decideForKeys(*o, keys, options)
}

// TrackEvent generates a conversion event with the given event key if it exists and queues it up to be sent to the Optimizely
// log endpoint for results processing.
func (o *OptimizelyUserContext) TrackEvent(eventKey string, eventTags map[string]interface{}) (err error) {
	return o.optimizely.Track(eventKey, o.GetUserContext(), eventTags)
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
