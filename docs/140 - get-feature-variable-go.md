---
title: "Get Feature Variable"
slug: "get-feature-variable-go"
hidden: true
createdAt: "2019-09-12T14:11:43.293Z"
updatedAt: "2019-10-29T23:43:46.140Z"
---
Evaluates the specified feature variable of a specific variable type and returns its value.  

This method is used to evaluate and return a feature variable. Multiple versions of this method are available and are named according to the data type they return:
  * [Boolean](#section-boolean)
  * [Double](#section-double)
  * [Integer](#section-integer)
  * [String](#section-string)

This method takes into account the user `attributes` passed in, to determine if the user is part of the audience that qualifies for the experiment.

### Boolean

Returns the value of the specified Boolean variable.
[block:code]
{
  "codes": [
    {
      "code": "func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey, variableKey string, userContext entities.UserContext) (value bool, err error)\n\n",
      "language": "go"
    }
  ]
}
[/block]
### Double

Returns the value of the specified double variable.
[block:code]
{
  "codes": [
    {
      "code": "func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey, variableKey string, userContext entities.UserContext) (value float64, err error)\n\n",
      "language": "go"
    }
  ]
}
[/block]
### Integer

Returns the value of the specified integer variable.
[block:code]
{
  "codes": [
    {
      "code": "func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey, variableKey string, userContext entities.UserContext) (value int, err error)\n\n",
      "language": "go"
    }
  ]
}
[/block]
### String

Returns the value of the specified string variable.
[block:code]
{
  "codes": [
    {
      "code": "func (o *OptimizelyClient) GetFeatureVariableString(featureKey, variableKey string, userContext entities.UserContext) (value string, err error)\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Version"
}
[/block]
SDK v1.0
[block:api-header]
{
  "title": "Description"
}
[/block]
Each of the Get Feature Variable methods follows the same logic as [Is Feature Enabled](doc:is-feature-enabled-go):
1. Evaluate any feature tests running for a user.
2. Check the default configuration on a rollout.

The default value is returned if neither of these are applicable for the specified user, or if the user is in a variation where the feature is disabled.
[block:callout]
{
  "type": "warning",
  "title": "Important",
  "body": "Unlike [IsFeatureEnabled](doc:is-feature-enabled-go), the Get Feature Variable methods do not trigger an impression event. This means that if you're running a feature test, events won't be counted until you call IsFeatureEnabled. If you don't call IsFeatureEnabled, you won't see any visitors on your results page."
}
[/block]

[block:api-header]
{
  "title": "Parameters"
}
[/block]
The table below lists the required and optional parameters in Go.
[block:parameters]
{
  "data": {
    "h-0": "Parameter",
    "h-1": "Type",
    "0-0": "**featureKey**\n*required*",
    "0-1": "string",
    "1-0": "**variableKey**\n*required*",
    "1-1": "string",
    "h-2": "Description",
    "0-2": "The feature key is defined from the Features dashboard; see [Use feature flags](doc:use-feature-flags).",
    "1-2": "The key that identifies the feature variable. For more information, see: [Define feature variables](doc:define-feature-variables).",
    "2-0": "**userContext**\n*required*",
    "2-1": "entities.UserContext",
    "2-2": "Holds information about the user, such as the userID and the user's attributes."
  },
  "cols": 3,
  "rows": 3
}
[/block]

[block:api-header]
{
  "title": "Returns"
}
[/block]
The value this method returns is determined by your feature flags. This example shows the specific value returned for Go:
[block:parameters]
{
  "data": {
    "h-0": "API",
    "h-1": "Return",
    "0-0": "GetFeatureVariableBoolean",
    "1-0": "GetFeatureVariableDouble",
    "2-0": "GetFeatureVariableInteger",
    "3-0": "GetFeatureVariableString",
    "0-1": "@return The value of the variable, or `false` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.",
    "1-1": "@return The value of the variable, or `0.0` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.",
    "2-1": "@return The value of the variable, or `0` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.",
    "3-1": "@return The value of the variable, or `empty-string` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable."
  },
  "cols": 2,
  "rows": 4
}
[/block]

[block:api-header]
{
  "title": "Example"
}
[/block]
This section shows an example of how you can use the `getFeatureVariableInteger` method.
[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nvalue, err := optlyClient.GetFeatureVariableInteger(\"feature_key\", \"variable_key\", user)\n\n",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Side effects"
}
[/block]
The table lists other Optimizely functionality that may be triggered by using this method.
[block:parameters]
{
  "data": {
    "h-0": "Functionality",
    "h-1": "Description",
    "0-0": "Notifications",
    "0-1": "Invokes the `DECISION` [notification] is this notification is subscribed to."
  },
  "cols": 2,
  "rows": 1
}
[/block]
The example code below shows how to add and remove a decision listener.
[block:code]
{
  "codes": [
    {
      "code": "import (\n\t\"fmt\"\n\n\t\"github.com/optimizely/go-sdk/pkg/client\"\n\t\"github.com/optimizely/go-sdk/pkg/notification\"\n)\n\n// Callback for decision notification\n\tcallback := func(notification notification.DecisionNotification) {\n\n\t\t// Access type on decisionObject to get type of decision\n\t\tfmt.Print(notification.Type)\n\t\t// Access decisionInfo on decisionObject which\n\t\t// will have form as per type of decision.\n\t\tfmt.Print(notification.DecisionInfo)\n\t}\n\n\toptimizelyClient, err := optimizelyFactory.Client()\n\n\t// Add callback for decision notification\n\tid, err := optimizelyClient.DecisionService.OnDecision(callback)\n\n\t// Remove callback for decision notification\n\terr = optimizelyClient.DecisionService.RemoveOnDecision(id)",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Exceptions"
}
[/block]
None
[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L160).