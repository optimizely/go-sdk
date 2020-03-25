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

import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/logging"
)

// PersistingExperimentService attempts to retrieve a saved decision from the user profile service
// for the user before having the ExperimentBucketerService compute it.
// If computed, the decision is saved back to the user profile service if provided.
type PersistingExperimentService struct {
	experimentBucketedService ExperimentService
	userProfileService        UserProfileService
	logger 					  logging.OptimizelyLogProducer
}

// NewPersistingExperimentService returns a new instance of the PersistingExperimentService
func NewPersistingExperimentService(userProfileService UserProfileService, experimentBucketerService ExperimentService, logger logging.OptimizelyLogProducer) *PersistingExperimentService {
	persistingExperimentService := &PersistingExperimentService{
		logger: logger,
		experimentBucketedService: experimentBucketerService,
		userProfileService:        userProfileService,
	}

	return persistingExperimentService
}

// GetDecision returns the decision with the variation the user is bucketed into
func (p PersistingExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (experimentDecision ExperimentDecision, err error) {
	if p.userProfileService == nil {
		return p.experimentBucketedService.GetDecision(decisionContext, userContext)
	}

	var userProfile UserProfile
	// check to see if there is a saved decision for the user
	experimentDecision, userProfile = p.getSavedDecision(decisionContext, userContext)
	if experimentDecision.Variation != nil {
		return experimentDecision, nil
	}

	experimentDecision, err = p.experimentBucketedService.GetDecision(decisionContext, userContext)
	if experimentDecision.Variation != nil {
		// save decision if a user profile service is provided
		userProfile.ID = userContext.ID
		p.saveDecision(userProfile, decisionContext.Experiment, experimentDecision)
	}

	return experimentDecision, err
}

func (p PersistingExperimentService) getSavedDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, UserProfile) {
	experimentDecision := ExperimentDecision{}
	userProfile := p.userProfileService.Lookup(userContext.ID)

	// look up experiment decision from user profile
	decisionKey := NewUserDecisionKey(decisionContext.Experiment.ID)
	if userProfile.ExperimentBucketMap == nil {
		return experimentDecision, userProfile
	}

	if savedVariationID, ok := userProfile.ExperimentBucketMap[decisionKey]; ok {
		if variation, ok := decisionContext.Experiment.Variations[savedVariationID]; ok {
			experimentDecision.Variation = &variation
			p.logger.Debug(fmt.Sprintf(`User "%s" was previously bucketed into variation "%s" of experiment "%s".`, userContext.ID, variation.Key, decisionContext.Experiment.Key))
		} else {
			p.logger.Warning(fmt.Sprintf(`User "%s" was previously bucketed into variation with ID "%s" for experiment "%s", but no matching variation was found.`, userContext.ID, savedVariationID, decisionContext.Experiment.Key))
		}
	}

	return experimentDecision, userProfile
}

func (p PersistingExperimentService) saveDecision(userProfile UserProfile, experiment *entities.Experiment, decision ExperimentDecision) {
	if p.userProfileService != nil {
		decisionKey := NewUserDecisionKey(experiment.ID)
		if userProfile.ExperimentBucketMap == nil {
			userProfile.ExperimentBucketMap = map[UserDecisionKey]string{}
		}
		userProfile.ExperimentBucketMap[decisionKey] = decision.Variation.ID
		p.userProfileService.Save(userProfile)
		p.logger.Debug(fmt.Sprintf(`Decision saved for user "%s".`, userProfile.ID))
	}
}
