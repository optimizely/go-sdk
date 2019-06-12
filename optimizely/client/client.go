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

package client

import (
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	decisionEngine decision.Engine
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
func (optly *OptimizelyClient) IsFeatureEnabled(featureKey string, userID string, attributes map[string]interface{}) bool {
	// @TODO(mng): we should fetch the Feature entity from the config service instead of manually creating it here
	feature := entities.Feature{Key: featureKey}
	userContext := entities.UserContext{ID: userID, Attributes: attributes}
	featureDecision := optly.decisionEngine.GetFeatureDecision(feature, userContext)
	return featureDecision.FeatureEnabled
}
