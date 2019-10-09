package models

import "time"

// DispatcherType - represents event-dispatcher type
type DispatcherType string

const (
	// ProxyEventDispatcher - the event-dispatcher type is proxy
	ProxyEventDispatcher DispatcherType = "ProxyEventDispatcher"
	// NoOpEventDispatcher - the event-dispatcher type is no-op
	NoOpEventDispatcher DispatcherType = "NoopEventDispatcher"
)

// EventProcessorDefaultBatchSize - The default value for event processor batch size
const EventProcessorDefaultBatchSize = 1

// EventProcessorDefaultQueueSize - The default value for event processor queue size
const EventProcessorDefaultQueueSize = 1

// EventProcessorDefaultFlushInterval - The default value for event processor flush interval
const EventProcessorDefaultFlushInterval = 250 * time.Millisecond
