---
title: "Is Feature Enabled"
slug: "is-feature-enabled-go"
hidden: false
createdAt: "2019-08-21T21:18:35.742Z"
updatedAt: "2019-09-06T20:20:14.663Z"
---
Determines whether a feature is enabled for a given user. The purpose of this method is to allow you to separate the process of developing and deploying features from the decision to turn on a feature. Build your feature and deploy it to your application behind this flag, then turn the feature on or off for specific users.
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
This method traverses the client's [datafile](doc:get-the-datafile) and checks the feature flag for the feature key that you specify.
1. Analyzes the user's attributes.
2. Hashes the [userID](doc:handle-user-ids).

The method then evaluates the [feature rollout](doc:create-feature-flags) for a user. The method checks whether the rollout is enabled, whether the user qualifies for the [audience targeting](doc:target-audiences), and then randomly assigns either `on` or `off` based on the appropriate traffic allocation. If the feature rollout is on for a qualified user, the method returns `true`. 
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
    "0-0": "**featureKey**\n*required*",
    "0-1": "string",
    "1-0": "**userContext**\n*required*",
    "1-1": "entities.UserContext",
    "h-2": "Description",
    "0-2": "The key of the feature to check. \n\nThe feature key is defined from the Features dashboard. For more information, see [Create feature flags](doc:create-feature-flags).",
    "1-2": "Holds information about the user, such as the userID and the user's attributes."
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
The value this method returns is determined by your feature flags. This example shows the specific value returned for Go:
[block:code]
{
  "codes": [
    {
      "code": "True if feature is enabled. Otherwise, false or null.\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Examples"
}
[/block]
This section shows a simple example of how you can use the `IsFeatureEnabled` method.
[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nisEnabled, err := optlyClient.IsFeatureEnabled(\"feature_key\", user)\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Exceptions"
}
[/block]
None
[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/go-alpha/optimizely/client/client.go#L102).