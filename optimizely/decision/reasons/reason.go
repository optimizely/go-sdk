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

package reasons

// Reason is the reason for which a decision was made
type Reason string

const (
	// BucketedVariationNotFound - the bucketed variation ID is not in the config
	BucketedVariationNotFound Reason = "Bucketed variation not found"
	// BucketedIntoVariation - the user is bucketed into a variation for the given experiment
	BucketedIntoVariation Reason = "Bucketed into variation"
	// DoesNotMeetRolloutTargeting - the user does not meet the rollout targeting rules
	DoesNotMeetRolloutTargeting Reason = "Does not meet rollout targeting"
	// DoesNotQualify - the user did not qualify for the experiment
	DoesNotQualify Reason = "Does not qualify for experiment"
	// NoRolloutForFeature - there is no rollout for the given feature
	NoRolloutForFeature Reason = "No rollout for feature"
	// RolloutHasNoExperiments - the rollout has no assigned experiments
	RolloutHasNoExperiments Reason = "Rollout has no experiments"
	// NotBucketedIntoVariation - the user is not bucketed into a variation for the given experiment
	NotBucketedIntoVariation Reason = "Not bucketed into a variation"
	// NotInGroup - the user is not bucketed into the mutex group
	NotInGroup Reason = "Not bucketed into any experiment in mutex group"
)
