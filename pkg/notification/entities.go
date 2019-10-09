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

// Package notification //
package notification

import (
	"github.com/optimizely/go-sdk/pkg/entities"
)

// Type is the type of notification
type Type string

// DecisionNotificationType is the type of decision notification
type DecisionNotificationType string

const (
	// Decision notification type
	Decision Type = "decision"
	// ProjectConfigUpdate notification type
	ProjectConfigUpdate Type = "project_config_update"

	// ABTest is used when the decision is returned as part of evaluating an ab test
	ABTest DecisionNotificationType = "ab-test"
	// Feature is used when the decision is returned as part of evaluating a feature
	Feature DecisionNotificationType = "feature"
	// LogEvent notification type
	LogEvent Type = "log_event_notification"
)

// DecisionNotification is a notification triggered when a decision is made for either a feature or an experiment
type DecisionNotification struct {
	Type         DecisionNotificationType
	UserContext  entities.UserContext
	DecisionInfo map[string]interface{}
}

// ProjectConfigUpdateNotification is a notification triggered when a project config is updated
type ProjectConfigUpdateNotification struct {
	Type     Type
	Revision string
}

// LogEventNotification is the notification triggered before log event is dispatched.
type LogEventNotification struct {
	Type         Type
	LogEvent  	 interface{}
}
