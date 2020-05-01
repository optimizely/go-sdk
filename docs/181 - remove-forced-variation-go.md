---
title: "Remove Forced Variation"
slug: "remove-forced-variation-go"
hidden: true
createdAt: "2019-12-03T19:23:24.666Z"
updatedAt: "2019-12-03T19:26:39.873Z"
---
Clears the forced variation set by [Set Variation](doc:set-forced-variation-go), so that the normal bucketing flow can occur.
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
Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Variations are overwritten when  [Set Forced Variation](doc:set-forced-variation-go) is invoked.
[block:api-header]
{
  "title": "Parameters"
}
[/block]
This table lists the required parameters for the GO SDK.
[block:parameters]
{
  "data": {
    "h-0": "Parameter",
    "h-1": "Type",
    "h-2": "Description",
    "0-0": "**overrideKey**\n*All keys inside object are also required*",
    "0-1": "ExperimentOverrideKey",
    "1-0": "**userId**\n*required*",
    "1-1": "string",
    "0-2": "ExperimentOverrideKey contains,\n- *ExperimentKey (string)*: The key of the experiment to set with the forced variation.\n- *UserID (string)*: The ID of the user to force into the variation.",
    "1-2": "The ID of the user to force into the variation.",
    "2-2": "The key of the forced variation. Set the value to `null` to clear the existing experiment-to-variation mapping.",
    "2-0": "**variationKey**\n*optional*",
    "2-1": "string"
  },
  "cols": 3,
  "rows": 1
}
[/block]

[block:api-header]
{
  "title": "Example"
}
[/block]
Creating a client with Forced Variations:
[block:code]
{
  "codes": [
    {
      "code": "overrideStore := decision.NewMapExperimentOverridesStore()\nclient, err := optimizelyFactory.Client(\n       client.WithExperimentOverrides(overrideStore),\n   )\n\n  \n  ",
      "language": "go"
    }
  ]
}
[/block]
Removing a forced variation using Remove Forced Variation:
[block:code]
{
  "codes": [
    {
      "code": "overrideKey := decision.ExperimentOverrideKey{ExperimentKey: \"test_experiment\", UserID: \"test_user\"}\noverrideStore.RemoveVariation(overrideKey)",
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
[Set Variation](doc:set-forced-variation-go) 
[Get Variation](doc:get-forced-variation-go) 
[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for GO is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).