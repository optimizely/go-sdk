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

package userprofileservice

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/decision"
)

// NormalUserProfileService represents the default implementation of UserProfileService interface
type NormalUserProfileService struct {
	sync.RWMutex
	profiles map[string]decision.UserProfile
}

// Lookup is used to retrieve past bucketing decisions for users
func (s *NormalUserProfileService) Lookup(userID string) decision.UserProfile {
	s.RLock()
	profile := s.profiles[userID]
	s.RUnlock()
	return profile
}

// Save is used to save bucketing decisions for users
func (s *NormalUserProfileService) Save(userProfile decision.UserProfile) {
	if userProfile.ID == "" {
		return
	}
	s.Lock()
	if s.profiles == nil {
		s.profiles = make(map[string]decision.UserProfile)
	}
	if savedProfile, ok := s.profiles[userProfile.ID]; ok {
		for k, v := range userProfile.ExperimentBucketMap {
			savedProfile.ExperimentBucketMap[k] = v
		}
		s.profiles[userProfile.ID] = savedProfile
	} else {
		s.profiles[userProfile.ID] = userProfile
	}
	s.Unlock()
}

// SaveUserProfiles saves multiple user profiles
func (s *NormalUserProfileService) SaveUserProfiles(userProfiles []decision.UserProfile) {
	for _, profile := range userProfiles {
		s.Save(profile)
	}
}

// GetUserProfiles returns currently saved user profiles
func (s *NormalUserProfileService) GetUserProfiles() (savedProfiles []decision.UserProfile) {
	for _, v := range s.profiles {
		savedProfiles = append(savedProfiles, v)
	}
	return savedProfiles
}
