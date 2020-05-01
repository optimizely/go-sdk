---
title: "Get Forced Variation"
slug: "get-forced-variation-go"
hidden: true
createdAt: "2019-09-12T14:13:14.683Z"
updatedAt: "2019-12-03T19:25:12.910Z"
---
Returns (string, bool), representing the current forced variation for the argument experiment key and user ID. If no variation was previously set, returns "", false. Otherwise, returns the previously set variation key as the first return value, and true as the second return value.

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
Forced bucketing variations take precedence over whitelisted variations, variations saved in a User Profile Service (if one exists), and the normal bucketed variation. Variations are overwritten when  [Set Variation](doc:set-forced-variation-go) is invoked.
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
    "1-2": "The ID of the user in the forced variation."
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
(string, bool) 
- The key of the currently set forced variation for the argument user and experiment, or an empty string if no forced variation is currently set
- true if a forced variation is currently set for the argument user and experiment, false otherwise

[block:api-header]
{
  "title": "Example"
}
[/block]
Creating a client with Forced Variation service:
[block:code]
{
  "codes": [
    {
      "code": "overrideStore := decision.NewMapExperimentOverridesStore()\nclient, err := optimizelyFactory.Client(\n       client.WithExperimentOverrides(overrideStore),\n)\n\n\n",
      "language": "go"
    }
  ]
}
[/block]
Using Forced Variation service Get Variation:
[block:code]
{
  "codes": [
    {
      "code": "overrideKey := decision.ExperimentOverrideKey{ExperimentKey: \"test_experiment\", UserID: \"test_user\"}\nvariation, success := overrideStore.GetVariation(overrideKey)",
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
[Remove Variation](doc:remove-forced-variation-go)
[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [experiment_override_service.go](https://github.com/optimizely/go-sdk/blob/v1.0.0-beta7/pkg/decision/experiment_override_service.go).