---
title: "Get Forced Variation"
excerpt: ""
slug: "get-forced-variation-go"
hidden: true
createdAt: "2019-09-12T14:13:14.683Z"
updatedAt: "2019-12-03T19:25:12.910Z"
---
Returns (string, bool), representing the current forced variation for the argument experiment key and user ID. If no variation was previously set, returns "", false. Otherwise, returns the previously set variation key as the first return value, and true as the second return value.

### Version
SDK v1.0.0-beta7

### Description
Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Variations are overwritten when  [Set Variation](doc:set-forced-variation-go) is invoked.

### Parameters
This table lists the required parameters for the GO SDK.

| Parameter                                                       | Type                  | Description                                                                                                                                                                                            |
|-----------------------------------------------------------------|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **overrideKey**<br/> *All keys inside object are also required* | ExperimentOverrideKey | ExperimentOverrideKey contains, <br/>- *ExperimentKey (string)*: The key of the experiment to set with the forced variation. <br/>- *UserID (string)*: The ID of the user to force into the variation. |

### Returns
(string, bool) 
- The key of the currently set forced variation for the argument user and experiment, or an empty string if no forced variation is currently set
- true if a forced variation is currently set for the argument user and experiment, false otherwise

### Example
Creating a client with Forced Variation service:

```go
overrideStore := decision.NewMapExperimentOverridesStore()
client, err := optimizelyFactory.Client(
       client.WithExperimentOverrides(overrideStore),
)
```

Using Forced Variation service Get Variation:
```go
overrideKey := decision.ExperimentOverrideKey{ExperimentKey: "test_experiment", UserID: "test_user"}
variation, success := overrideStore.GetVariation(overrideKey)
```

### See also
[Set Variation](doc:set-forced-variation-go)
[Remove Variation](doc:remove-forced-variation-go)

### Source files
The language/platform source files containing the implementation for Go is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).