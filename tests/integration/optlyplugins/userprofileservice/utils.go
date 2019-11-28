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
	"github.com/optimizely/go-sdk/pkg"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// UPSHelper defines Helper methods for UPS
type UPSHelper interface {
	SaveUserProfiles(userProfiles []decision.UserProfile)
	GetUserProfiles() (savedProfiles []decision.UserProfile)
}

// CreateUserProfileService creates a user profile service with the given parameters
func CreateUserProfileService(config pkg.ProjectConfig, apiOptions models.APIOptions) decision.UserProfileService {
	var userProfileService decision.UserProfileService
	switch apiOptions.UserProfileServiceType {
	case "NormalService":
		userProfileService = new(NormalUserProfileService)
		break
	case "LookupErrorService":
		userProfileService = new(LookupErrorUserProfileService)
		break
	case "SaveErrorService":
		userProfileService = new(SaveErrorUserProfileService)
		break
	default:
		userProfileService = new(NoOpUserProfileService)
		break
	}

	var profilesArray []decision.UserProfile
	for userID, bucketMap := range apiOptions.UPSMapping {
		var profile decision.UserProfile
		profile.ID = userID
		profile.ExperimentBucketMap = make(map[decision.UserDecisionKey]string)
		for experimentKey, variationKey := range bucketMap {
			if experiment, err := config.GetExperimentByKey(experimentKey); err == nil {
				decisionKey := decision.NewUserDecisionKey(experiment.ID)
				if variation, ok := experiment.VariationsKeyMap[variationKey]; ok {
					profile.ExperimentBucketMap[decisionKey] = variation.ID
				}
			}
		}
		profilesArray = append(profilesArray, profile)
	}
	userProfileService.(UPSHelper).SaveUserProfiles(profilesArray)
	return userProfileService
}

// ParseUserProfiles converts raw profiles into an array of user profiles
func ParseUserProfiles(rawProfiles []map[string]interface{}) (parsedProfiles []decision.UserProfile) {
	for _, profile := range rawProfiles {
		userProfile := decision.UserProfile{}
		if userID, ok := profile["user_id"]; ok {
			userProfile.ID = userID.(string)
		}
		if experimentBucketMap, ok := profile["experiment_bucket_map"]; ok {
			userProfile.ExperimentBucketMap = make(map[decision.UserDecisionKey]string)
			for k, v := range experimentBucketMap.(map[string]interface{}) {
				decisionKey := decision.NewUserDecisionKey(k)
				if bucketMap, ok := v.(map[string]interface{}); ok {
					userProfile.ExperimentBucketMap[decisionKey] = bucketMap[decisionKey.Field].(string)
				}
			}
		}
		parsedProfiles = append(parsedProfiles, userProfile)
	}
	return parsedProfiles
}
