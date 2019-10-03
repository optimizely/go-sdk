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
	"context"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueueEventDispatcher_DispatchEvent(t *testing.T) {
	ctx := context.TODO()
	q := NewQueueEventDispatcher(ctx)

	if qed, ok := q.(*QueueEventDispatcher); ok {
		qed.Dispatcher = &MockDispatcher{Events: NewInMemoryQueue(100)}
	}

	eventTags := map[string]interface{}{"revenue": 55.0, "value": 25.1}
	config := TestConfig{}

	conversionUserEvent := CreateConversionUserEvent(config, entities.Event{ExperimentIds: []string{"15402980349"}, ID: "15368860886", Key: "sample_conversion"}, userContext, eventTags)

	batch := createBatchEvent(conversionUserEvent, createVisitorFromUserEvent(conversionUserEvent))

	logEvent := createLogEvent(batch)

	qd, _ := q.(*QueueEventDispatcher)

	success, _ := qd.DispatchEvent(logEvent)

	assert.True(t, success)

	// its been queued
	assert.Equal(t,1, qd.eventQueue.Size())

	// give the queue a chance to run
	time.Sleep(1 * time.Second)

	// check the queue
	assert.Equal(t,0, qd.eventQueue.Size())

}
