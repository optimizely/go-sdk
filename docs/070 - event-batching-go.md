---
title: "Event batching"
excerpt: ""
slug: "event-batching-go"
hidden: true
createdAt: "2019-10-29T23:36:28.978Z"
updatedAt: "2020-01-16T20:42:04.941Z"
---
The [Optimizely Full Stack Go SDK](https://github.com/optimizely/go-sdk) batches impression and conversion events into a single payload before sending it to Optimizely. This is achieved through an SDK component called the event processor.

Event batching has the advantage of reducing the number of outbound requests to Optimizely depending on how you define, configure, and use the event processor. It means less network traffic for the same number of Impression and conversion events tracked.

In the Go SDK, `QueueingEventProcessor` provides implementation of the `EventProcessor` interface and batches events. You can control batching based on two parameters:

- Batch size: Defines the number of events that are batched together before sending to Optimizely.
- Flush interval: Defines the amount of time after which any batched events should be sent to Optimizely.

An event consisting of the batched payload is sent as soon as the batch size reaches the specified limit or flush interval reaches the specified time limit. `Batchcessor` options are described in more detail below.

>⚠️ Note
>
> Event batching works with both out-of-the-box and custom event dispatchers.\n\nThe event batching process doesn't remove any personally identifiable information (PII) from events. Please ensure that you aren't sending any unnecessary PII to Optimizely.

### Basic example

```go
import optly "github.com/optimizely/go-sdk"

// the default client will have a BatchEventProcessor with the default options
optlyClient, err := optly.Client("SDK_KEY_HERE")
```
By default, batch size is 10 and flush interval is 30 seconds.

### Advanced Example
To customize the event processor, you can use the client factory methods.

```go
import (
  "time"
  
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/event"
  "github.com/optimizely/go-sdk/pkg/utils"
)

optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "SDK_KEY",	
}

// You can configure the batch size and flush interval
eventProcessor := event.NewBatchEventProcessor(
  event.WithBatchSize(10), 
  event.WithFlushInterval(30 * time.Second),
)
optlyClient, err := optimizelyFactory.Client(
  client.WithEventProcessor(eventProcessor),
)

```

### BatchEventProcessor
`BatchEventProcessor ` is an implementation of `EventProcessor` where events are batched. The class maintains a single consumer thread that pulls events off of an in-memory queue and buffers them for either a configured batch size or a maximum duration before the resulting `LogEvent` is sent to the `EventDispatcher` and `NotificationCenter`.

The following properties can be used to customize the BatchEventProcessor configuration.

| Property                  | Default value           | Description                                                                                                                                  |
|---------------------------|-------------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| **event.EventDispatcher** | NewQueueEventDispatcher | Used to dispatch event payload to Optimizely.                                                                                                |
| **event.BatchSize**       | 10                      | The maximum number of events to batch before dispatching. Once this number is reached, all queued events are flushed and sent to Optimizely. |
| **event.FlushInterval**   | 30000 (30 Seconds)      | Milliseconds to wait before batching and dispatching events.                                                                                 |
| **event.Q**               | NewInMemoryQueue        | BlockingCollection that queues individual events to be batched and dispatched by the executor.                                               |
| **event.MaxQueueSize**    | 2000                    | The maximum number of events that can be queued.                                                                                             |

For more information, see [Initialize SDK](doc:initialize-sdk-go).
### Side Effects
The table lists other Optimizely functionality that may be triggered by using this method.

| Functionality          | Description                                                                                                                                                                                                                                                                                               |
|------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| LogEvent               | Whenever the event processor produces a batch of events, a LogEvent object will be created using the event factory. It contains batch of conversion and impression events. This object will be dispatched using the provided event dispatcher and also it will be sent to the notification subscribers |
| Notification Listeners | Flush invokes the LOGEVENT [notification listener](doc:set-up-notification-listener-go) if this listener is subscribed to.                                                                                                                                                                                |
### Registering and Unregistering LogEvent listener

The example code below shows how to add and remove a LogEvent notification listener.
```go
import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/event"
)

// Callback for log event notification
	callback := func(notification event.LogEvent) {

		// URL to dispatch log event to
		fmt.Print(notification.EndPoint)
		// Batched event
		fmt.Print(notification.Event)
	}

	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "SDK_KEY",
}
	optimizelyClient, err := optimizelyFactory.Client()

	// Add callback for logEvent notification
	id, err := optimizelyClient.EventProcessor.(*event.BatchEventProcessor).OnEventDispatch(callback)

	// Remove callback for logEvent notification
	err = optimizelyClient.EventProcessor.(*event.BatchEventProcessor).RemoveOnEventDispatch(id)
```
###  LogEvent

LogEvent object gets created using [factory](https://github.com/optimizely/go-sdk/blob/8a8fb7e959f2597d26d2a0dc3a6a072dcbc15f0f/pkg/event/factory.go#L46). It represents the batch of impression and conversion events we send to the Optimizely backend.

| Object                                                                                                                                  | Type          | Description                                                                                                                  |
|-----------------------------------------------------------------------------------------------------------------------------------------|---------------|------------------------------------------------------------------------------------------------------------------------------|
| **EndPoint** <br/>*Required (non null)*                                                                                                 | String        | URL to dispatch log event to.                                                                                                |
| **[Event](https://github.com/optimizely/go-sdk/blob/8a8fb7e959f2597d26d2a0dc3a6a072dcbc15f0f/pkg/event/events.go#L70)** <br/>*Required* | [event.Batch] | It contains all the information regarding every event which is batched. including list of visitors which contains UserEvent. |

### Close Optimizely on application exit

If you enable event batching, it's important that you call the Close method (`optimizelyClient.Close()`) prior to exiting. This ensures that queued events are flushed as soon as possible to avoid any data loss.

>⚠️ Important
>
> Because the Optimizely client maintains a buffer of queued events, we recommend that you call `Close()` on the Optimizely instance before shutting down your application or whenever dereferencing the instance.

| Method      | Description                                                                                                                                                |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Close()** | Stops all executor threads and flushes the event queue. This method will also stop any scheduledExecutorService that is running for the data-file manager. |