/****************************************************************************
 * Copyright 2019-2025, Optimizely, Inc. and contributors                   *
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

package evaluator

import (
	"fmt"

	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/entities"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// CheckIfUserInAudience evaluates if user meets experiment audience conditions
func CheckIfUserInAudience(experiment *entities.Experiment, userContext entities.UserContext, projectConfig config.ProjectConfig, audienceEvaluator TreeEvaluator, options *decide.Options, logger logging.OptimizelyLogProducer) (bool, decide.DecisionReasons) {
	decisionReasons := decide.NewDecisionReasons(options)

	if experiment == nil {
		logMessage := decisionReasons.AddInfo("Experiment is nil, defaulting to false")
		logger.Debug(logMessage)
		return false, decisionReasons
	}

	if experiment.AudienceConditionTree != nil {
		condTreeParams := entities.NewTreeParameters(&userContext, projectConfig.GetAudienceMap())
		logger.Debug(fmt.Sprintf("Evaluating audiences for experiment \"%q\".", experiment.Key))

		evalResult, _, audienceReasons := audienceEvaluator.Evaluate(experiment.AudienceConditionTree, condTreeParams, options)
		decisionReasons.Append(audienceReasons)

		logMessage := decisionReasons.AddInfo("Audiences for experiment %s collectively evaluated to %v.", experiment.Key, evalResult)
		logger.Debug(logMessage)

		return evalResult, decisionReasons
	}

	logMessage := decisionReasons.AddInfo("Audiences for experiment %s collectively evaluated to true.", experiment.Key)
	logger.Debug(logMessage)
	return true, decisionReasons
}
