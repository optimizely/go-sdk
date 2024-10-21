/****************************************************************************
 * Copyright 2020-2022, 2024 Optimizely, Inc. and contributors              *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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
	"errors"
	"sync"

	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/decision"
	pkgDecision "github.com/optimizely/go-sdk/v2/pkg/decision"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	pkgOdpSegment "github.com/optimizely/go-sdk/v2/pkg/odp/segment"
)

// OptimizelyUserContext defines user contexts that the SDK will use to make decisions for.
type OptimizelyUserContext struct {
	UserID     string                 `json:"userId"`
	Attributes map[string]interface{} `json:"attributes"`

	qualifiedSegments     []string
	optimizely            *OptimizelyClient
	forcedDecisionService *pkgDecision.ForcedDecisionService
	mutex                 *sync.RWMutex
	userProfile           *pkgDecision.UserProfile
}

// returns an instance of the optimizely user context.
func newOptimizelyUserContext(optimizely *OptimizelyClient, userID string, attributes map[string]interface{}, forcedDecisionService *pkgDecision.ForcedDecisionService, qualifiedSegments []string) OptimizelyUserContext {
	// store a copy of the provided attributes so it isn't affected by changes made afterwards.
	if attributes == nil {
		attributes = map[string]interface{}{}
	}
	attributesCopy := copyUserAttributes(attributes)
	qualifiedSegmentsCopy := copyQualifiedSegments(qualifiedSegments)
	return OptimizelyUserContext{
		UserID:                userID,
		Attributes:            attributesCopy,
		qualifiedSegments:     qualifiedSegmentsCopy,
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

// GetQualifiedSegments returns qualified segments for Optimizely user context
func (o *OptimizelyUserContext) GetQualifiedSegments() []string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return copyQualifiedSegments(o.qualifiedSegments)
}

func (o OptimizelyUserContext) getForcedDecisionService() *pkgDecision.ForcedDecisionService {
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

// FetchQualifiedSegments fetches all qualified segments for the user context.
func (o *OptimizelyUserContext) FetchQualifiedSegments(options []pkgOdpSegment.OptimizelySegmentOption) (success bool) {
	o.optimizely.fetchQualifiedSegments(o, options, func(result bool) {
		success = result
	})
	return
}

// FetchQualifiedSegmentsAsync fetches all qualified segments aysnchronously for the user context.
func (o *OptimizelyUserContext) FetchQualifiedSegmentsAsync(options []pkgOdpSegment.OptimizelySegmentOption, callback func(success bool)) {
	go o.optimizely.fetchQualifiedSegments(o, options, callback)
}

// SetQualifiedSegments clears and adds qualified segments for Optimizely user context
func (o *OptimizelyUserContext) SetQualifiedSegments(qualifiedSegments []string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.qualifiedSegments = copyQualifiedSegments(qualifiedSegments)
}

// IsQualifiedFor returns true if the user is qualified for the given segment name
func (o *OptimizelyUserContext) IsQualifiedFor(segment string) bool {
	userContext := entities.UserContext{
		QualifiedSegments: o.GetQualifiedSegments(),
	}
	return userContext.IsQualifiedFor(segment)
}

// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (o *OptimizelyUserContext) Decide(key string, options []decide.OptimizelyDecideOptions) OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService(), o.GetQualifiedSegments())
	decideOptions := convertDecideOptions(options)

	if !decideOptions.IgnoreUserProfileService && o.optimizely.UserProfileService != nil {
		userProfile := decision.UserProfile{
			ID:                  userContextCopy.GetUserID(),
			ExperimentBucketMap: make(map[decision.UserDecisionKey]string),
		}
		userContextCopy.SetUserProfile(&userProfile)
	}

	decision := o.optimizely.decide(userContextCopy, key, decideOptions)
	if userContextCopy.userProfile != nil && len(userContextCopy.userProfile.ExperimentBucketMap) > 0 {
		o.optimizely.UserProfileService.Save(*userContextCopy.userProfile)
	}
	return decision
}

// DecideAll returns a key-map of decision results for all active flag keys with options.
func (o *OptimizelyUserContext) DecideAll(options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService(), o.GetQualifiedSegments())
	decideOptions := convertDecideOptions(options)

	if !decideOptions.IgnoreUserProfileService && o.optimizely.UserProfileService != nil {
		userProfile := decision.UserProfile{
			ID:                  userContextCopy.GetUserID(),
			ExperimentBucketMap: make(map[decision.UserDecisionKey]string),
		}
		userContextCopy.SetUserProfile(&userProfile)
	}

	decision := o.optimizely.decideAll(userContextCopy, decideOptions)
	if userContextCopy.userProfile != nil && len(userContextCopy.userProfile.ExperimentBucketMap) > 0 {
		o.optimizely.UserProfileService.Save(*userContextCopy.userProfile)
	}
	return decision
}

// DecideForKeys returns a key-map of decision results for multiple flag keys and options.
func (o *OptimizelyUserContext) DecideForKeys(keys []string, options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision {
	// use a copy of the user context so that any changes to the original context are not reflected inside the decision
	userContextCopy := newOptimizelyUserContext(o.GetOptimizely(), o.GetUserID(), o.GetUserAttributes(), o.getForcedDecisionService(), o.GetQualifiedSegments())
	decideOptions := convertDecideOptions(options)

	if !decideOptions.IgnoreUserProfileService && o.optimizely.UserProfileService != nil {
		userProfile := decision.UserProfile{
			ID:                  userContextCopy.GetUserID(),
			ExperimentBucketMap: make(map[decision.UserDecisionKey]string),
		}
		userContextCopy.SetUserProfile(&userProfile)
	}

	decision := o.optimizely.decideForKeys(userContextCopy, keys, convertDecideOptions(options))
	if userContextCopy.userProfile != nil && len(userContextCopy.userProfile.ExperimentBucketMap) > 0 {
		o.optimizely.UserProfileService.Save(*userContextCopy.userProfile)
	}
	return decision
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

// SetForcedDecision sets the forced decision (variation key) for a given decision context (flag key and optional rule key).
// returns true if the forced decision has been set successfully.
func (o *OptimizelyUserContext) SetForcedDecision(ctx pkgDecision.OptimizelyDecisionContext, decision pkgDecision.OptimizelyForcedDecision) bool {
	if o.forcedDecisionService == nil {
		o.forcedDecisionService = pkgDecision.NewForcedDecisionService(o.GetUserID())
	}
	return o.forcedDecisionService.SetForcedDecision(ctx, decision)
}

// GetForcedDecision returns the forced decision for a given flag and an optional rule
func (o *OptimizelyUserContext) GetForcedDecision(ctx pkgDecision.OptimizelyDecisionContext) (pkgDecision.OptimizelyForcedDecision, error) {
	if o.forcedDecisionService == nil {
		return pkgDecision.OptimizelyForcedDecision{}, errors.New("decision not found")
	}
	return o.forcedDecisionService.GetForcedDecision(ctx)
}

// RemoveForcedDecision removes the forced decision for a given flag and an optional rule.
func (o *OptimizelyUserContext) RemoveForcedDecision(ctx pkgDecision.OptimizelyDecisionContext) bool {
	if o.forcedDecisionService == nil {
		return false
	}
	return o.forcedDecisionService.RemoveForcedDecision(ctx)
}

// RemoveAllForcedDecisions removes all forced decisions bound to this user context.
func (o *OptimizelyUserContext) RemoveAllForcedDecisions() bool {
	if o.forcedDecisionService == nil {
		return true
	}
	return o.forcedDecisionService.RemoveAllForcedDecisions()
}

// SetUserProfile set the user profile for the user context
func (o *OptimizelyUserContext) SetUserProfile(userProfile *pkgDecision.UserProfile) {
	o.userProfile = userProfile
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

func copyQualifiedSegments(qualifiedSegments []string) (qualifiedSegmentsCopy []string) {
	if qualifiedSegments == nil {
		return nil
	}
	qualifiedSegmentsCopy = make([]string, len(qualifiedSegments))
	copy(qualifiedSegmentsCopy, qualifiedSegments)
	return
}
