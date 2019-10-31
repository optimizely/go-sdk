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

var pesLogger = logging.GetLogger("pkg/decision/persisting_experiment_service")

// PersistingExperimentService attempts to retrieve a saved decision from the user profile service
// for the user before having the ExperimentBucketerService compute it.
// If computed, the decision is saved back to the user profile service if provided.
type PersistingExperimentService struct {
	experimentBucketedService ExperimentService
	userProfileService        UserProfileService
}

// PESOptionFunc is used to extend the PersistingExperimentService with additional options
type PESOptionFunc func(*PersistingExperimentService)

// WithUserProfileService sets the user profile service on the persistingExperimentService instance
func WithUserProfileService(userProfileService UserProfileService) PESOptionFunc {
	return func(f *PersistingExperimentService) {
		f.userProfileService = userProfileService
	}
}

// NewPersistingExperimentService returns a new instance of the PersistingExperimentService
func NewPersistingExperimentService(experimentBucketerService ExperimentService, options ...PESOptionFunc) *PersistingExperimentService {
	persistingExperimentService := &PersistingExperimentService{
		experimentBucketedService: experimentBucketerService,
	}

	for _, opt := range options {
		opt(persistingExperimentService)
	}

	return persistingExperimentService
}

// GetDecision returns the decision with the variation the user is bucketed into
func (p PersistingExperimentService) GetDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, error) {
	var experimentDecision ExperimentDecision
	var err error
	var userProfile UserProfile
	// check to see if there is a saved decision for the user
	if p.userProfileService != nil {
		experimentDecision, userProfile = p.getSavedDecision(decisionContext, userContext)
		if experimentDecision.Variation != nil {
			return experimentDecision, nil
		}
	}

	experimentDecision, err = p.experimentBucketedService.GetDecision(decisionContext, userContext)
	if experimentDecision.Variation != nil && p.userProfileService != nil {
		// save decision if a user profile service is provided
		p.saveDecision(userProfile, decisionContext.Experiment, experimentDecision)
	}

	return experimentDecision, err
}

func (p PersistingExperimentService) getSavedDecision(decisionContext ExperimentDecisionContext, userContext entities.UserContext) (ExperimentDecision, UserProfile) {
	experimentDecision := ExperimentDecision{}

	if p.userProfileService == nil {
		return experimentDecision, UserProfile{}
	}

	userProfile, err := p.userProfileService.Lookup(userContext.ID)
	if err != nil {
		errMessage := fmt.Sprintf(`Error looking up user from user profile service: %s`, err)
		bLogger.Warning(errMessage)
		return experimentDecision, UserProfile{}
	}

	// look up experiment decision from user profile
	decisionKey := UserDecisionKey{ExperimentID: decisionContext.Experiment.ID}
	if savedVariationID, ok := userProfile.ExperimentBucketMap[decisionKey]; ok {
		if variation, ok := decisionContext.Experiment.Variations[savedVariationID]; ok {
			experimentDecision.Variation = &variation
			bLogger.Debug(fmt.Sprintf(`User "%s" was previously bucketed into variation "%s" of experiment "%s".`, userContext.ID, variation.Key, decisionContext.Experiment.Key))
		} else {
			bLogger.Warning(fmt.Sprintf(`User "%s" was previously bucketed into variation with ID "%s" for experiment "%s", but no matching variation was found.`, userContext.ID, savedVariationID, decisionContext.Experiment.Key))
		}
	}

	return experimentDecision, userProfile
}

func (p PersistingExperimentService) saveDecision(userProfile UserProfile, experiment *entities.Experiment, decision ExperimentDecision) {
	if p.userProfileService != nil {
		decisionKey := UserDecisionKey{ExperimentID: experiment.ID}
		userProfile.ExperimentBucketMap[decisionKey] = decision.Variation.ID
		err := p.userProfileService.Save(userProfile)
		if err != nil {
			bLogger.Error(`Unable to save decision to user profile service`, err)
		}
	}
}
