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
