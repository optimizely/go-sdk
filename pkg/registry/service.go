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

// Package registry is the global access point for retrieving instances of services by SDK Key //
package registry

import (
	"sync"

	"github.com/optimizely/go-sdk/pkg/notification"
)

var notificationCenterCache = make(map[string]notification.Center)
var notificationLock sync.Mutex

// GetNotificationCenter returns the notification center instance associated with the given SDK Key or creates a new one if not found
func GetNotificationCenter(sdkKey string) notification.Center {
	notificationLock.Lock()
	defer notificationLock.Unlock()

	notificationCenter, ok := notificationCenterCache[sdkKey]
	if !ok {
		notificationCenter = notification.NewNotificationCenter()
		notificationCenterCache[sdkKey] = notificationCenter
	}

	return notificationCenter
}
