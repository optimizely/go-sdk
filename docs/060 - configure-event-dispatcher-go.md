---
title: "Configure event dispatcher"
slug: "configure-event-dispatcher-go"
hidden: true
createdAt: "2019-09-12T13:58:15.596Z"
updatedAt: "2019-12-10T00:19:39.940Z"
---
The Optimizely SDKs make HTTP requests for every impression or conversion that gets triggered. Each SDK has a built-in **event dispatcher** for handling these events, but we recommend overriding it based on the specifics of your environment.

The Go SDK has an out-of-the-box asynchronous dispatcher. We recommend customizing the event dispatcher you use in production to ensure that you queue and send events in a manner that scales to the volumes handled by your application. Customizing the event dispatcher allows you to take advantage of features like batching, which makes it easier to handle large event volumes efficiently or to implement retry logic when a request fails. You can build your dispatcher from scratch or start with the provided dispatcher.

The examples show that to customize the event dispatcher, initialize the Optimizely client (or manager) with an event dispatcher instance.
[block:code]
{
  "codes": [
    {
      "code": "import \"github.com/optimizely/go-sdk/pkg/event\"\n\ntype CustomEventDispatcher struct {\n}\n\n// DispatchEvent dispatches event with callback\nfunc (d *CustomEventDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {\n\tdispatchedEvent := map[string]interface{}{\n\t\t\"url\":       event.EndPoint,\n\t\t\"http_verb\": \"POST\",\n\t\t\"headers\":   map[string]string{\"Content-Type\": \"application/json\"},\n\t\t\"params\":    event.Event,\n\t}\n\treturn true, nil\n}\n",
      "language": "go"
    }
  ]
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "import (\n\t\"github.com/optimizely/go-sdk/pkg/client\"\n\t\"github.com/optimizely/go-sdk/pkg/event\"\n)\n\noptimizelyFactory := &client.OptimizelyFactory{\n  SDKKey: \"SDK_KEY_HERE\",\n\t}\n\ncustomEventDispatcher := &CustomEventDispatcher{}\n\n// Create an Optimizely client with the custom event dispatcher\noptlyClient, e := optimizelyFactory.Client(client.WithEventDispatcher(customEventDispatcher))\n\n",
      "language": "go"
    }
  ]
}
[/block]
The event dispatcher should implement a `DispatchEvent` function, which takes in one argument: `event.LogEvent`. In this function, you should send a `POST` request to the given `event.EndPoint` using the `event.EndPoint` as the body of the request (be sure to stringify it to JSON) and `{content-type: 'application/json'}` in the headers.
[block:callout]
{
  "type": "warning",
  "title": "Important",
  "body": "If you are using a custom event dispatcher, do not modify the event payload returned from Optimizely. Modifying this payload will alter your results."
}
[/block]
By default, our Go SDK uses a [basic asynchronous event dispatcher](https://github.com/optimizely/go-sdk/blob/master/pkg/event/dispatcher.go).