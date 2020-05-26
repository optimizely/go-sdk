---
title: "Remove Forced Variation"
excerpt: ""
slug: "remove-forced-variation-go"
hidden: true
createdAt: "2019-12-03T19:23:24.666Z"
updatedAt: "2019-12-03T19:26:39.873Z"
---
Clears the forced variation set by [Set Variation](doc:set-forced-variation-go), so that the normal bucketing flow can occur.
### Version
SDK v1.0.0-beta7

### Description
Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Variations are overwritten when  [Set Forced Variation](doc:set-forced-variation-go) is invoked.

### Parameters
This table lists the required parameters for the GO SDK.

| Parameter                                                      | Type                  | Description                                                                                                                                                                                          |
|----------------------------------------------------------------|-----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **overrideKey**<br/>*All keys inside object are also required* | ExperimentOverrideKey | ExperimentOverrideKey contains,<br/>- *ExperimentKey (string)*: The key of the experiment to set with the forced variation.<br/>- *UserID (string)*: The ID of the user to force into the variation. |

### Example
Creating a client with Forced Variations:

```go
overrideStore := decision.NewMapExperimentOverridesStore()
client, err := optimizelyFactory.Client(
       client.WithExperimentOverrides(overrideStore),
   )
```

Removing a forced variation using Remove Forced Variation:
```go
overrideKey := decision.ExperimentOverrideKey{ExperimentKey: "test_experiment", UserID: "test_user"}
overrideStore.RemoveVariation(overrideKey)
```

### See also
[Set Variation](doc:set-forced-variation-go) 
[Get Variation](doc:get-forced-variation-go) 

### Source files
The language/platform source files containing the implementation for GO is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).