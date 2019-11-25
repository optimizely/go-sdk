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

import "github.com/optimizely/go-sdk/pkg/decision"

// NoOpUserProfileService represents a user profile service with save and lookup error
type NoOpUserProfileService struct {
	NormalUserProfileService
}

// Lookup is used to retrieve past bucketing decisions for users
func (s *NoOpUserProfileService) Lookup(userID string) decision.UserProfile {
	return decision.UserProfile{}
}

// Save is used to save bucketing decisions for users
func (s *NoOpUserProfileService) Save(userProfile decision.UserProfile) {
}
