---
title: "Example usage"
slug: "example-usage-go"
hidden: true
createdAt: "2019-09-11T22:26:55.008Z"
updatedAt: "2019-12-13T01:40:10.425Z"
---
Once you've installed the Go SDK, import the Optimizely library into your code, get your Optimizely project's datafile, and instantiate a client. Then, you can use the client to evaluate feature flags, activate an A/B test, or feature test.

This example demonstrates the basic usage of each of these concepts. Each concept is described in detail in this guide, and you can find each method's arguments and return values in the [API Reference](doc:activate). 

This example shows how to: 
1. Evaluate a feature with the key `price_filter` and check a configuration variable on it called `min_price`. The SDK evaluates your feature test and rollouts to determine whether the feature is enabled for a particular user, and which minimum price they should see if so.

2. Run an A/B test called `app_redesign`. This experiment has two variations, `control` and `treatment`. It uses the `activate` method to assign the user to a variation, returning its key. As a side effect, the activate function also sends an impression event to Optimizely to record that the current user has been exposed to the experiment. 

3. Use event tracking to track an event called `purchased`. This conversion event measures the impact of an experiment. Using the track method, the purchase is automatically attributed back to the running A/B and feature tests we've activated, and the SDK sends a network request to Optimizely via the customizable event dispatcher so we can count it in your results page.
[block:code]
{
  "codes": [
    {
      "code": "import (\n\toptly \"github.com/optimizely/go-sdk\"\n)\n\nattributes := map[string]interface{}{\n  \"DEVICE\": \"iPhone\",\n  \"hey\":    2,\n}\n\nuser := optly.UserContext(\"userId\", attributes)\n\n// Instantiate an Optimizely client\nif client, err := optly.Client(\"SDK_KEY_HERE\"); err == nil {\n  client.IsFeatureEnabled(\"price_filter\", user)\n  minPrice, _ := client.GetFeatureVariableInteger(\"price_filter\", \"min_price\", user)\n\n  // Activate an A/B test\n  variation, _ := client.Activate(\"app_redesign\", user)\n  if variation == \"control\" {\n    // Execute code for variation A\n  } else if variation == \"treatment\" {\n    // Execute code for variation B\n  } else {\n    // Execute code for users who don't qualify for the experiment\n  }\n  client.Track(\"purchased\", user, map[string]interface{}{})\n}\n",
      "language": "go"
    }
  ]
}
[/block]