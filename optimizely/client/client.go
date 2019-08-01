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
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

var logger = logging.GetLogger("Client")

// OptimizelyClient is the entry point to the Optimizely SDK
type OptimizelyClient struct {
	configManager   optimizely.ProjectConfigManager
	decisionService decision.DecisionService
	isValid         bool
}

// IsFeatureEnabled returns true if the feature is enabled for the given user
func (o *OptimizelyClient) IsFeatureEnabled(featureKey string, userContext entities.UserContext) (result bool, err error) {
	if !o.isValid {
		errorMessage := "Optimizely instance is not valid. Failing IsFeatureEnabled."
		err := errors.New(errorMessage)
		logger.Error(errorMessage, nil)
		return false, err
	}

	defer func() {
		if r := recover(); r != nil {
			errorMessage := fmt.Sprintf(`Optimizely SDK is panicking with the error "%s"`, string(debug.Stack()))
			err = errors.New(errorMessage)
			logger.Error(errorMessage, err)
			result = false
		}
	}()

	projectConfig := o.configManager.GetConfig()
	feature, err := projectConfig.GetFeatureByKey(featureKey)
	if err != nil {
		logger.Error("Error retrieving feature", err)
		return result, err
	}
	featureDecisionContext := decision.FeatureDecisionContext{
		Feature:       &feature,
		ProjectConfig: projectConfig,
	}

	userID := userContext.ID
	logger.Debug(fmt.Sprintf(`Evaluating feature "%s" for user "%s".`, featureKey, userID))
	featureDecision, err := o.decisionService.GetFeatureDecision(featureDecisionContext, userContext)
	if err != nil {
		logger.Error("Received an error while computing feature decision", err)
		return result, err
	}

	logger.Debug(fmt.Sprintf(`Decision made for feature "%s" for user "%s" with the following reason: "%s". Source: "%s".`, featureKey, userID, featureDecision.Reason, featureDecision.Source))

	if featureDecision.Variation.FeatureEnabled == true {
		result = true
		logger.Info(fmt.Sprintf(`Feature "%s" is enabled for user "%s".`, featureKey, userID))
	} else {
		logger.Info(fmt.Sprintf(`Feature "%s" is not enabled for user "%s".`, featureKey, userID))
	}

	// @TODO(mng): send impression event
	return result, nil
}
