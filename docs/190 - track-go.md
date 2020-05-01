---
title: "Track"
slug: "track-go"
hidden: true
createdAt: "2019-09-12T14:12:06.393Z"
updatedAt: "2020-02-21T19:34:42.881Z"
---
Tracks a conversion event. Logs an error message if the specified event key doesn't match any existing events.

[block:api-header]
{
  "title": "Version"
}
[/block]
SDK v1.0
[block:api-header]
{
  "title": "Description"
}
[/block]
Use this method to [track events](doc:event-tracking) across multiple experiments. You should only send one tracking event per conversion, even if many feature tests or A/B tests are measuring it.
[block:callout]
{
  "type": "info",
  "title": "Note",
  "body": "Events are counted in an experiment when an impression was sent as a result of the [Activate](doc:activate) or [Is Feature Enabled](doc:is-feature-enabled) method being called."
}
[/block]
The attributes passed to Track are only used for [results segmentation](doc:analyze-results#section-segment-results).

The [Track](doc:track) function can be used to track events across multiple experiments. It will be counted for each experiment only if [Is Feature Enabled](doc:is-feature-enabled) or [Activate](doc:activate) has previously been called for the current user.

As long as the event key is valid (the key appears in the datafile) Optimizely will track events. Use attributes to track events for segmentation purposes.
[block:callout]
{
  "type": "warning",
  "body": "**The [Results page](doc:analyze-results) only shows events that are tracked after [Is Feature Enabled](doc:is-feature-enabled) or [Activate](doc:activate) has been called**. If you do not see results on the Results page, make sure that you are evaluating the feature flag before tracking conversion events.",
  "title": "Important"
}
[/block]

[block:api-header]
{
  "title": "Track events across platforms"
}
[/block]
For offline event tracking and other advanced use cases, you can also use the [Events API](https://developers.optimizely.com/x/events/api/).

You can use any of our SDKs to track events, so you can run experiments that span multiple applications, services, or devices. All of our SDKs have the same audience evaluation and targeting behavior, so you'll see the same output from feature experiment evaluation and tracking as long as you are using the same datafile and user IDs.

For example, if you're running feature experiments on your server, you can evaluate them with our Go, Python, Java, Ruby, C#, Node, or PHP SDKs, but track user actions client-side using our JavaScript, Objective-C or Android SDKs.

If you plan to use multiple SDKs for the same project, make sure that all SDKs share the same datafile and user IDs.

For more information, see [Multiple languages](doc:multiple-languages).
[block:api-header]
{
  "title": "Parameters"
}
[/block]
This table lists the required and optional parameters for the Go SDK.
[block:parameters]
{
  "data": {
    "h-0": "Parameter",
    "h-1": "Type",
    "h-2": "Description",
    "0-0": "**event key**\n*required* ",
    "0-1": "string",
    "1-0": "**userContext**\n*required*",
    "1-1": "entities.UserContext",
    "2-0": "**eventTags**\n*optional*",
    "3-0": "**event tags**\n*optional* ",
    "3-1": "map",
    "2-1": "map",
    "0-2": "The key of the event to be tracked. This key must match the event key provided when the event was created in the Optimizely app.",
    "1-2": "Holds information about the user, such as the userID and the user's attributes.\n\n**Important**: userID must match the user ID provided to Activate or Is Feature Enabled.",
    "2-2": "A map of key-value pairs specifying tag names and their corresponding tag values for this particular event occurrence. Values can be strings, numbers, or booleans.\n\nThese can be used to track numeric metrics, allowing you to track actions beyond conversions, for example: revenue, load time, or total value. [See details on reserved tag keys.](https://docs.developers.optimizely.com/full-stack/docs/include-event-tags#section-reserved-tag-keys)",
    "3-2": "A map of key-value pairs specifying tag names and their corresponding tag values for this particular event occurrence. Values can be strings, numbers, or booleans."
  },
  "cols": 3,
  "rows": 3
}
[/block]

[block:api-header]
{
  "title": "Returns"
}
[/block]
This method sends conversion data to Optimizely. It doesn't provide return values. 
[block:api-header]
{
  "title": "Example"
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nerr := optlyClient.Track(\"add_to_cart\", user)\n  \n  ",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "See also"
}
[/block]
For more guidance and examples, see [Track events](doc:event-tracking).
[block:api-header]
{
  "title": "Side effects"
}
[/block]
The table lists other other Optimizely functionality that may be triggered by using this method.
[block:parameters]
{
  "data": {
    "h-0": "Functionality",
    "h-1": "Description",
    "0-0": "Conversions",
    "0-1": "Calling this method records a conversion and attributes it to the variations that the user has seen.\n \nFull Stack 3.x supports retroactive metrics calculation. You can create [metrics](doc:identify-metrics) on this conversion event and add metrics to experiments even after the conversion has been tracked.\n\nFor more information, see the paragraph **Events are always on** in the introduction of [Events: Tracking clicks, pageviews, and other visitor actions](https://help.optimizely.com/Measure_success%3A_Track_visitor_behaviors/Events%3A_Tracking_clicks%2C_pageviews%2C_and_other_visitor_actions).\n\n**Important!** \n - This method won't track events when the specified event key is invalid.\n - Changing the traffic allocation of running experiments affects how conversions are recorded and variations are attributed to users.",
    "1-0": "Impressions",
    "1-1": "Track doesn't trigger impressions."
  },
  "cols": 2,
  "rows": 2
}
[/block]

[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L296).