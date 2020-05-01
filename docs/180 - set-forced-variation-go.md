---
title: "Set Forced Variation"
slug: "set-forced-variation-go"
hidden: true
createdAt: "2019-09-12T14:11:59.917Z"
updatedAt: "2019-12-03T22:33:30.184Z"
---
The purpose of this method is to force a user into a specific variation for a given experiment. The forced variation value doesn't persist across application launches.
[block:api-header]
{
  "title": "Version"
}
[/block]
SDK v1.0.0-beta7
[block:api-header]
{
  "title": "Description"
}
[/block]
Forces a user into a variation for a given experiment for the lifetime of the Optimizely client. Any future calls to [Activate](doc:activate), [Is Feature Enabled](doc:is-feature-enabled-go), [Get Feature Variable](doc:get-feature-variable-go), and [Track](doc:track-go) for the given user ID returns the forced variation.

Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Impression and conversion events are still tracked when forced bucketing is enabled.

Variations are overwritten with each set method call. To clear the forced variations so that the normal bucketing flow can occur, use [Remove Variation](doc:remove-forced-variation-go) method of passed ExperimentOverridesStore service. To get the variation that has been forced, use [Get Variation](doc:get-forced-variation-go) method of passed ExperimentOverridesStore service.
[block:api-header]
{
  "title": "Parameters"
}
[/block]
This table lists the required parameters for the Go SDK.
[block:parameters]
{
  "data": {
    "h-0": "Parameter",
    "h-1": "Type",
    "h-2": "Description",
    "0-0": "**overrideKey**\n*All keys inside object are also required*",
    "0-1": "ExperimentOverrideKey",
    "1-0": "**variationKey**\n*required*",
    "1-1": "string",
    "0-2": "ExperimentOverrideKey contains,\n- *ExperimentKey (string)*: The key of the experiment to set with the forced variation.\n- *UserID (string)*: The ID of the user to force into the variation.",
    "1-2": "The key of the forced variation.",
    "2-2": "The key of the forced variation. Set the value to `null` to clear the existing experiment-to-variation mapping.",
    "2-0": "**variationKey**\n*optional*",
    "2-1": "string"
  },
  "cols": 3,
  "rows": 2
}
[/block]

[block:api-header]
{
  "title": "Example"
}
[/block]
Creating client with Forced variation service and then setting forced variation:
[block:code]
{
  "codes": [
    {
      "code": "overrideStore := decision.NewMapExperimentOverridesStore()\nclient, err := optimizelyFactory.Client(\n       client.WithExperimentOverrides(overrideStore),\n   )\noverrideKey := decision.ExperimentOverrideKey{ExperimentKey: \"test_experiment\", UserID: \"test_user\"}\noverrideStore.SetVariation(overrideKey, \"test_variation\")\n  ",
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
[Get Variation](doc:get-forced-variation-go) 
[Remove Variation](doc:remove-forced-variation-go) 
[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).