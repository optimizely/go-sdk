/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
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
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
	"github.com/optimizely/go-sdk/pkg/odp/config"
	"github.com/optimizely/go-sdk/pkg/odp/utils"
	pkgUtils "github.com/optimizely/go-sdk/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type EventManagerTestSuite struct {
	suite.Suite
	eventManager    *BatchEventManager
	eventAPIManager *MockEventAPIManager
}

func (e *EventManagerTestSuite) SetupTest() {
	e.eventAPIManager = &MockEventAPIManager{}
	e.eventManager = NewBatchEventManager(WithAPIManager(e.eventAPIManager), WithBatchSize(2), WithOdpConfig(config.NewConfig("a", "b", []string{"c"})))
}

func (e *EventManagerTestSuite) TestEventManagerWithOptions() {
	batchSize := 1
	queueSize := 2
	flushInterval := 3 * time.Second
	sdkKey := "abc"
	eventAPIManager := NewEventAPIManager(sdkKey, nil)
	conf := config.NewConfig("", "", nil)
	eventQueue := event.NewInMemoryQueue(queueSize)
	em := NewBatchEventManager(WithBatchSize(batchSize), WithQueueSize(queueSize), WithFlushInterval(flushInterval), WithSDKKey(sdkKey),
		WithAPIManager(eventAPIManager), WithOdpConfig(conf), WithQueue(eventQueue))
	e.Equal(batchSize, em.batchSize)
	e.Equal(queueSize, em.maxQueueSize)
	e.Equal(flushInterval, em.flushInterval)
	e.Equal(sdkKey, em.sdkKey)
	e.Equal(eventAPIManager, em.apiManager)
	e.Equal(conf, em.OdpConfig)
	e.Equal(eventQueue, em.eventQueue)
}

func (e *EventManagerTestSuite) TestEventManagerWithInvalidOptions() {
	batchSize := -1
	queueSize := -1
	flushInterval := -1 * time.Second
	em := NewBatchEventManager(WithBatchSize(batchSize), WithQueueSize(queueSize), WithFlushInterval(flushInterval))
	e.Equal(utils.DefaultBatchSize, em.batchSize)
	e.Equal(utils.DefaultEventQueueSize, em.maxQueueSize)
	e.Equal(utils.DefaultEventFlushInterval, em.flushInterval)
}

func (e *EventManagerTestSuite) TestEventManagerWithoutOptions() {
	em := NewBatchEventManager()
	e.Equal(utils.DefaultBatchSize, em.batchSize)
	e.Equal(utils.DefaultEventQueueSize, em.maxQueueSize)
	e.Equal(utils.DefaultEventFlushInterval, em.flushInterval)
	e.Equal("", em.sdkKey)
	e.NotNil(em.apiManager)
	e.NotNil(em.eventQueue)
}

func (e *EventManagerTestSuite) TestTickerNotStartedIfODPNotIntegrated() {
	eg := newExecutionContext()
	e.eventManager.OdpConfig = config.NewConfig("", "", nil)
	eg.Go(e.eventManager.Start)
	eg.TerminateAndWait()
	e.Nil(e.eventManager.ticker)
}

func (e *EventManagerTestSuite) TestTickerStartedIfODPIntegrated() {
	eg := newExecutionContext()
	eg.Go(e.eventManager.Start)
	eg.TerminateAndWait()
	e.NotNil(e.eventManager.ticker)
}

func (e *EventManagerTestSuite) TestTickerIsNotReinitializedIfStartIsCalledAgain() {
	eg := newExecutionContext()
	eg.Go(e.eventManager.Start)
	eg.TerminateAndWait()
	e.NotNil(e.eventManager.ticker)
	ticker := e.eventManager.ticker

	eg.Go(e.eventManager.Start)
	eg.TerminateAndWait()
	e.Equal(ticker, e.eventManager.ticker)
}

func (e *EventManagerTestSuite) TestEventsDispatchedWhenContextIsTerminated() {
	eg := newExecutionContext()
	e.eventManager.eventQueue.Add(Event{})
	e.eventAPIManager.wg.Add(1)
	e.Equal(1, e.eventManager.eventQueue.Size())
	eg.Go(e.eventManager.Start)
	eg.TerminateAndWait()
	e.NotNil(e.eventManager.ticker)
	e.Equal(0, e.eventManager.eventQueue.Size())
	e.Equal(1, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestEventsDispatchedWhenFlushIntervalReached() {
	eg := newExecutionContext()
	e.eventManager.flushInterval = 50 * time.Millisecond
	e.eventManager.eventQueue.Add(Event{})
	e.eventAPIManager.wg.Add(1)
	e.Equal(1, e.eventManager.eventQueue.Size())
	eg.Go(e.eventManager.Start)
	e.eventAPIManager.wg.Wait()
	eg.TerminateAndWait()
	e.Equal(0, e.eventManager.eventQueue.Size())
	e.Equal(1, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestIdentifyUserWhenODPNotIntegrated() {
	e.eventManager.OdpConfig = config.NewConfig("", "", nil)
	e.eventManager.IdentifyUser("123")
	e.Nil(e.eventManager.ticker)
	e.Equal(0, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestIdentifyUserWhenODPIntegrated() {
	userID := "123"
	expectedEvent := Event{Identifiers: map[string]string{utils.OdpFSUserIDKey: userID}, Type: utils.OdpEventType, Action: utils.OdpActionIdentified}
	e.eventManager.addCommonData(&expectedEvent)
	e.eventAPIManager.wg.Add(1)
	e.eventManager.batchSize = 1
	e.eventManager.IdentifyUser(userID)
	e.eventAPIManager.wg.Wait()
	e.Equal(1, e.eventAPIManager.timesSendEventsCalled)

	actualEvent := e.eventAPIManager.eventsSent[0]
	e.NotNil(actualEvent.Data["idempotence_id"])
	// Making idempotence_id similar for both for comparison purposes
	actualEvent.Data["idempotence_id"] = expectedEvent.Data["idempotence_id"]
	e.Equal(expectedEvent, actualEvent)
}

func (e *EventManagerTestSuite) TestProcessEventWithInvalidODPConfig() {
	em := NewBatchEventManager(WithAPIManager(&MockEventAPIManager{}), WithBatchSize(2), WithOdpConfig(config.NewConfig("", "", nil)))
	e.False(em.ProcessEvent(Event{}))
	e.Equal(0, em.eventQueue.Size())
}

func (e *EventManagerTestSuite) TestProcessEventWithValidEventData() {
	tmpEvent := Event{
		Type:        "t1",
		Action:      "a1",
		Identifiers: map[string]string{"id-key-1": "id-value-1"},
		Data: map[string]interface{}{
			"key11": "value-1",
			"key12": true,
			"key13": 3.5,
			"key14": nil,
			"key15": 1,
		},
	}

	e.True(e.eventManager.ProcessEvent(tmpEvent))
	e.Equal(1, e.eventManager.eventQueue.Size())
}

func (e *EventManagerTestSuite) TestProcessEventWithInvalidEventData() {
	tmpEvent := Event{
		Type:        "t1",
		Action:      "a1",
		Identifiers: map[string]string{"id-key-1": "id-value-1"},
		Data: map[string]interface{}{
			"key11": map[string]interface{}{},
			"key12": []string{},
		},
	}
	e.False(e.eventManager.ProcessEvent(tmpEvent))
	e.Equal(0, e.eventManager.eventQueue.Size())
}

func (e *EventManagerTestSuite) TestProcessEventDiscardsEventExceedingMaxQueueSize() {
	e.eventManager.maxQueueSize = 1
	e.eventManager.eventQueue.Add(Event{})
	e.False(e.eventManager.ProcessEvent(Event{}))
	e.Equal(1, e.eventManager.eventQueue.Size())
}

func (e *EventManagerTestSuite) TestProcessEventWithBatchSizeNotReached() {
	em := NewBatchEventManager(WithAPIManager(&MockEventAPIManager{}), WithBatchSize(2), WithOdpConfig(config.NewConfig("a", "b", []string{"c"})))
	e.True(em.ProcessEvent(Event{}))
	e.Equal(1, em.eventQueue.Size())
	e.Equal(0, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventWithBatchSizeReached() {
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(1, e.eventManager.eventQueue.Size())
	e.eventAPIManager.wg.Add(1)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for event fire through go routine
	e.eventAPIManager.wg.Wait()
	e.Equal(0, e.eventManager.eventQueue.Size())
	e.Equal(1, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventsExceedingBatchSize() {
	e.eventManager.eventQueue.Add(Event{})
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(2, e.eventManager.eventQueue.Size())
	// Two batch events should be fired
	e.eventAPIManager.wg.Add(2)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for event fire through go routine
	e.eventAPIManager.wg.Wait()
	// Since all events fired successfully, queue should be empty
	e.Equal(0, e.eventManager.eventQueue.Size())
	e.Equal(2, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventFirstEventFailsWithRetries() {
	e.eventManager.eventQueue.Add(Event{})
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(2, e.eventManager.eventQueue.Size())
	// Return true for retry for all calls
	e.eventAPIManager.retryResponses = []bool{true, true, true}
	tmpError := errors.New("")
	// Return nil error for for all calls
	e.eventAPIManager.errResponses = []error{tmpError, tmpError, tmpError}
	// Total 3 events in queue which make 2 batches
	// first batch will be retried thrice, second one wont be fired since first failed thrice
	e.eventAPIManager.wg.Add(maxRetries)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for three retries
	e.eventAPIManager.wg.Wait()
	// Since all events failed, queue should contain all events
	e.Equal(3, e.eventManager.eventQueue.Size())
	e.Equal(3, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventFirstEventFailsWithRetryNotAllowed() {
	e.eventManager.eventQueue.Add(Event{})
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(2, e.eventManager.eventQueue.Size())
	e.eventAPIManager.retryResponses = []bool{false}
	tmpError := errors.New("")
	e.eventAPIManager.errResponses = []error{tmpError}
	// Total 3 events in queue which make 2 batches
	// first batch will not be retried, second one wont be fired since first failed
	e.eventAPIManager.wg.Add(1)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for three retries
	e.eventAPIManager.wg.Wait()
	// Since first batch of 2 events failed with no retry allowed, queue should only contain 1 event
	e.Equal(1, e.eventManager.eventQueue.Size())
	e.Equal(1, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventSecondEventFailsWithRetriesLaterPasses() {
	e.eventManager.eventQueue.Add(Event{})
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(2, e.eventManager.eventQueue.Size())
	// Return true for retry for all second batch calls
	e.eventAPIManager.retryResponses = []bool{false, true, true, true, false}
	tmpError := errors.New("")
	// Return error for all second batch calls
	e.eventAPIManager.errResponses = []error{nil, tmpError, tmpError, tmpError, nil}
	// Total 3 events in queue which make 2 batches
	// first batch will be successfully dispatched, second will be retried thrice
	e.eventAPIManager.wg.Add(4)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for events to fire
	e.eventAPIManager.wg.Wait()
	// Since second batch of 1 event failed, queue should be contain 1 event
	e.Equal(1, e.eventManager.eventQueue.Size())
	// SendOdpEvents should be called 4 times with 1 success and 3 failures
	e.Equal(4, e.eventAPIManager.timesSendEventsCalled)

	// Wait for lock to be released
	time.Sleep(200 * time.Millisecond)

	e.eventAPIManager.wg.Add(1)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for events to fire
	e.eventAPIManager.wg.Wait()
	// Queue should be empty since remaining event was sent now
	e.Equal(0, e.eventManager.eventQueue.Size())
	// SendOdpEvents should be called 5 times with 2 success and 3 failures
	e.Equal(5, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestProcessEventFirstEventPassesWithRetries() {
	e.eventManager.eventQueue.Add(Event{})
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(2, e.eventManager.eventQueue.Size())
	// Return true for first batch call only
	e.eventAPIManager.retryResponses = []bool{true, false, false}
	tmpError := errors.New("")
	// Return error first time only
	e.eventAPIManager.errResponses = []error{tmpError, nil, nil}
	// Total 3 events in queue which make 2 batches
	// first batch will be retried once, second will be successful immediately
	e.eventAPIManager.wg.Add(3)
	e.True(e.eventManager.ProcessEvent(Event{}))
	// Wait for events to fire
	e.eventAPIManager.wg.Wait()
	// Since all events were successful, queue should be empty
	e.Equal(0, e.eventManager.eventQueue.Size())
	e.Equal(3, e.eventAPIManager.timesSendEventsCalled)
}

func (e *EventManagerTestSuite) TestEventManagerAsyncBehaviour() {
	eventAPIManager := &MockEventAPIManager{}
	eventManager := NewBatchEventManager(WithAPIManager(eventAPIManager), WithBatchSize(2), WithOdpConfig(config.NewConfig("-1", "-1", []string{"-1"})))

	iterations := 100
	eventAPIManager.shouldNotInformWaitgroup = true
	eg := newExecutionContext()
	callAllMethods := func(id string) {
		eventManager.IdentifyUser(id)
		eventManager.ProcessEvent(Event{})
	}
	for i := 0; i < iterations; i++ {
		eg.Go(eventManager.Start)
		go callAllMethods(fmt.Sprintf("%d", i))
	}
	// Wait for all go routines to complete
	time.Sleep(300 * time.Millisecond)
	eg.TerminateAndWait()

	// Total expected events sent should be equal to event in queue and events sent
	e.Equal(iterations*2, eventManager.eventQueue.Size()+len(eventAPIManager.eventsSent))
	// It is possible that during concurrent dispatching, timesSendEventsCalled can exceed our expected value
	// This is because there might be odd number of events in queue when flush is called in which case
	// Flush will send the last incomplete batch too
	e.True(eventAPIManager.timesSendEventsCalled >= iterations)
}

func (e *EventManagerTestSuite) TestFlushEventsAsyncBehaviour() {
	eventAPIManager := &MockEventAPIManager{}
	batchSize := 2
	eventManager := NewBatchEventManager(WithAPIManager(eventAPIManager), WithBatchSize(batchSize), WithOdpConfig(config.NewConfig("-1", "-1", []string{"-1"})))
	iterations := 100
	eventAPIManager.wg.Add(50)
	// Add 100 events to queue
	for i := 0; i < iterations; i++ {
		eventManager.eventQueue.Add(Event{Type: fmt.Sprintf("%d", i)})
	}

	callAllMethods := func() {
		eventManager.FlushEvents()
	}
	// Call flushEvents on different go routines
	for i := 0; i < iterations; i++ {
		go callAllMethods()
	}
	// Wait for all go routines to complete
	eventAPIManager.wg.Wait()

	e.Equal(0, eventManager.eventQueue.Size())
	e.Equal(iterations/batchSize, eventAPIManager.timesSendEventsCalled)
	e.Equal(iterations, len(eventAPIManager.eventsSent))
}

func (e *EventManagerTestSuite) TestAddCommonData() {
	userEvent := Event{}
	e.eventManager.addCommonData(&userEvent)
	e.NotNil(userEvent.Data)
	e.NotEmpty(userEvent.Data["idempotence_id"])
	e.Equal("sdk", userEvent.Data["data_source_type"])
	e.Equal(event.ClientName, userEvent.Data["data_source"])
	e.Equal(event.Version, userEvent.Data["data_source_version"])
}

func (e *EventManagerTestSuite) TestUserDataOverridesCommonData() {
	userEvent := Event{Data: map[string]interface{}{
		"abc":                 nil,
		"idempotence_id":      234,
		"data_source_type":    "456",
		"data_source":         true,
		"data_source_version": 6.78,
	}}
	e.eventManager.addCommonData(&userEvent)
	e.NotNil(userEvent.Data)
	e.Equal(nil, userEvent.Data["abc"])
	e.Equal(234, userEvent.Data["idempotence_id"])
	e.Equal("456", userEvent.Data["data_source_type"])
	e.Equal(true, userEvent.Data["data_source"])
	e.Equal(6.78, userEvent.Data["data_source_version"])
}

func (e *EventManagerTestSuite) TestIsOdpServiceIntegrated() {
	e.True(e.eventManager.IsOdpServiceIntegrated())
	e.eventManager.eventQueue.Add(Event{})
	e.Equal(1, e.eventManager.eventQueue.Size())

	e.eventManager.OdpConfig = config.NewConfig("", "", nil)
	e.False(e.eventManager.IsOdpServiceIntegrated())
	e.Equal(0, e.eventManager.eventQueue.Size())
}

func (e *EventManagerTestSuite) TestEventQueueRaceCondition() {
	testIterations := 10
	var wg sync.WaitGroup
	asyncfunc := func() {
		e.eventManager.eventQueue.Add(Event{})
		e.eventManager.eventQueue.Size()
		e.eventManager.eventQueue.Get(1)
		e.eventManager.eventQueue.Remove(1)
		wg.Done()
	}
	wg.Add(testIterations)
	for i := 0; i < testIterations; i++ {
		go asyncfunc()
	}
	wg.Wait()
}

func TestEventManagerTestSuite(t *testing.T) {
	suite.Run(t, new(EventManagerTestSuite))
}

type MockEventAPIManager struct {
	wg                       sync.WaitGroup
	retryResponses           []bool  // retry responses to return
	errResponses             []error // errors to return
	timesSendEventsCalled    int     // To assert the number of times SendOdpEvents was called
	eventsSent               []Event // To assert number of events successfully sent
	shouldNotInformWaitgroup bool    // Should not call done to inform waitgroup
}

func (m *MockEventAPIManager) SendOdpEvents(apiKey, apiHost string, events []Event) (canRetry bool, err error) {
	if len(m.retryResponses) > m.timesSendEventsCalled {
		canRetry = m.retryResponses[m.timesSendEventsCalled]
	}
	if len(m.errResponses) > m.timesSendEventsCalled {
		err = m.errResponses[m.timesSendEventsCalled]
	}
	m.timesSendEventsCalled++
	// Only add to eventsSent if event was sent successfully
	if !canRetry && err == nil {
		if events == nil {
			m.eventsSent = events
		} else {
			m.eventsSent = append(m.eventsSent, events...)
		}
	}
	// Since Done can cause crashes in case when sendODPEvents is called more times then we expect it to
	// This usually occurs when multiple go routines are trying to send events and some incomplete batches are sent
	if !m.shouldNotInformWaitgroup {
		m.wg.Done()
	}
	return
}

func newExecutionContext() *pkgUtils.ExecGroup {
	return pkgUtils.NewExecGroup(context.Background(), logging.GetLogger("", "NewExecGroup"))
}