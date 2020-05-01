---
title: "Get Enabled Features"
slug: "get-enabled-features-go"
hidden: false
createdAt: "2019-08-21T21:17:54.754Z"
updatedAt: "2019-09-13T11:51:11.242Z"
---
Retrieves a list of features that are enabled for the user. Invoking this method is equivalent to running [Is Feature Enabled](doc:is-feature-enabled-go) for each feature in the datafile sequentially.

This method takes into account the user `attributes` passed in, to determine if the user is part of the audience that qualifies for the experiment.  
[block:api-header]
{
  "title": "Version"
}
[/block]
SDK v0.1.0
[block:api-header]
{
  "title": "Description"
}
[/block]
This method iterates through all feature flags and for each feature flag invokes [Is Feature Enabled](doc:is-feature-enabled-go). If a feature is enabled, this method adds the feature’s key to the return list.
[block:parameters]
{
  "data": {
    "0-0": "**userContext**\n*required*",
    "0-1": "entities.UserContext",
    "0-2": "Holds information about the user, such as the userID and the user's attributes."
  },
  "cols": 3,
  "rows": 1
}
[/block]

[block:api-header]
{
  "title": "Returns"
}
[/block]
A list of keys corresponding to the features that are enabled for the user, or an empty list if no features could be found for the specified user.
[block:api-header]
{
  "title": "Examples"
}
[/block]
This section shows a simple example of how you can use the method.
[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nisEnabled, err := optlyClient.GetEnabledFeatures(user)",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/go-alpha/optimizely/client/client.go#L102).