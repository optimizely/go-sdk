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

package decision

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupWithEmptyUserID(t *testing.T) {
	defaultUserProfileService := DefaultUserProfileService{}
	profile := defaultUserProfileService.Lookup("")
	assert.Equal(t, UserProfile{}, profile)
}

func TestLookupNotFound(t *testing.T) {
	defaultUserProfileService := DefaultUserProfileService{}
	profile := defaultUserProfileService.Lookup("1111")
	assert.Equal(t, UserProfile{}, profile)
}

func TestSaveWithEmptyID(t *testing.T) {

	decisionKey := NewUserDecisionKey("1111")
	savedUserProfile := UserProfile{
		ID:                  "",
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: "2222"},
	}
	defaultUserProfileService := DefaultUserProfileService{}
	defaultUserProfileService.Save(savedUserProfile)
	fetchedProfile := defaultUserProfileService.Lookup("")
	assert.Equal(t, UserProfile{}, fetchedProfile)
}

func TestSaveProfileWithValidID(t *testing.T) {

	decisionKey := NewUserDecisionKey("1111")
	savedUserProfile := UserProfile{
		ID:                  testUserContext.ID,
		ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: "2222"},
	}
	defaultUserProfileService := DefaultUserProfileService{}
	defaultUserProfileService.Save(savedUserProfile)
	fetchedProfile := defaultUserProfileService.Lookup(testUserContext.ID)
	assert.Equal(t, savedUserProfile, fetchedProfile)
}

func TestSaveAndLookupMultipleTimes(t *testing.T) {

	defaultUserProfileService := DefaultUserProfileService{}
	// Creating Profiles
	var profiles []UserProfile
	for i := 1; i <= 5; i++ {
		decisionKey := NewUserDecisionKey("1111")
		userProfile := UserProfile{
			ID:                  strconv.Itoa(i),
			ExperimentBucketMap: map[UserDecisionKey]string{decisionKey: "2222"},
		}
		profiles = append(profiles, userProfile)
	}
	// Saving Profiles
	for _, profile := range profiles {
		defaultUserProfileService.Save(profile)
	}
	// Looking up Profiles
	for _, profile := range profiles {
		fetchedProfile := defaultUserProfileService.Lookup(profile.ID)
		assert.Equal(t, profile, fetchedProfile)
	}
}
