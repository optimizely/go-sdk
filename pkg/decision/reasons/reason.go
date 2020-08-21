/****************************************************************************
 * Copyright 2019-2020, Optimizely, Inc. and contributors                        *
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

// Package reasons //
package reasons

// Reason is the reason for which a decision was made
type Reason string

const (
	// AttributeFormatInvalid - invalid format for attributes
	AttributeFormatInvalid Reason = "Provided attributes are in an invalid format."
	// BucketedVariationNotFound - the bucketed variation ID is not in the config
	BucketedVariationNotFound Reason = "Bucketed variation not found"
	// BucketedIntoVariation - the user is bucketed into a variation for the given experiment
	BucketedIntoVariation Reason = "Bucketed into variation"
	// BucketedIntoFeatureTest - the user is bucketed into a variation for the given feature test
	BucketedIntoFeatureTest Reason = "Bucketed into feature test"
	// BucketedIntoRollout - the user is bucketed into a variation for the given feature rollout
	BucketedIntoRollout Reason = "Bucketed into feature rollout"
	// FailedRolloutBucketing - the user is not bucketed into the feature rollout
	FailedRolloutBucketing Reason = "Not bucketed into rollout"
	// FailedRolloutTargeting - the user does not meet the rollout targeting rules
	FailedRolloutTargeting Reason = "Does not meet rollout targeting rule"
	// FailedAudienceTargeting - the user failed the audience targeting conditions
	FailedAudienceTargeting Reason = "Does not meet audience targeting conditions"
	// NoRolloutForFeature - there is no rollout for the given feature
	NoRolloutForFeature Reason = "No rollout for feature"
	// RolloutHasNoExperiments - the rollout has no assigned experiments
	RolloutHasNoExperiments Reason = "Rollout has no experiments"
	// NotBucketedIntoVariation - the user is not bucketed into a variation for the given experiment
	NotBucketedIntoVariation Reason = "Not bucketed into a variation"
	// NotInGroup - the user is not bucketed into the mutex group
	NotInGroup Reason = "Not bucketed into any experiment in mutex group"
	// NoWhitelistVariationAssignment - there is no variation assignment for the given user and experiment
	NoWhitelistVariationAssignment Reason = "No whitelist variation assignment"
	// InvalidWhitelistVariationAssignment - A variation assignment was found for the given user and experiment, but no variation with that key exists in the given experiment
	InvalidWhitelistVariationAssignment Reason = "Invalid whitelist variation assignment"
	// WhitelistVariationAssignmentFound - a valid variation assignment was found for the given user and experiment
	WhitelistVariationAssignmentFound Reason = "Whitelist variation assignment found"
	// NoOverrideVariationAssignment - No override variation was found for the given user and experiment
	NoOverrideVariationAssignment Reason = "No override variation assignment"
	// InvalidOverrideVariationAssignment - An override variation was found for the given user and experiment, but no variation with that key exists in the given experiment
	InvalidOverrideVariationAssignment Reason = "Invalid override variation assignment"
	// OverrideVariationAssignmentFound - A valid override variation was found for the given user and experiment
	OverrideVariationAssignmentFound Reason = "Override variation assignment found"
)
