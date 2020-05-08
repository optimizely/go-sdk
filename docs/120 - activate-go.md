---
title: "Activate"
excerpt: ""
slug: "activate-go"
hidden: true
createdAt: "2019-09-12T13:58:43.131Z"
updatedAt: "2019-10-29T23:41:25.988Z"
---
Activates an A/B test for the specified user to start an experiment: determines whether they qualify for the experiment, buckets a qualified user into a variation, and sends an impression event to Optimizely.

### Version
SDK v1.0

### Description
This method requires an experiment key, user ID, and (optionally) attributes. The experiment key must match the experiment key you created when you set up the experiment in the Optimizely app. The user ID string uniquely identifies the participant in the experiment. See [Identify users](doc:identify-users) for more information.

If the user qualifies for the experiment, the method returns the variation key that was chosen. If the user was not eligible—for example, because the experiment was not running in this environment or the user didn't match the targeting attributes and audience conditions—then the method returns null.

Activate respects the configuration of the experiment specified in the datafile. The method:
 * Evaluates the user attributes for audience targeting.
 * Includes the user attributes in the impression event to support [results segmentation](doc:analyze-results#section-segment-results).
 * Hashes the user ID or bucketing ID to apply traffic allocation.
 * Respects forced bucketing and whitelisting.
 * Triggers an impression event if the user qualifies for the experiment.

Activate also respects customization of the SDK client. Throughout this process, this method:
  * Logs its decisions via the logger.
  * Triggers impressions via the event dispatcher.
  * Remembers variation assignments via the User Profile Service. **(Coming soon!)**
  * Triggers decision notifications, if subscribed to.

>ℹ️ Note
>
> For more information on how the variation is chosen, see [How bucketing works](how-bucketing-works).

### Parameters
The table below lists the required and optional parameters in Go.

| Parameter                          | Type                 | Description                                                                     |
|------------------------------------|----------------------|---------------------------------------------------------------------------------|
| **experiment key** <br/>*required* | string               | The experiment to activate.                                                     |
| **userContext** <br/>*required*    | entities.UserContext | Holds information about the user, such as the userID and the user's attributes. |

### Returns
A variation key, or an empty string if no experiment was activated.

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

variationKey, err := optlyClient.Activate("experiment_key",user)

```

### See also
[Get Variation](doc:get-variation)
[Identify users](doc:identify-users) 
[How bucketing works](how-bucketing-works)
[Implement impressions](doc:implement-impressions)

### Side effects
The table lists other Optimizely functionality that may be triggered by using this method.

| Functionality | Description                                                                                                                                                                                                                          |
|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Impressions   | Accessing this method triggers an impression if the user is included in an active A/B test.   See [Implement impressions](doc:implement-impressions) for guidance on when to use Activate versus [Get Variation](doc:get-variation). |
| Notifications | Invokes the `DECISION` [notification] if this notification is subscribed to.                                                                                                                                                         |

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

### Notes

### Activate versus Get Variation
Use Activate when the visitor actually sees the experiment. Use Get Variation when you need to know which bucket a visitor is in before showing the visitor the experiment. Impressions are tracked by [Is Feature Enabled](doc:is-feature-enabled-go) when there is a feature test running on the feature and the visitor qualifies for that feature test.

For example, suppose you want your web server to show a visitor variation_1 but don't want the visitor to count until they open a feature that isn't visible when the variation loads, like a modal. In this case, use Get Variation in the backend to specify that your web server should respond with variation_1, and use Activate in the front end when the visitor sees the experiment.

Also, use Get Variation when you're trying to align your Optimizely results with client-side third-party analytics. In this case, use Get Variation to retrieve the variation&mdash;and even show it to the visitor&mdash;but only call Activate when the analytics call goes out.

See [Implement impressions](doc:implement-impressions) for more information about whether to use Activate or Get Variation for a call.

>ℹ️ Note
>
> Conversion events can only be attributed to experiments with previously tracked impressions. Impressions are tracked by Activate, not by Get Variation. As a general rule, Optimizely impressions are required for experiment results and not only for billing.

### Source files
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L46).