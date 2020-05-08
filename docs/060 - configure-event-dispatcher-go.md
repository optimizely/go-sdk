---
title: "Configure event dispatcher"
excerpt: ""
slug: "configure-event-dispatcher-go"
hidden: true
createdAt: "2019-09-12T13:58:15.596Z"
updatedAt: "2019-12-10T00:19:39.940Z"
---
The Optimizely SDKs make HTTP requests for every impression or conversion that gets triggered. Each SDK has a built-in **event dispatcher** for handling these events, but we recommend overriding it based on the specifics of your environment.

The Go SDK has an out-of-the-box asynchronous dispatcher. We recommend customizing the event dispatcher you use in production to ensure that you queue and send events in a manner that scales to the volumes handled by your application. Customizing the event dispatcher allows you to take advantage of features like batching, which makes it easier to handle large event volumes efficiently or to implement retry logic when a request fails. You can build your dispatcher from scratch or start with the provided dispatcher.

The examples show that to customize the event dispatcher, initialize the Optimizely client (or manager) with an event dispatcher instance.
```go
import "github.com/optimizely/go-sdk/pkg/event"

type CustomEventDispatcher struct {
}

// DispatchEvent dispatches event with callback
func (d *CustomEventDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	dispatchedEvent := map[string]interface{}{
		"url":       event.EndPoint,
		"http_verb": "POST",
		"headers":   map[string]string{"Content-Type": "application/json"},
		"params":    event.Event,
	}
	return true, nil
}

```

```go
import (
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/event"
)

optimizelyFactory := &client.OptimizelyFactory{
  SDKKey: "SDK_KEY_HERE",
	}

customEventDispatcher := &CustomEventDispatcher{}

// Create an Optimizely client with the custom event dispatcher
optlyClient, e := optimizelyFactory.Client(client.WithEventDispatcher(customEventDispatcher))


```

The event dispatcher should implement a `DispatchEvent` function, which takes in one argument: `event.LogEvent`. In this function, you should send a `POST` request to the given `event.EndPoint` using the `event.EndPoint` as the body of the request (be sure to stringify it to JSON) and `{content-type: 'application/json'}` in the headers.

>⚠️ Important
>
> If you are using a custom event dispatcher, do not modify the event payload returned from Optimizely. Modifying this payload will alter your results.

By default, our Go SDK uses a [basic asynchronous event dispatcher](https://github.com/optimizely/go-sdk/blob/master/pkg/event/dispatcher.go).