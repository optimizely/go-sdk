---
title: "Set Forced Variation"
excerpt: ""
slug: "set-forced-variation-go"
hidden: true
createdAt: "2019-09-12T14:11:59.917Z"
updatedAt: "2019-12-03T22:33:30.184Z"
---
The purpose of this method is to force a user into a specific variation for a given experiment. The forced variation value doesn't persist across application launches.

### Version
SDK v1.0.0-beta7

### Description
Forces a user into a variation for a given experiment for the lifetime of the Optimizely client. Any future calls to [Activate](doc:activate), [Is Feature Enabled](doc:is-feature-enabled-go), [Get Feature Variable](doc:get-feature-variable-go), and [Track](doc:track-go) for the given user ID returns the forced variation.

Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Impression and conversion events are still tracked when forced bucketing is enabled.

Variations are overwritten with each set method call. To clear the forced variations so that the normal bucketing flow can occur, use [Remove Variation](doc:remove-forced-variation-go) method of passed ExperimentOverridesStore service. To get the variation that has been forced, use [Get Variation](doc:get-forced-variation-go) method of passed ExperimentOverridesStore service.

### Parameters
This table lists the required parameters for the Go SDK.

| Parameter                                                      | Type                  | Description                                                                                                                                                                                          |
|----------------------------------------------------------------|-----------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **overrideKey**<br/>*All keys inside object are also required* | ExperimentOverrideKey | ExperimentOverrideKey contains,<br/>- *ExperimentKey (string)*: The key of the experiment to set with the forced variation.<br/>- *UserID (string)*: The ID of the user to force into the variation. |
| **variationKey**<br/>*required*                                | string                | The key of the forced variation.                                                                                                                                                                     |

### Example
Creating client with Forced variation service and then setting forced variation:

```go
overrideStore := decision.NewMapExperimentOverridesStore()
client, err := optimizelyFactory.Client(
       client.WithExperimentOverrides(overrideStore),
   )
overrideKey := decision.ExperimentOverrideKey{ExperimentKey: "test_experiment", UserID: "test_user"}
overrideStore.SetVariation(overrideKey, "test_variation")
  
```

### See also
[Get Variation](doc:get-forced-variation-go) 
[Remove Variation](doc:remove-forced-variation-go) 

### Source files
The language/platform source files containing the implementation for Go is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).