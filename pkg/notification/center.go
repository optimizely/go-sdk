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
	"github.com/optimizely/go-sdk/pkg/logging"
)

// Center handles all notification listeners. It keeps track of the Manager for each type of notification.
type Center interface {
	AddHandler(Type, func(interface{})) (int, error)
	RemoveHandler(int, Type) error
	Send(Type, interface{}) error
}

// DefaultCenter contains all the notification managers
type DefaultCenter struct {
	managerMap map[Type]Manager
}

// NewNotificationCenter returns a new notification center
func NewNotificationCenter() *DefaultCenter {
	decisionNotificationManager := NewAtomicManager(logging.GetLogger("", "AtomicManager"))
	projectConfigUpdateNotificationManager := NewAtomicManager(logging.GetLogger("", "AtomicManager"))
	processLogEventNotificationManager := NewAtomicManager(logging.GetLogger("", "AtomicManager"))
	trackNotificationManager := NewAtomicManager(logging.GetLogger("", "AtomicManager"))
	managerMap := make(map[Type]Manager)
	managerMap[Decision] = decisionNotificationManager
	managerMap[ProjectConfigUpdate] = projectConfigUpdateNotificationManager
	managerMap[LogEvent] = processLogEventNotificationManager
	managerMap[Track] = trackNotificationManager
	return &DefaultCenter{
		managerMap: managerMap,
	}
}

// AddHandler adds a handler for the given notification type
func (c *DefaultCenter) AddHandler(notificationType Type, handler func(interface{})) (int, error) {
	if manager, ok := c.managerMap[notificationType]; ok {
		return manager.Add(handler)
	}

	return -1, fmt.Errorf("no notification manager found for type %s", notificationType)
}

// RemoveHandler removes a handler for the given id and notification type
func (c *DefaultCenter) RemoveHandler(id int, notificationType Type) error {
	if manager, ok := c.managerMap[notificationType]; ok {
		manager.Remove(id)
		return nil
	}

	return fmt.Errorf("no notification manager found for type %s", notificationType)
}

// Send sends the given notification payload to all listeners of type
func (c *DefaultCenter) Send(notificationType Type, notification interface{}) error {
	if manager, ok := c.managerMap[notificationType]; ok {
		manager.Send(notification)
		return nil
	}

	return fmt.Errorf("no notification manager found for type %s", notificationType)
}
