package event

import (
	"context"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueueEventDispatcher_DispatchEvent(t *testing.T) {
	ctx := context.TODO()
	q := NewQueueEventDispatcher(ctx)

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))

	logEvent := createLogEvent(batch)

	qd, _ := q.(*QueueEventDispatcher)

	qd.DispatchEvent(logEvent, func(success bool) {
		assert.True(t, success)
	})

	// its been queued
	assert.Equal(t,1, qd.eventQueue.Size())

	// give the queue a chance to run
	time.Sleep(1 * time.Second)

	// check the queue
	assert.Equal(t,0, qd.eventQueue.Size())

}
