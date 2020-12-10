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
	userID     string
	attributes map[string]interface{}
	optimizely *OptimizelyClient
	mutex      *sync.RWMutex
}

// returns an instance of the optimizely user context.
func newOptimizelyUserContext(optimizely *OptimizelyClient, userID string, attributes map[string]interface{}) OptimizelyUserContext {
	// store a copy of the provided attributes so it isn't affected by changes made afterwards.
	attributesCopy := copyUserAttributes(attributes)

	return OptimizelyUserContext{
		userID:     userID,
		attributes: attributesCopy,
		optimizely: optimizely,
		mutex:      new(sync.RWMutex),
	}
}

// GetOptimizely returns optimizely client instance for Optimizely user context
func (o *OptimizelyUserContext) GetOptimizely() *OptimizelyClient {
	return o.optimizely
}

// GetUserID returns userID for Optimizely user context
func (o *OptimizelyUserContext) GetUserID() string {
	return o.userID
}

// GetUserAttributes returns user attributes for Optimizely user context
func (o *OptimizelyUserContext) GetUserAttributes() map[string]interface{} {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return copyUserAttributes(o.attributes)
}

// SetAttribute sets an attribute for a given key.
func (o *OptimizelyUserContext) SetAttribute(key string, value interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.attributes == nil {
		o.attributes = make(map[string]interface{})
	}
	o.attributes[key] = value
}

// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (o *OptimizelyUserContext) Decide(key string, options decide.OptimizelyDecideOptions) OptimizelyDecision {
	return NewErrorDecision(key, OptimizelyUserContext{}, decide.GetError(decide.SDKNotReady))
}

// DecideAll returns a key-map of decision results for all active flag keys with options.
func (o *OptimizelyUserContext) DecideAll(options decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// DecideForKeys returns a key-map of decision results for multiple flag keys and options.
func (o *OptimizelyUserContext) DecideForKeys(keys []string, options decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	return map[string]OptimizelyDecision{}
}

// TrackEvent generates a conversion event with the given event key if it exists and queues it up to be sent to the Optimizely
// log endpoint for results processing.
func (o *OptimizelyUserContext) TrackEvent(eventKey string, eventTags map[string]interface{}) (err error) {
	userContext := entities.UserContext{
		ID:         o.GetUserID(),
		Attributes: o.GetUserAttributes(),
	}
	return o.optimizely.Track(eventKey, userContext, eventTags)
}

func copyUserAttributes(attributes map[string]interface{}) (attributesCopy map[string]interface{}) {
	if attributes != nil {
		attributesCopy = make(map[string]interface{})
		for k, v := range attributes {
			attributesCopy[k] = v
		}
	}
	return attributesCopy
}
