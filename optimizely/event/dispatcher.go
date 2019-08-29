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

// Package event //
package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

const jsonContentType = "application/json"

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent, callback func(success bool))
}

// HTTPEventDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPEventDispatcher struct {
}

var dispatcherLogger = logging.GetLogger("EventDispatcher")

// DispatchEvent dispatches event with callback
func (*HTTPEventDispatcher) DispatchEvent(event LogEvent, callback func(bool)) {

	jsonValue, _ := json.Marshal(event.event)
	resp, err := http.Post(event.endPoint, jsonContentType, bytes.NewBuffer(jsonValue))
	// also check response codes
	// resp.StatusCode == 400 is an error
	var success bool
	if err != nil {
		dispatcherLogger.Error("http.Post failed:", err)
		success = false
	} else {
		if resp.StatusCode == 204 {
			success = true
		} else {
			fmt.Printf("http.Post invalid response %d", resp.StatusCode)
			success = false
		}
	}
	callback(success)

}
