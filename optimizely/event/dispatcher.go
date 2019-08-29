package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/optimizely/go-sdk/optimizely/logging"
)

const jsonContentType = "application/json"
var dispatcherLogger = logging.GetLogger("EventDispatcher")

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent, callback func(success bool))
}

// HTTPEventDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPEventDispatcher struct {
}

// DispatchEvent dispatches event with callback
func (*HTTPEventDispatcher) DispatchEvent(event LogEvent, callback func(success bool)) {
	dispatchEvent(event, callback)
}

type QueueEventDispatcher struct {
	q Queue
	mux sync.Mutex
}

func (ed *QueueEventDispatcher) DispatchEvent(event LogEvent, callback func(success bool)) {

	ed.q.Add(event)
	callback(true)
	go func() {
		ed.flushEvents()
	}()
}

func dispatchEvent(event LogEvent, callback func(success bool)) {
	jsonValue, _ := json.Marshal(event)
	resp, err := http.Post(event.endPoint, jsonContentType, bytes.NewBuffer(jsonValue))
	// also check response codes
	// resp.StatusCode == 400 is an error
	success := true

	if err != nil {
		dispatcherLogger.Error("http.Post failed:", err)
		success = false
	} else {
		if resp.StatusCode == 204 {
			success = true
		} else {
			dispatcherLogger.Error(fmt.Sprintf("http.Post invalid response %d", resp.StatusCode), nil)
			success = false
		}
	}
	callback(success)
}

func (ed *QueueEventDispatcher) flushEvents() {

	ed.mux.Lock()

	defer func(){
		ed.mux.Unlock()
	}()

	for ed.q.Size() > 0 {
		items := ed.q.Get(1)
		if len(items) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}
		event, ok := items[0].(LogEvent)
		if !ok {
			//remove it?
			dispatcherLogger.Error("invalid type passed to event dispatcher", nil)
			ed.q.Remove(1)
		}

		dispatchEvent(event, func(success bool) {
			if success {
				ed.q.Remove(1)
			} else {
				// we failed.  Sleep 5 seconds and try again.  Or, should we exit and try the next time
				// a log event is added?
				time.Sleep(5 * time.Second)
			}
		})
	}
}

func NewQueueEventDispatcher() Dispatcher {
	dispatcher := QueueEventDispatcher{q:NewInMemoryQueue(1000)}
	return &dispatcher
}