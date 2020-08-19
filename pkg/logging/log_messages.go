/****************************************************************************
 * Copyright 2020, Optimizely, Inc. and contributors                   		*
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

// Package logging //
package logging

// LogMessage defines string type for log messages
type LogMessage string

func (l LogMessage) String() string {
	return string(l)
}

const (
	// Debug Logs

	// AudienceEvaluationStarted when single audience evaluation is started
	AudienceEvaluationStarted LogMessage = `Starting to evaluate audience "%s".`
	// AudienceEvaluatedTo when single audience evaluation is completed
	AudienceEvaluatedTo LogMessage = `Audience "%s" evaluated to %t.`
	// ExperimentAudiencesEvaluatedTo when collective audience evaluation for experiment is completed
	ExperimentAudiencesEvaluatedTo LogMessage = `Audiences for experiment %s collectively evaluated to %t.`
	// RolloutAudiencesEvaluatedTo when collective audience evaluation for rule is completed
	RolloutAudiencesEvaluatedTo LogMessage = `Audiences for rule %s collectively evaluated to %t.`
	// EvaluatingAudiencesForExperiment when audience evaluation is started for an experiment
	EvaluatingAudiencesForExperiment LogMessage = `Evaluating audiences for experiment "%s".`
	// EvaluatingAudiencesForRollout when audience evaluation is started for a rule
	EvaluatingAudiencesForRollout LogMessage = `Evaluating audiences for rule %s.".`
	// NullUserAttribute when user attribute is missing or nil
	NullUserAttribute LogMessage = `Audience condition %s evaluated to UNKNOWN because a null value was passed for user attribute "%s".`
	// UserInEveryoneElse when user is in last rule
	UserInEveryoneElse LogMessage = `User "%s" meets conditions for targeting rule "Everyone Else".`
	// UserNotInRollout when user is not in rollout/rule
	UserNotInRollout LogMessage = `User "%s" does not meet conditions for targeting rule %s.`
	// UserNotInExperiment when user is not in experiment
	UserNotInExperiment LogMessage = `User "%s" does not meet conditions to be in experiment "%s".`

	// Warning logs

	// UnknownConditionType when when condition type is unknown
	UnknownConditionType LogMessage = `Audience condition "%s" uses an unknown condition type. You may need to upgrade to a newer release of the Optimizely SDK.`
	// UnknownMatchType when match type is unknown
	UnknownMatchType LogMessage = `Audience condition "%s" uses an unknown match type. You may need to upgrade to a newer release of the Optimizely SDK.`
	// UnsupportedConditionValue when condition value is unsupported
	UnsupportedConditionValue LogMessage = `Audience condition "%s" has an unsupported condition value. You may need to upgrade to a newer release of the Optimizely SDK.`
	// InvalidAttributeValueType when user attribute value is invalid
	InvalidAttributeValueType LogMessage = `Audience condition "%s" evaluated to UNKNOWN because a value of type "%T" was passed for user attribute "%s".`
)
