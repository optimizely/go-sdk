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

func TestSendtRaceCondition(t *testing.T) {
	sync := make(chan interface{})
	payload := map[string]interface{}{
		"key": "test",
	}
	atomicManager := NewAtomicManager()
	result1, result2 := 0, 0
	listenerCalled := false

	listener1 := func(interface{}) {
	}

	listener2 := func(interface{}) {
		// Add listener2 internally to assert deadlock
		result2, _ = atomicManager.Add(listener1)
		// Remove all added listeners
		atomicManager.Remove(result1)
		atomicManager.Remove(result2)
		listenerCalled = true
	}
	result1, _ = atomicManager.Add(listener2)

	go func() {
		atomicManager.Send(payload)
		// notifying that notification is sent.
		sync <- ""
	}()

	<-sync

	assert.Equal(t, 1, result1)
	assert.Equal(t, 2, result2)
	assert.Equal(t, true, listenerCalled)
}

func TestAddRaceCondition(t *testing.T) {
	sync := make(chan interface{})
	atomicManager := NewAtomicManager()

	listener1 := func(interface{}) {

	}
	result1, _ := atomicManager.Add(listener1)
	go func() {
		atomicManager.Remove(result1)
		sync <- ""
	}()

	<-sync
	assert.Equal(t, len(atomicManager.handlers), 0)
}
