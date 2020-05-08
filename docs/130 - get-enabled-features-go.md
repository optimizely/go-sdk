---
title: "Get Enabled Features"
excerpt: ""
slug: "get-enabled-features-go"
hidden: false
createdAt: "2019-08-21T21:17:54.754Z"
updatedAt: "2019-09-13T11:51:11.242Z"
---
Retrieves a list of features that are enabled for the user. Invoking this method is equivalent to running [Is Feature Enabled](doc:is-feature-enabled-go) for each feature in the datafile sequentially.

This method takes into account the user `attributes` passed in, to determine if the user is part of the audience that qualifies for the experiment.  

### Version
SDK v0.1.0

### Description
This method iterates through all feature flags and for each feature flag invokes [Is Feature Enabled](doc:is-feature-enabled-go). If a feature is enabled, this method adds the feature’s key to the return list.

| **userContext** <br/>*required* | entities.UserContext | Holds information about the user, such as the userID and the user's attributes. |
|---------------------------------|----------------------|---------------------------------------------------------------------------------|

### Returns
A list of keys corresponding to the features that are enabled for the user, or an empty list if no features could be found for the specified user.

### Examples
This section shows a simple example of how you can use the method.
```go
attributes := map[string]interface{}{
        "DEVICE": "iPhone",
        "hey":    2,
}

user := entities.UserContext{
        ID:         "userId",
        Attributes: attributes,
}

isEnabled, err := optlyClient.GetEnabledFeatures(user)
```

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/go-alpha/optimizely/client/client.go#L102).