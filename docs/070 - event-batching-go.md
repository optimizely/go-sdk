---
title: "Event batching"
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
[block:callout]
{
  "type": "info",
  "title": "Note",
  "body": "Event batching works with both out-of-the-box and custom event dispatchers.\n\nThe event batching process doesn't remove any personally identifiable information (PII) from events. Please ensure that you aren't sending any unnecessary PII to Optimizely."
}
[/block]

[block:api-header]
{
  "title": "Basic example"
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "import optly \"github.com/optimizely/go-sdk\"\n\n// the default client will have a BatchEventProcessor with the default options\noptlyClient, err := optly.Client(\"SDK_KEY_HERE\")",
      "language": "go"
    }
  ]
}
[/block]
By default, batch size is 10 and flush interval is 30 seconds.
[block:api-header]
{
  "title": "Advanced Example"
}
[/block]
To customize the event processor, you can use the client factory methods.
[block:code]
{
  "codes": [
    {
      "code": "import (\n  \"time\"\n  \n\t\"github.com/optimizely/go-sdk/pkg/client\"\n\t\"github.com/optimizely/go-sdk/pkg/event\"\n  \"github.com/optimizely/go-sdk/pkg/utils\"\n)\n\noptimizelyFactory := &client.OptimizelyFactory{\n\t\tSDKKey: \"SDK_KEY\",\t\n}\n\n// You can configure the batch size and flush interval\neventProcessor := event.NewBatchEventProcessor(\n  event.WithBatchSize(10), \n  event.WithFlushInterval(30 * time.Second),\n)\noptlyClient, err := optimizelyFactory.Client(\n  client.WithEventProcessor(eventProcessor),\n)\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "BatchEventProcessor"
}
[/block]
`BatchEventProcessor ` is an implementation of `EventProcessor` where events are batched. The class maintains a single consumer thread that pulls events off of an in-memory queue and buffers them for either a configured batch size or a maximum duration before the resulting `LogEvent` is sent to the `EventDispatcher` and `NotificationCenter`.

The following properties can be used to customize the BatchEventProcessor configuration.
[block:parameters]
{
  "data": {
    "h-0": "Property",
    "h-1": "Default value",
    "0-0": "**event.EventDispatcher**",
    "0-1": "NewQueueEventDispatcher",
    "1-1": "10",
    "1-0": "**event.BatchSize**",
    "h-2": "Description",
    "h-3": "Server",
    "0-2": "Used to dispatch event payload to Optimizely.",
    "1-2": "The maximum number of events to batch before dispatching. Once this number is reached, all queued events are flushed and sent to Optimizely.",
    "0-3": "Based on your organization's requirements.",
    "1-3": "Based on your organization's requirements.",
    "3-0": "**event.Q**",
    "3-1": "NewInMemoryQueue",
    "3-2": "BlockingCollection that queues individual events to be batched and dispatched by the executor.",
    "2-0": "**event.FlushInterval**",
    "2-1": "30000 (30 Seconds)",
    "2-2": "Milliseconds to wait before batching and dispatching events.",
    "4-0": "**event.MaxQueueSize**",
    "4-1": "2000",
    "4-2": "The maximum number of events that can be queued."
  },
  "cols": 3,
  "rows": 5
}
[/block]
For more information, see [Initialize SDK](doc:initialize-sdk-go).
[block:api-header]
{
  "title": "Side Effects"
}
[/block]
The table lists other Optimizely functionality that may be triggered by using this method.
[block:parameters]
{
  "data": {
    "h-0": "Functionality",
    "h-1": "Description",
    "0-0": "LogEvent",
    "0-1": "Whenever the event processor produces a batch of events, a LogEvent object will be created using the event factory.\nIt contains batch of conversion and impression events. \nThis object will be dispatched using the provided event dispatcher and also it will be sent to the notification subscribers",
    "1-0": "Notification Listeners",
    "1-1": "Flush invokes the LOGEVENT [notification listener](doc:set-up-notification-listener-go) if this listener is subscribed to."
  },
  "cols": 2,
  "rows": 2
}
[/block]
### Registering and Unregistering LogEvent listener

The example code below shows how to add and remove a LogEvent notification listener.
[block:code]
{
  "codes": [
    {
      "code": "import (\n\t\"fmt\"\n\n\t\"github.com/optimizely/go-sdk/pkg/client\"\n\t\"github.com/optimizely/go-sdk/pkg/event\"\n)\n\n// Callback for log event notification\n\tcallback := func(notification event.LogEvent) {\n\n\t\t// URL to dispatch log event to\n\t\tfmt.Print(notification.EndPoint)\n\t\t// Batched event\n\t\tfmt.Print(notification.Event)\n\t}\n\n\toptimizelyFactory := &client.OptimizelyFactory{\n\t\tSDKKey: \"SDK_KEY\",\n}\n\toptimizelyClient, err := optimizelyFactory.Client()\n\n\t// Add callback for logEvent notification\n\tid, err := optimizelyClient.EventProcessor.(*event.BatchEventProcessor).OnEventDispatch(callback)\n\n\t// Remove callback for logEvent notification\n\terr = optimizelyClient.EventProcessor.(*event.BatchEventProcessor).RemoveOnEventDispatch(id)",
      "language": "go"
    }
  ]
}
[/block]
###  LogEvent

LogEvent object gets created using [factory](https://github.com/optimizely/go-sdk/blob/8a8fb7e959f2597d26d2a0dc3a6a072dcbc15f0f/pkg/event/factory.go#L46). It represents the batch of impression and conversion events we send to the Optimizely backend.
[block:parameters]
{
  "data": {
    "h-0": "Object",
    "h-1": "Type",
    "h-2": "Description",
    "0-0": "**EndPoint**\n*Required (non null)*",
    "0-1": "String",
    "0-2": "URL to dispatch log event to.",
    "1-0": "**[Event](https://github.com/optimizely/go-sdk/blob/8a8fb7e959f2597d26d2a0dc3a6a072dcbc15f0f/pkg/event/events.go#L70)**\n*Required*",
    "1-1": "[event.Batch]",
    "1-2": "It contains all the information regarding every event which is batched. including list of visitors which contains UserEvent."
  },
  "cols": 3,
  "rows": 2
}
[/block]

[block:api-header]
{
  "title": "Close Optimizely on application exit"
}
[/block]
If you enable event batching, it's important that you call the Close method (`optimizelyClient.Close()`) prior to exiting. This ensures that queued events are flushed as soon as possible to avoid any data loss.
[block:callout]
{
  "type": "warning",
  "body": "Because the Optimizely client maintains a buffer of queued events, we recommend that you call `Close()` on the Optimizely instance before shutting down your application or whenever dereferencing the instance."
}
[/block]

[block:parameters]
{
  "data": {
    "h-0": "Method",
    "h-1": "Description",
    "0-0": "**Close()**",
    "0-1": "Stops all executor threads and flushes the event queue. This method will also stop any scheduledExecutorService that is running for the data-file manager."
  },
  "cols": 2,
  "rows": 1
}
[/block]