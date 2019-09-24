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
package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type managerMockReceiver struct {
	mock.Mock
}

func (m *managerMockReceiver) handle(notification interface{}) {
	m.Called(notification)
}

func (m *managerMockReceiver) handleBetter(notification interface{}) {
	m.Called(notification)
}

func TestAtomicManager(t *testing.T) {
	payload := map[string]interface{}{
		"key": "test",
	}

	mockReceiver := new(managerMockReceiver)
	mockReceiver.On("handle", payload)
	mockReceiver.On("handleBetter", payload)

	atomicManager := NewAtomicManager()
	result1, _ := atomicManager.Add(mockReceiver.handle)
	assert.Equal(t, 1, result1)

	result2, _ := atomicManager.Add(mockReceiver.handleBetter)
	assert.Equal(t, 2, result2)

	atomicManager.Send(payload)
	mockReceiver.AssertExpectations(t)

	atomicManager.Remove(result1)
	atomicManager.Send(payload)
	mockReceiver.AssertNumberOfCalls(t, "handle", 1)
	mockReceiver.AssertNumberOfCalls(t, "handleBetter", 2)

	atomicManager.Remove(result2)
	atomicManager.Send(payload)
	mockReceiver.AssertNumberOfCalls(t, "handleBetter", 2)

	// Sanity check by calling remove with a incorrect handler id
	atomicManager.Remove(55)
}
