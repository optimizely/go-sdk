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

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
)

type MockReceiver struct {
	mock.Mock
}

func (m *MockReceiver) handleNotification(notification interface{}) {
	m.Called(notification)
}

func TestNotificationCenter(t *testing.T) {
	mockReceiver := new(MockReceiver)
	mockReceiver2 := new(MockReceiver)

	testUser := entities.UserContext{}
	testDecisionNotification := DecisionNotification{
		Type:        Feature,
		UserContext: testUser,
		DecisionInfo: map[string]interface{}{
			"feature": map[string]interface{}{
				"source": "Rollout",
			},
		},
	}
	mockReceiver.On("handleNotification", testDecisionNotification)
	mockReceiver2.On("handleNotification", testDecisionNotification)
	notificationCenter := NewNotificationCenter()
	id1, err1 := notificationCenter.AddHandler(Decision, mockReceiver.handleNotification)
	id2, err2 := notificationCenter.AddHandler(Decision, mockReceiver2.handleNotification)
	notificationCenter.Send(Decision, testDecisionNotification)

	mockReceiver.AssertExpectations(t)
	mockReceiver2.AssertExpectations(t)
	assert.Nil(t, err1)
	assert.Nil(t, err2)

	notificationCenter.RemoveHandler(id1, Decision)
	notificationCenter.Send(Decision, testDecisionNotification)

	mockReceiver.AssertNumberOfCalls(t, "handleNotification", 1)
	mockReceiver2.AssertNumberOfCalls(t, "handleNotification", 2)

	notificationCenter.RemoveHandler(id2, Decision)
	notificationCenter.Send(Decision, testDecisionNotification)

	mockReceiver.AssertNumberOfCalls(t, "handleNotification", 1)
	mockReceiver2.AssertNumberOfCalls(t, "handleNotification", 2)
}
