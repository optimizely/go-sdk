package optlyplugins

import (
	"github.com/optimizely/go-sdk/pkg/event"
)

// EventReceiver returns dispatched events
type EventReceiver interface {
	GetEvents() []event.Batch
}

// ProxyEventDispatcher represents a valid HTTP implementation of the Dispatcher interface
type ProxyEventDispatcher struct {
	events []event.Batch
}

// DispatchEvent dispatches event with callback
func (d *ProxyEventDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	d.events = append(d.events, event.Event)
	return true, nil
}

// GetEvents returns dispatched events
func (d *ProxyEventDispatcher) GetEvents() []event.Batch {
	if d.events == nil {
		d.events = []event.Batch{}
	}
	return d.events
}
