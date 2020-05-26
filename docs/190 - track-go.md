---
title: "Track"
excerpt: ""
slug: "track-go"
hidden: true
createdAt: "2019-09-12T14:12:06.393Z"
updatedAt: "2020-02-21T19:34:42.881Z"
---
Tracks a conversion event. Logs an error message if the specified event key doesn't match any existing events.

### Version
SDK v1.0

### Description
Use this method to [track events](doc:event-tracking) across multiple experiments. You should only send one tracking event per conversion, even if many feature tests or A/B tests are measuring it.

>ℹ️ Note
>
> Events are counted in an experiment when an impression was sent as a result of the [Activate](doc:activate) or [Is Feature Enabled](doc:is-feature-enabled) method being called.
The attributes passed to Track are only used for [results segmentation](doc:analyze-results#section-segment-results).

The [Track](doc:track) function can be used to track events across multiple experiments. It will be counted for each experiment only if [Is Feature Enabled](doc:is-feature-enabled) or [Activate](doc:activate) has previously been called for the current user.

As long as the event key is valid (the key appears in the datafile) Optimizely will track events. Use attributes to track events for segmentation purposes.

⚠️ Important

**The [Results page](doc:analyze-results) only shows events that are tracked after [Is Feature Enabled](doc:is-feature-enabled) or [Activate](doc:activate) has been called**. If you do not see results on the Results page, make sure that you are evaluating the feature flag before tracking conversion events.

### Track events across platforms
For offline event tracking and other advanced use cases, you can also use the [Events API](https://developers.optimizely.com/x/events/api/).

You can use any of our SDKs to track events, so you can run experiments that span multiple applications, services, or devices. All of our SDKs have the same audience evaluation and targeting behavior, so you'll see the same output from feature experiment evaluation and tracking as long as you are using the same datafile and user IDs.

For example, if you're running feature experiments on your server, you can evaluate them with our Go, Python, Java, Ruby, C#, Node, or PHP SDKs, but track user actions client-side using our JavaScript, Objective-C or Android SDKs.

If you plan to use multiple SDKs for the same project, make sure that all SDKs share the same datafile and user IDs.

For more information, see [Multiple languages](doc:multiple-languages).
### Parameters
This table lists the required and optional parameters for the Go SDK.

| Parameter                      | Type                 | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
|--------------------------------|----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **event key**<br/>*required*   | string               | The key of the event to be tracked. This key must match the event key provided when the event was created in the Optimizely app.                                                                                                                                                                                                                                                                                                                                |
| **userContext**<br/>*required* | entities.UserContext | Holds information about the user, such as the userID and the user's attributes.<br/>**Important**: userID must match the user ID provided to Activate or Is Feature Enabled.                                                                                                                                                                                                                                                                                    |
| **eventTags**<br/>*optional*   | map                  | A map of key-value pairs specifying tag names and their corresponding tag values for this particular event occurrence. Values can be strings, numbers, or booleans.<br/>These can be used to track numeric metrics, allowing you to track actions beyond conversions, for example: revenue, load time, or total value. [See details on reserved tag keys.](https://docs.developers.optimizely.com/full-stack/docs/include-event-tags#section-reserved-tag-keys) |

### Returns
This method sends conversion data to Optimizely. It doesn't provide return values. 

### Example
```go
attributes := map[string]interface{}{
        "DEVICE": "iPhone",
        "hey":    2,
}

user := entities.UserContext{
        ID:         "userId",
        Attributes: attributes,
}

err := optlyClient.Track("add_to_cart", user)
```

### See also
For more guidance and examples, see [Track events](doc:event-tracking).

### Side effects
The table lists other other Optimizely functionality that may be triggered by using this method.

| Functionality | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|---------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Conversions   | Calling this method records a conversion and attributes it to the variations that the user has seen.<br/>Full Stack 3.x supports retroactive metrics calculation. You can create [metrics](doc:identify-metrics) on this conversion event and add metrics to experiments even after the conversion has been tracked.<br/>For more information, see the paragraph **Events are always on** in the introduction of [Events: Tracking clicks, pageviews, and other visitor actions](https://help.optimizely.com/Measure_success%3A_Track_visitor_behaviors/Events%3A_Tracking_clicks%2C_pageviews%2C_and_other_visitor_actions).<br/>**Important!** <br/> - This method won't track events when the specified event key is invalid.<br/> - Changing the traffic allocation of running experiments affects how conversions are recorded and variations are attributed to users. |
| Impressions   | Track doesn't trigger impressions.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L296).