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
	"fmt"
)

// Maintains mapping of SDK key to notification center
var notificationCenterRegistry = make(map[string]Center)

// RegisterHandler registers the given handler for the specified type against the given SDK key
func RegisterHandler(sdkKey string, notificationType Type, handler func(interface{})) (int, error) {
	var center Center
	var ok bool
	if center, ok = notificationCenterRegistry[sdkKey]; !ok {
		// create a notification center for the given SDK key if we do not find one
		center = NewNotificationCenter()
		notificationCenterRegistry[sdkKey] = center
	}

	return center.AddHandler(notificationType, handler)
}

// RemoveHandler removes the handle with the given ID of the specified type
func RemoveHandler(sdkKey string, notificationType Type, notificationID int) error {
	if center, ok := notificationCenterRegistry[sdkKey]; ok {
		return center.RemoveHandler(notificationID, notificationType)
	}

	return fmt.Errorf(`no notification center found for SDK Key "%s"`, sdkKey)
}

// Send sends the specified notification to handlers of the given type registered against the given SDK key
func Send(sdkKey string, notificationType Type, notification interface{}) error {
	if center, ok := notificationCenterRegistry[sdkKey]; ok {
		return center.Send(notificationType, notification)
	}

	return fmt.Errorf(`no notification center found for SDK Key "%s"`, sdkKey)
}
