/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

// Package decision //
package decision

import (
	"errors"
	"sync"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/optimizely/go-sdk/pkg/entities"
)

type forcedDecision struct {
	flagKey string
	ruleKey string
}

// ForcedDecisionService defines user contexts that the SDK will use to make decisions for.
type ForcedDecisionService struct {
	UserID          string
	forcedDecisions map[forcedDecision]string
	mutex           *sync.RWMutex
}

// NewForcedDecisionService returns an instance of the optimizely user context.
func NewForcedDecisionService(userID string) *ForcedDecisionService {
	return &ForcedDecisionService{
		UserID:          userID,
		forcedDecisions: map[forcedDecision]string{},
		mutex:           new(sync.RWMutex),
	}
}

// SetForcedDecision sets the forced decision (variation key) for a given flag and an optional rule.
// if rule key is empty, forced decision will be mapped against the flagKey.
// returns true if the forced decision has been set successfully.
func (f *ForcedDecisionService) SetForcedDecision(flagKey, ruleKey, variationKey string) bool {
	if flagKey == "" {
		return false
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.forcedDecisions[forcedDecision{flagKey: flagKey, ruleKey: ruleKey}] = variationKey
	return true
}

// GetForcedDecision returns the forced decision for a given flag and an optional rule
// if rule key is empty, forced decision will be returned for the flagKey.
func (f *ForcedDecisionService) GetForcedDecision(flagKey, ruleKey string) string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	if len(f.forcedDecisions) == 0 {
		return ""
	}
	if variationKey, ok := f.forcedDecisions[forcedDecision{flagKey: flagKey, ruleKey: ruleKey}]; ok {
		return variationKey
	}
	return ""
}

// RemoveForcedDecision removes the forced decision for a given flag and an optional rule.
// if rule key is empty, forced decision will be removed for the flagKey.
func (f *ForcedDecisionService) RemoveForcedDecision(flagKey, ruleKey string) bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	decision := forcedDecision{flagKey: flagKey, ruleKey: ruleKey}
	if f.forcedDecisions[decision] != "" {
		f.forcedDecisions[decision] = ""
		return true
	}
	return false
}

// RemoveAllForcedDecisions removes all forced decisions bound to this user context.
func (f *ForcedDecisionService) RemoveAllForcedDecisions() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.forcedDecisions = map[forcedDecision]string{}
	return true
}

// FindValidatedForcedDecision returns validated forced decision.
func (f *ForcedDecisionService) FindValidatedForcedDecision(projectConfig config.ProjectConfig, flagKey, ruleKey string, options *decide.Options) (variation *entities.Variation, reasons decide.DecisionReasons, err error) {
	decisionReasons := decide.NewDecisionReasons(options)
	variationKey := f.GetForcedDecision(flagKey, ruleKey)
	if variationKey == "" {
		return nil, decisionReasons, errors.New("decision not found")
	}

	_variation, err := f.getFlagVariationByKey(projectConfig, flagKey, variationKey)
	target := "flag (" + flagKey + ")"
	if ruleKey != "" {
		target += ", rule (" + ruleKey + ")"
	}

	if err != nil {
		decisionReasons.AddInfo("Invalid variation is mapped to %s and user (%s) in the forced decision map.", target, f.UserID)
		return nil, decisionReasons, err
	}
	decisionReasons.AddInfo("Variation (%s) is mapped to %s and user (%s) in the forced decision map.", variationKey, target, f.UserID)
	return _variation, decisionReasons, nil
}

func (f *ForcedDecisionService) getFlagVariationByKey(projectConfig config.ProjectConfig, flagKey, variationKey string) (*entities.Variation, error) {
	if variations, ok := projectConfig.GetFlagVariationsMap()[flagKey]; ok {
		for _, variation := range variations {
			if variation.Key == variationKey {
				return &variation, nil
			}
		}
	}
	return nil, errors.New("variation not found")
}

// CreateCopy creates and returns a copy of the forced decision service.
func (f *ForcedDecisionService) CreateCopy() *ForcedDecisionService {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	forceDecisions := map[forcedDecision]string{}
	for k, v := range f.forcedDecisions {
		forceDecisions[k] = v
	}
	return &ForcedDecisionService{
		UserID:          f.UserID,
		forcedDecisions: forceDecisions,
		mutex:           new(sync.RWMutex),
	}
}
