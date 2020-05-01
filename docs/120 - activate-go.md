---
title: "Activate"
slug: "activate-go"
hidden: true
createdAt: "2019-09-12T13:58:43.131Z"
updatedAt: "2019-10-29T23:41:25.988Z"
---
Activates an A/B test for the specified user to start an experiment: determines whether they qualify for the experiment, buckets a qualified user into a variation, and sends an impression event to Optimizely.
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
[block:callout]
{
  "type": "info",
  "title": "Note",
  "body": "For more information on how the variation is chosen, see [How bucketing works](how-bucketing-works)."
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
    "h-2": "Description",
    "0-0": "**experiment key**\n*required*",
    "0-1": "string",
    "1-0": "**userContext**\n*required*",
    "1-1": "entities.UserContext",
    "0-2": "The experiment to activate.",
    "1-2": "Holds information about the user, such as the userID and the user's attributes."
  },
  "cols": 3,
  "rows": 2
}
[/block]

[block:api-header]
{
  "title": "Returns"
}
[/block]
A variation key, or an empty string if no experiment was activated.
[block:api-header]
{
  "title": "Example"
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "attributes := map[string]interface{}{\n        \"DEVICE\": \"iPhone\",\n        \"hey\":    2,\n}\n\nuser := entities.UserContext{\n        ID:         \"userId\",\n        Attributes: attributes,\n}\n\nvariationKey, err := optlyClient.Activate(\"experiment_key\",user)\n",
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
[Get Variation](doc:get-variation)
[Identify users](doc:identify-users) 
[How bucketing works](how-bucketing-works)
[Implement impressions](doc:implement-impressions)
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
    "0-0": "Impressions",
    "0-1": "Accessing this method triggers an impression if the user is included in an active A/B test. \n\nSee [Implement impressions](doc:implement-impressions) for guidance on when to use Activate versus [Get Variation](doc:get-variation).",
    "1-0": "Notifications",
    "1-1": "Invokes the `DECISION` [notification] if this notification is subscribed to."
  },
  "cols": 2,
  "rows": 2
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
  "title": "Notes"
}
[/block]
### Activate versus Get Variation
Use Activate when the visitor actually sees the experiment. Use Get Variation when you need to know which bucket a visitor is in before showing the visitor the experiment. Impressions are tracked by [Is Feature Enabled](doc:is-feature-enabled-go) when there is a feature test running on the feature and the visitor qualifies for that feature test.

For example, suppose you want your web server to show a visitor variation_1 but don't want the visitor to count until they open a feature that isn't visible when the variation loads, like a modal. In this case, use Get Variation in the backend to specify that your web server should respond with variation_1, and use Activate in the front end when the visitor sees the experiment.

Also, use Get Variation when you're trying to align your Optimizely results with client-side third-party analytics. In this case, use Get Variation to retrieve the variation&mdash;and even show it to the visitor&mdash;but only call Activate when the analytics call goes out.

See [Implement impressions](doc:implement-impressions) for more information about whether to use Activate or Get Variation for a call.
[block:callout]
{
  "type": "info",
  "title": "Note",
  "body": "Conversion events can only be attributed to experiments with previously tracked impressions. Impressions are tracked by Activate, not by Get Variation. As a general rule, Optimizely impressions are required for experiment results and not only for billing."
}
[/block]

[block:api-header]
{
  "title": "Source files"
}
[/block]
The language/platform source files containing the implementation for Go is [Go](https://github.com/optimizely/go-sdk/blob/master/pkg/client/client.go#L46).