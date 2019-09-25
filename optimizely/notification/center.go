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

import "fmt"

// Maintains a map of notification managers mapped by SDK Key and Type
var managerMaps = make(map[string]map[Type]Manager)

// AddHandler adds a handler for the given notification type
func AddHandler(sdkKey string, notificationType Type, handler func(interface{})) (int, error) {
	var managerMap map[Type]Manager
	var ok bool
	if managerMap, ok = managerMaps[sdkKey]; !ok {
		// create a managerMap for the given SDK key if there isn't one yet
		managerMap = make(map[Type]Manager)
		managerMaps[sdkKey] = managerMap
	}

	var manager Manager
	if manager, ok = managerMap[notificationType]; !ok {
		// create a manager for the given notification type if there isn't one yet
		manager = NewAtomicManager()
		managerMap[notificationType] = manager
	}

	return manager.Add(handler)
}

// RemoveHandler removes a handler for the given id and notification type
func RemoveHandler(sdkKey string, id int, notificationType Type) error {
	var managerMap map[Type]Manager
	var ok bool
	if managerMap, ok = managerMaps[sdkKey]; !ok {
		return fmt.Errorf(`no notification managers found for SDK Key %s`, sdkKey)
	}

	if manager, ok := managerMap[notificationType]; ok {
		manager.Remove(id)
		return nil
	}

	return fmt.Errorf("no notification manager found for type %s", notificationType)
}

// Send sends the given notification payload to all listeners of type
func Send(sdkKey string, notificationType Type, notification interface{}) error {
	var managerMap map[Type]Manager
	var ok bool
	if managerMap, ok = managerMaps[sdkKey]; !ok {
		return fmt.Errorf(`no notification managers found for SDK Key %s`, sdkKey)
	}

	if manager, ok := managerMap[notificationType]; ok {
		manager.Send(notification)
		return nil
	}

	return fmt.Errorf("no notification manager found for type %s", notificationType)
}

// ClearAllHandlers removes all registered notification handlers
func ClearAllHandlers() {
	// reset to an empty map
	managerMaps = make(map[string]map[Type]Manager)
}
