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

package optlyplugins

import (
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/tests/integration/models"
)

// TrackListenerManager helper for track notification
type TrackListenerManager struct {
	listenersCalled []models.TrackListener
}

// TrackCallback represents callback function for track notification
type TrackCallback = func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent)

// GetListenerCallbacks - Creates and returns track callback functions
func (c *TrackListenerManager) GetListenerCallbacks(listeners map[string]int) []TrackCallback {
	var callbackArray []TrackCallback
	for listenerType, count := range listeners {
		for i := 0; i < count; i++ {
			switch listenerType {
			case "Track":
				var callbackFunc TrackCallback = func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
					listener := models.TrackListener{
						EventKey:   eventKey,
						UserID:     userContext.ID,
						Attributes: userContext.Attributes,
						EventTags:  eventTags,
					}
					c.listenersCalled = append(c.listenersCalled, listener)
				}
				callbackArray = append(callbackArray, callbackFunc)
				break
			default:
				break
			}
		}
	}
	return callbackArray
}

// GetListenersCalled - Returns listeners called
func (c *TrackListenerManager) GetListenersCalled() []models.TrackListener {
	listenerCalled := c.listenersCalled
	// Since for every scenario, a new sdk instance is created, emptying listenersCalled is required for scenario's
	// where multiple requests are executed but no session is to be maintained among them.
	// @TODO: Make it optional once event-batching(sessioned) tests are implemented.
	c.listenersCalled = nil
	return listenerCalled
}
