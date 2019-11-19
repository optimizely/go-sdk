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

// Package decision //
package decision

import "sync"

// DefaultUserProfileService represents the default implementation of UserProfileService interface
type DefaultUserProfileService struct {
	sync.RWMutex
	profiles map[string]UserProfile
}

// Lookup is used to retrieve past bucketing decisions for users
func (s *DefaultUserProfileService) Lookup(userID string) UserProfile {
	s.RLock()
	profile := s.profiles[userID]
	s.RUnlock()
	return profile
}

// Save is used to save bucketing decisions for users
func (s *DefaultUserProfileService) Save(userProfile UserProfile) {
	if userProfile.ID == "" {
		return
	}
	s.Lock()
	if s.profiles == nil {
		s.profiles = make(map[string]UserProfile)
	}
	s.profiles[userProfile.ID] = userProfile
	s.Unlock()
}
