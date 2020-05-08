---
title: "Get Variation"
excerpt: ""
slug: "get-variation-go"
hidden: true
createdAt: "2019-09-12T14:11:49.333Z"
updatedAt: "2019-10-29T23:44:33.591Z"
---
Returns a variation where the visitor will be bucketed, without triggering an impression.
### Version
SDK v1.0

### Description
Takes the same arguments and returns the same values as [Activate](doc:activate), but without sending an impression event. The behavior of the two methods is identical otherwise. 

Use **GetVariation** if **Activate** has been called and the current variation assignment is needed for a given experiment and user. This method bypasses redundant network requests to Optimizely.

See [Implement impressions](doc:implement-impressions) for guidance on when to use each method.
### Parameters

The table below lists the required and optional parameters in Go.

| Parameter                          | Type                 | Description                                                                     |
|------------------------------------|----------------------|---------------------------------------------------------------------------------|
| **experimentKey** <br/> *required* | string               | The key of the experiment.                                                      |
| **userContext**<br/>*required*     | entities.UserContext | Holds information about the user, such as the userID and the user's attributes. |

### Returns
A variation key, or an empty string if no variation is found.

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

variationKey, err := optlyClient.GetVariation("experiment_key", user)

```

### See also
[Activate](doc:activate) 
[Implement impressions](doc:implement-impressions)

### Notes

### Activate versus Get Variation
Use Activate when the visitor actually sees the experiment. Use Get Variation when you need to know which bucket a visitor is in before showing the visitor the experiment. Impressions are tracked by [Is Feature Enabled](doc:is-feature-enabled-go) when there is a feature test running on the feature and the visitor qualifies for that feature test.

For example, suppose you want your web server to show a visitor variation_1 but don't want the visitor to count until they open a feature that isn't visible when the variation loads, like a modal. In this case, use Get Variation in the backend to specify that your web server should respond with variation_1, and use Activate in the front end when the visitor sees the experiment.

Also, use Get Variation when you're trying to align your Optimizely results with client-side third-party analytics. In this case, use Get Variation to retrieve the variation&mdash;and even show it to the visitor&mdash;but only call Activate when the analytics call goes out.

See [Implement impressions](doc:implement-impressions) for more information about whether to use Activate or Get Variation for a call.

>⚠️ Important
>
> Conversion events can only be attributed to experiments with previously tracked impressions. Impressions are tracked by Activate, not by Get Variation. As a general rule, Optimizely impressions are required for experiment results and not only for billing.

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L265).