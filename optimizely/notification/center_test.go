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

	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/mock"
)

type MockReceiver struct {
	mock.Mock
}

func (m *MockReceiver) handleNotification(notification interface{}) {
	m.Called(notification)
}

type NotificationCenterTestSuite struct {
	suite.Suite
}

func (s *NotificationCenterTestSuite) AfterTest(suiteName, testName string) {
	ClearAllHandlers()
}

func (s *NotificationCenterTestSuite) TestAddHandler() {
	mockReceiver := new(MockReceiver)
	mockReceiver2 := new(MockReceiver)
	id1, e1 := AddHandler("sdk_key_1", Decision, mockReceiver.handleNotification)
	s.Equal(1, id1)
	s.NoError(e1)

	id2, e2 := AddHandler("sdk_key_2", Decision, mockReceiver2.handleNotification)
	s.Equal(1, id2)
	s.NoError(e2)
}

func (s *NotificationCenterTestSuite) TestRemoveHandler() {
	mockReceiver := new(MockReceiver)
	mockReceiver2 := new(MockReceiver)
	id1, _ := AddHandler("sdk_key_1", Decision, mockReceiver.handleNotification)
	id2, _ := AddHandler("sdk_key_2", Decision, mockReceiver2.handleNotification)

	e1 := RemoveHandler("sdk_key_1", id1, Decision)
	s.NoError(e1)
	e2 := RemoveHandler("sdk_key_2", id2, Decision)
	s.NoError(e2)
}

func TestNotificationCenterTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationCenterTestSuite))
}
