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
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	configManager   config.ProjectConfigManager
	decisionService decision.DecisionService
	isValid         bool
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
func (optly *OptimizelyClient) IsFeatureEnabled(featureKey string, userID string, attributes map[string]interface{}) bool {
	if !optly.isValid {
		logging.Error("Optimizely instance is not valid. Failing IsFeatureEnabled.")
		return false
	}

	userContext := entities.UserContext{ID: userID, Attributes: entities.UserAttributes{Attributes: attributes}}

	// @TODO(mng): we should fetch the Feature entity from the config service instead of manually creating it here
	featureExperiment := entities.Experiment{}
	feature := entities.Feature{
		Key:                featureKey,
		FeatureExperiments: []entities.Experiment{featureExperiment},
	}
	featureDecisionContext := decision.FeatureDecisionContext{
		Feature: feature,
	}

	featureDecision, err := optly.decisionService.GetFeatureDecision(featureDecisionContext, userContext)
	if err != nil {
		// @TODO(mng): log error
		return false
	}
	return featureDecision.FeatureEnabled
}
