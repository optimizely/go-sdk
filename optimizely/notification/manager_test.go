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
	result, _ := atomicManager.AddHandler(mockReceiver.handle)
	assert.Equal(t, 1, result)

	result, _ = atomicManager.AddHandler(mockReceiver.handleBetter)
	assert.Equal(t, 2, result)

	atomicManager.Send(payload)
	mockReceiver.AssertExpectations(t)
}
