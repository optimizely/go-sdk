/****************************************************************************
 * Copyright 2020-2021, Optimizely, Inc. and contributors                   *
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
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
)

// OptimizelyUserContext defines user contexts that the SDK will use to make decisions for.
type OptimizelyUserContext struct {
	UserID     string                 `json:"userId"`
	Attributes map[string]interface{} `json:"attributes"`

	optimizely            *OptimizelyClient
	forcedDecisionService *decision.ForcedDecisionService
	mutex                 *sync.RWMutex
}

// returns an instance of the optimizely user context.
func newOptimizelyUserContext(optimizely *OptimizelyClient, userID string, attributes map[string]interface{}, forcedDecisionService *decision.ForcedDecisionService) OptimizelyUserContext {
	// store a copy of the provided attributes so it isn't affected by changes made afterwards.
	if attributes == nil {
		attributes = map[string]interface{}{}
	}
	attributesCopy := copyUserAttributes(attributes)
	return OptimizelyUserContext{
		UserID:                userID,
		Attributes:            attributesCopy,
		optimizely:            optimizely,
		forcedDecisionService: forcedDecisionService,
		mutex:                 new(sync.RWMutex),
	}
}

// GetOptimizely returns optimizely client instance for Optimizely user context
func (o OptimizelyUserContext) GetOptimizely() *OptimizelyClient {
	return o.optimizely
}

// GetUserID returns userID for Optimizely user context
func (o OptimizelyUserContext) GetUserID() string {
	return o.UserID
}

// GetUserAttributes returns user attributes for Optimizely user context
func (o OptimizelyUserContext) GetUserAttributes() map[string]interface{} {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return copyUserAttributes(o.Attributes)
}

func (o OptimizelyUserContext) getForcedDecisionService() *decision.ForcedDecisionService {
	if o.forcedDecisionService != nil {
		return o.forcedDecisionService.CreateCopy()
	}
	return nil
}

// SetAttribute sets an attribute for a given key.
func (o *OptimizelyUserContext) SetAttribute(key string, value interface{}) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.Attributes == nil {
		o.Attributes = make(map[string]interface{})
	}
	o.Attributes[key] = value
}

// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (o *OptimizelyUserContext) Decide(key string, options []decide.OptimizelyDecideOptions) OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService())
	return o.optimizely.decide(userContextCopy, key, convertDecideOptions(options))
}

// DecideAll returns a key-map of decision results for all active flag keys with options.
func (o *OptimizelyUserContext) DecideAll(options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService())
	return o.optimizely.decideAll(userContextCopy, convertDecideOptions(options))
}

// DecideForKeys returns a key-map of decision results for multiple flag keys and options.
func (o *OptimizelyUserContext) DecideForKeys(keys []string, options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService())
	return o.optimizely.decideForKeys(userContextCopy, keys, convertDecideOptions(options))
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

// SetForcedDecision sets the forced decision (variation key) for a given flag and an optional rule.
// returns true if the forced decision has been set successfully.
func (o *OptimizelyUserContext) SetForcedDecision(flagKey, ruleKey, variationKey string) bool {
	if _, err := o.optimizely.getProjectConfig(); err != nil {
		o.optimizely.logger.Error("Optimizely instance is not valid, failing setForcedDecision call.", err)
		return false
	}
	if o.forcedDecisionService == nil {
		o.forcedDecisionService = decision.NewForcedDecisionService(o.GetUserID())
	}
	return o.forcedDecisionService.SetForcedDecision(flagKey, ruleKey, variationKey)
}

// GetForcedDecision returns the forced decision for a given flag and an optional rule
func (o *OptimizelyUserContext) GetForcedDecision(flagKey, ruleKey string) string {
	if _, err := o.optimizely.getProjectConfig(); err != nil {
		o.optimizely.logger.Error("Optimizely instance is not valid, failing getForcedDecision call.", err)
		return ""
	}
	if o.forcedDecisionService == nil {
		return ""
	}
	return o.forcedDecisionService.GetForcedDecision(flagKey, ruleKey)
}

// RemoveForcedDecision removes the forced decision for a given flag and an optional rule.
func (o *OptimizelyUserContext) RemoveForcedDecision(flagKey, ruleKey string) bool {
	if _, err := o.optimizely.getProjectConfig(); err != nil {
		o.optimizely.logger.Error("Optimizely instance is not valid, failing removeForcedDecision call.", err)
		return false
	}
	if o.forcedDecisionService == nil {
		return false
	}
	return o.forcedDecisionService.RemoveForcedDecision(flagKey, ruleKey)
}

// RemoveAllForcedDecisions removes all forced decisions bound to this user context.
func (o *OptimizelyUserContext) RemoveAllForcedDecisions() bool {
	if _, err := o.optimizely.getProjectConfig(); err != nil {
		o.optimizely.logger.Error("Optimizely instance is not valid, failing removeForcedDecision call.", err)
		return false
	}
	if o.forcedDecisionService == nil {
		return true
	}
	return o.forcedDecisionService.RemoveAllForcedDecisions()
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
