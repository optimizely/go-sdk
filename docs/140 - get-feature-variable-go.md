---
title: "Get Feature Variable"
excerpt: ""
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
```go
func (o *OptimizelyClient) GetFeatureVariableBoolean(featureKey, variableKey string, userContext entities.UserContext) (value bool, err error)

```

### Double

Returns the value of the specified double variable.
```go
func (o *OptimizelyClient) GetFeatureVariableDouble(featureKey, variableKey string, userContext entities.UserContext) (value float64, err error)

```

### Integer

Returns the value of the specified integer variable.
```go
func (o *OptimizelyClient) GetFeatureVariableInteger(featureKey, variableKey string, userContext entities.UserContext) (value int, err error)

```

### String

Returns the value of the specified string variable.
```go
func (o *OptimizelyClient) GetFeatureVariableString(featureKey, variableKey string, userContext entities.UserContext) (value string, err error)

```

### Version
SDK v1.0

### Description
Each of the Get Feature Variable methods follows the same logic as [Is Feature Enabled](doc:is-feature-enabled-go):
1. Evaluate any feature tests running for a user.
2. Check the default configuration on a rollout.

The default value is returned if neither of these are applicable for the specified user, or if the user is in a variation where the feature is disabled.

>⚠️ Important
>
> Unlike [IsFeatureEnabled](doc:is-feature-enabled-go), the Get Feature Variable methods do not trigger an impression event. This means that if you're running a feature test, events won't be counted until you call IsFeatureEnabled. If you don't call IsFeatureEnabled, you won't see any visitors on your results page.

### Parameters
The table below lists the required and optional parameters in Go.

| Parameter                      | Type                 | Description                                                                                                                        |
|--------------------------------|----------------------|------------------------------------------------------------------------------------------------------------------------------------|
| **featureKey**<br/>*required*  | string               | The feature key is defined from the Features dashboard; see [Use feature flags](doc:use-feature-flags).                            |
| **variableKey**<br/>*required* | string               | The key that identifies the feature variable. For more information, see: [Define feature variables](doc:define-feature-variables). |
| **userContext**<br/>*required* | entities.UserContext | Holds information about the user, such as the userID and the user's attributes.                                                    |

### Returns
The value this method returns is determined by your feature flags. This example shows the specific value returned for Go:

| API                       | Return                                                                                                                                                                            |
|---------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| GetFeatureVariableBoolean | @return The value of the variable, or `false` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.        |
| GetFeatureVariableDouble  | @return The value of the variable, or `0.0` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.          |
| GetFeatureVariableInteger | @return The value of the variable, or `0` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable.            |
| GetFeatureVariableString  | @return The value of the variable, or `empty-string` with error if the feature key is invalid, the variable key is invalid, or there is a mismatch with the type of the variable. |

### Example
This section shows an example of how you can use the `getFeatureVariableInteger` method.

```go
attributes := map[string]interface{}{
        "DEVICE": "iPhone",
        "hey":    2,
}

user := entities.UserContext{
        ID:         "userId",
        Attributes: attributes,
}

value, err := optlyClient.GetFeatureVariableInteger("feature_key", "variable_key", user)

```

### Side effects
The table lists other Optimizely functionality that may be triggered by using this method.

| Functionality | Description                                                                  |
|---------------|------------------------------------------------------------------------------|
| Notifications | Invokes the `DECISION` [notification] is this notification is subscribed to. |

The example code below shows how to add and remove a decision listener.
```go
import (
	"fmt"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/notification"
)

// Callback for decision notification
	callback := func(notification notification.DecisionNotification) {

		// Access type on decisionObject to get type of decision
		fmt.Print(notification.Type)
		// Access decisionInfo on decisionObject which
		// will have form as per type of decision.
		fmt.Print(notification.DecisionInfo)
	}

	optimizelyClient, err := optimizelyFactory.Client()

	// Add callback for decision notification
	id, err := optimizelyClient.DecisionService.OnDecision(callback)

	// Remove callback for decision notification
	err = optimizelyClient.DecisionService.RemoveOnDecision(id)
```

### Exceptions
None

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L160).