package notification

import (
	"testing"

	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/mock"
)

type MockReceiver struct {
	mock.Mock
}

func (m *MockReceiver) handleNotification(notification interface{}) {
	m.Called(notification)
}

func TestNotificationCenterAddHandler(t *testing.T) {
	mockReceiver := new(MockReceiver)

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
	notificationCenter := NewNotificationCenter()
	notificationCenter.AddHandler(Decision, mockReceiver.handleNotification)
	notificationCenter.Send(Decision, testDecisionNotification)

	mockReceiver.AssertExpectations(t)
}
