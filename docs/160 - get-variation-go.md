---
title: "Get Variation"
slug: "get-variation-go"
hidden: true
createdAt: "2019-09-12T14:11:49.333Z"
updatedAt: "2019-10-29T23:44:33.591Z"
---
Returns a variation where the visitor will be bucketed, without triggering an impression.
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
Takes the same arguments and returns the same values as [Activate](doc:activate), but without sending an impression event. The behavior of the two methods is identical otherwise. 

Use **GetVariation** if **Activate** has been called and the current variation assignment is needed for a given experiment and user. This method bypasses redundant network requests to Optimizely.

See [Implement impressions](doc:implement-impressions) for guidance on when to use each method.
[block:api-header]
{
  "title": "Parameters"
}
[/block]
The table below lists the required and optional parameters in Go.
[block:parameters]
{
  "data": {
    "h-0": "Parameter",
    "h-1": "Type",
    "h-2": "Description",
    "0-0": "**experimentKey**\n*required*",
    "0-1": "string",
    "1-0": "**userContext**\n*required*",
    "1-1": "entities.UserContext",
    "0-2": "The key of the experiment.",
    "1-2": "Holds information about the user, such as the userID and the user's attributes.",
    "2-2": "A map of custom key-value string pairs specifying attributes for the user that are used for [audience targeting](doc:define-audiences-and-attributes) and [results segmentation](doc:analyze-results#section-segment-results). Non-string values are only supported in the 3.0 SDK and above.",
    "2-1": "map",
    "2-0": "**attributes**\n*optional*"
  },
  "cols": 3,
  "rows": 2
}
[/block]

[block:api-header]
{
  "title": "Returns"
}
[/block]
A variation key, or an empty string if no variation is found.
[block:api-header]
{
  "title": "Example"
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nvariationKey, err := optlyClient.GetVariation(\"experiment_key\", user)\n\n",
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
[Activate](doc:activate) 
[Implement impressions](doc:implement-impressions)
[block:api-header]
{
  "title": "Notes"
}
[/block]
### Activate versus Get Variation
Use Activate when the visitor actually sees the experiment. Use Get Variation when you need to know which bucket a visitor is in before showing the visitor the experiment. Impressions are tracked by [Is Feature Enabled](doc:is-feature-enabled-go) when there is a feature test running on the feature and the visitor qualifies for that feature test.

For example, suppose you want your web server to show a visitor variation_1 but don't want the visitor to count until they open a feature that isn't visible when the variation loads, like a modal. In this case, use Get Variation in the backend to specify that your web server should respond with variation_1, and use Activate in the front end when the visitor sees the experiment.

Also, use Get Variation when you're trying to align your Optimizely results with client-side third-party analytics. In this case, use Get Variation to retrieve the variation&mdash;and even show it to the visitor&mdash;but only call Activate when the analytics call goes out.

See [Implement impressions](doc:implement-impressions) for more information about whether to use Activate or Get Variation for a call.
[block:callout]
{
  "type": "warning",
  "title": "Important",
  "body": "Conversion events can only be attributed to experiments with previously tracked impressions. Impressions are tracked by Activate, not by Get Variation. As a general rule, Optimizely impressions are required for experiment results and not only for billing."
}
[/block]

[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L265).