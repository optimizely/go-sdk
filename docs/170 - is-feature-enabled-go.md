---
title: "Is Feature Enabled"
excerpt: ""
slug: "is-feature-enabled-go"
hidden: false
createdAt: "2019-08-21T21:18:35.742Z"
updatedAt: "2019-09-06T20:20:14.663Z"
---
Determines whether a feature is enabled for a given user. The purpose of this method is to allow you to separate the process of developing and deploying features from the decision to turn on a feature. Build your feature and deploy it to your application behind this flag, then turn the feature on or off for specific users.

### Version
SDK v1.0

### Description
This method traverses the client's [datafile](doc:get-the-datafile) and checks the feature flag for the feature key that you specify.
1. Analyzes the user's attributes.
2. Hashes the [userID](doc:handle-user-ids).

The method then evaluates the [feature rollout](doc:create-feature-flags) for a user. The method checks whether the rollout is enabled, whether the user qualifies for the [audience targeting](doc:target-audiences), and then randomly assigns either `on` or `off` based on the appropriate traffic allocation. If the feature rollout is on for a qualified user, the method returns `true`.

### Parameters
The table below lists the required and optional parameters in Go.

| Parameter                       | Type                 | Description                                                                                                                                                               |
|---------------------------------|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **featureKey** <br/> *required* | string               | The key of the feature to check. <br/>The feature key is defined from the Features dashboard. For more information, see [Create feature flags](doc:create-feature-flags). |
| **userContext**<br/>*required*  | entities.UserContext | Holds information about the user, such as the userID and the user's attributes.                                                                                           |

### Returns
The value this method returns is determined by your feature flags. This example shows the specific value returned for Go:

```go
True if feature is enabled. Otherwise, false or null.
```

### Examples
This section shows a simple example of how you can use the `IsFeatureEnabled` method.

```go
attributes := map[string]interface{}{
        "DEVICE": "iPhone",
        "hey":    2,
}

user := entities.UserContext{
        ID:         "userId",
        Attributes: attributes,
}

isEnabled, err := optlyClient.IsFeatureEnabled("feature_key", user)

```

### Exceptions
None

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/go-alpha/optimizely/client/client.go#L102).