# Optimizely OpenFeature Provider for Go

This package provides an [OpenFeature](https://openfeature.dev/) provider that delegates flag evaluation to the [Optimizely Go SDK](https://github.com/optimizely/go-sdk). It allows you to use the OpenFeature API as a standard interface for feature flagging while leveraging Optimizely's experimentation platform underneath.

The provider is safe for concurrent use from multiple goroutines.

## What is OpenFeature?

[OpenFeature](https://openfeature.dev/) is an open specification for feature flag management. It provides a vendor-agnostic API so that applications can evaluate feature flags without being tightly coupled to a specific provider. By coding against the OpenFeature API, you can switch or combine feature flag vendors with minimal code changes.

Key concepts:

- **Provider** — the backend that performs flag evaluation (this package)
- **Client** — the OpenFeature client your application code calls
- **Evaluation Context** — user/request attributes passed to the provider for targeting
- **Tracking** — conversion event reporting for experimentation

## Requirements

- Go 1.21+
- `github.com/optimizely/go-sdk/v2`
- `github.com/open-feature/go-sdk` v1.14+

## Installation

```sh
go get github.com/optimizely/go-sdk/v2
go get github.com/open-feature/go-sdk
```

## Getting Started

There are two ways to set up the provider:

- **Option 1 (provider-managed)** is the simplest — the provider creates and owns the Optimizely client. Use this unless you have a specific reason not to.
- **Option 2 (caller-managed)** gives you direct access to the native Optimizely client alongside OpenFeature. Use this when you need capabilities not exposed through OpenFeature, such as forced decisions, batch evaluation, or notification listeners.

### Option 1: Provider manages the Optimizely client

Pass your SDK key and the provider handles client creation, initialization, and shutdown:

```go
package main

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/openfeature"

	// Aliased because both this package and the OpenFeature SDK share
	// the package name "openfeature".
	optimizely "github.com/optimizely/go-sdk/v2/pkg/openfeature"
)

func main() {
	// Create the provider with your Optimizely SDK key
	provider := optimizely.NewProvider("YOUR_SDK_KEY")

	// Register it with OpenFeature (blocks until the provider is ready)
	if err := openfeature.SetProviderAndWait(provider); err != nil {
		panic(fmt.Sprintf("failed to initialize provider: %v", err))
	}
	defer openfeature.Shutdown()

	// Get a client and evaluate flags
	client := openfeature.NewClient("my-app")
	ctx := context.Background()

	evalCtx := openfeature.NewEvaluationContext("user-123", map[string]interface{}{
		"plan": "enterprise",
	})

	enabled, err := client.BooleanValue(ctx, "checkout_redesign", false, evalCtx)
	if err != nil {
		fmt.Printf("evaluation error: %v\n", err)
	}
	fmt.Printf("checkout_redesign enabled: %v\n", enabled)
}
```

### Option 2: You manage the Optimizely client

Create the client yourself and wrap it. This lets you use the native API alongside OpenFeature:

```go
package main

import (
	"context"
	"fmt"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/optimizely/go-sdk/v2/pkg/client"
	optimizely "github.com/optimizely/go-sdk/v2/pkg/openfeature"
)

func main() {
	// Create and own the Optimizely client
	factory := &client.OptimizelyFactory{SDKKey: "YOUR_SDK_KEY"}
	optlyClient, err := factory.Client()
	if err != nil {
		panic(err)
	}
	defer optlyClient.Close()

	// Wrap it in the OpenFeature provider
	provider := optimizely.NewProviderWithClient(optlyClient)
	if err := openfeature.SetProviderAndWait(provider); err != nil {
		panic(err)
	}
	defer openfeature.Shutdown()

	// Use OpenFeature for standard evaluations
	ofClient := openfeature.NewClient("my-app")
	ctx := context.Background()

	evalCtx := openfeature.NewEvaluationContext("user-123", nil)
	enabled, err := ofClient.BooleanValue(ctx, "checkout_redesign", false, evalCtx)
	if err != nil {
		fmt.Printf("evaluation error: %v\n", err)
	}
	fmt.Printf("OpenFeature: %v\n", enabled)

	// Use the native client directly when needed
	user := optlyClient.CreateUserContext("user-123", nil)
	decision := user.Decide("checkout_redesign", nil)
	fmt.Printf("Native: variationKey=%s, enabled=%v\n", decision.VariationKey, decision.Enabled)
}
```

### Passing Optimizely Client Options

When using `NewProvider`, you can pass Optimizely factory options through:

```go
provider := optimizely.NewProvider("YOUR_SDK_KEY",
	optimizely.WithClientOptions(
		client.WithEventDispatcher(myDispatcher),
		client.WithBatchEventProcessor(10, 100, 0),
	),
)
```

## Evaluation Context

The OpenFeature evaluation context maps to the Optimizely user context as follows:

| OpenFeature | Optimizely | Notes |
|---|---|---|
| `targetingKey` | User ID | **Required.** Must be a non-empty string. |
| All other attributes | User attributes | Scalars only (`string`, `bool`, `int`, `int64`, `float64`). Maps and slices are silently dropped. |
| `variableKey` | *(special)* | Selects a specific feature variable for typed evaluations. Stripped from attributes before passing to Optimizely. |

### Why `variableKey`?

Unlike simple boolean flags, an Optimizely feature can carry multiple typed variables (a string, an int, a JSON object, etc.) under a single flag key. OpenFeature's evaluation methods return a single value, so the provider needs to know *which* variable you want. The `variableKey` attribute in the evaluation context serves this purpose.

Example:

```go
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"plan":        "enterprise",
	"age":         30,
	"variableKey": "banner_text",  // selects which variable to return
})
```

## Flag Evaluation

### Boolean Evaluation

Maps directly to the Optimizely `Decide` API. Returns the decision's `Enabled` field:

```go
enabled, err := client.BooleanValue(ctx, "my_feature", false, evalCtx)
```

This is equivalent to the native SDK's:

```go
user := optlyClient.CreateUserContext("user-123", attrs)
decision := user.Decide("my_feature", nil)
// decision.Enabled
```

### String, Int, and Float Evaluation

These return individual feature variable values. You **must** set `variableKey` in the evaluation context to specify which variable to retrieve:

```go
// String variable
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"variableKey": "banner_text",
})
text, err := client.StringValue(ctx, "my_feature", "default", evalCtx)

// Integer variable
evalCtx = openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"variableKey": "retry_count",
})
count, err := client.IntValue(ctx, "my_feature", int64(3), evalCtx)

// Float variable
evalCtx = openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"variableKey": "discount_rate",
})
rate, err := client.FloatValue(ctx, "my_feature", 0.0, evalCtx)
```

Omitting `variableKey` for string, int, or float evaluations returns an error with the default value.

### Object Evaluation

Returns structured data. Behavior depends on whether `variableKey` is set:

- **With `variableKey`**: returns that specific variable, parsed as JSON if it's a JSON string
- **Without `variableKey`**: returns the full variables map (`map[string]interface{}`)

```go
// Single JSON variable
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"variableKey": "hero_config",
})
config, err := client.ObjectValue(ctx, "my_feature", nil, evalCtx)

// All variables as a map
evalCtx = openfeature.NewEvaluationContext("user-123", nil)
allVars, err := client.ObjectValue(ctx, "my_feature", nil, evalCtx)
```

### Evaluation Details

For richer response data including the variation key, flag metadata, and resolution reason, use the `*ValueDetails` methods:

```go
details, err := client.BooleanValueDetails(ctx, "my_feature", false, evalCtx)
fmt.Printf("Value: %v\n", details.Value)
fmt.Printf("Variant: %s\n", details.Variant)       // Optimizely variation key
fmt.Printf("Reason: %s\n", details.Reason)          // TARGETING_MATCH, DEFAULT, DISABLED, ERROR
fmt.Printf("FlagKey: %v\n", details.FlagMetadata["flagKey"])
fmt.Printf("RuleKey: %v\n", details.FlagMetadata["ruleKey"])
```

## Tracking

Track conversion events for experiments and analytics:

```go
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]interface{}{
	"plan": "enterprise",
})

// Track with revenue
details := openfeature.NewTrackingEventDetails(49.99).
	Add("currency", "USD").
	Add("item_count", 3)

client.Track(ctx, "purchase", evalCtx, details)
```

This is equivalent to the native SDK's:

```go
user := entities.UserContext{ID: "user-123", Attributes: map[string]interface{}{"plan": "enterprise"}}
eventTags := map[string]interface{}{"revenue": 49.99, "currency": "USD", "item_count": 3}
optlyClient.Track("purchase", user, eventTags)
```

Custom attributes from `TrackingEventDetails` override evaluation context attributes on key conflict.

> **Note on revenue:** The `value` parameter in `NewTrackingEventDetails(value)` is mapped to the `"revenue"` event tag only when it is non-zero. If you pass `0` (e.g., for a free-tier conversion), the `"revenue"` tag is omitted entirely. To explicitly send a zero revenue, set it as a custom attribute: `details.Add("revenue", 0)`.

## Events

The provider emits OpenFeature lifecycle events that you can listen for:

```go
readyCb := func(details openfeature.EventDetails) {
	fmt.Printf("Provider %s is ready\n", details.ProviderName)
}
errorCb := func(details openfeature.EventDetails) {
	fmt.Printf("Provider %s error: %s\n", details.ProviderName, details.Message)
}
configCb := func(details openfeature.EventDetails) {
	fmt.Printf("Provider %s config changed: %s\n", details.ProviderName, details.Message)
}

openfeature.AddHandler(openfeature.ProviderReady, &readyCb)
openfeature.AddHandler(openfeature.ProviderError, &errorCb)
openfeature.AddHandler(openfeature.ProviderConfigurationChanged, &configCb)
```

| Event | When |
|---|---|
| `ProviderReady` | Optimizely client initialized successfully |
| `ProviderError` | Client initialization failed (with `PROVIDER_FATAL` error code) |
| `ProviderConfigurationChanged` | Optimizely datafile was updated (polling or push) |

## Native SDK Comparison

| Capability | Native Optimizely SDK | OpenFeature Provider |
|---|---|---|
| Boolean flag (enabled/disabled) | `user.Decide("flag", nil).Enabled` | `client.BooleanValue(ctx, "flag", false, evalCtx)` |
| Variation key | `user.Decide("flag", nil).VariationKey` | `details.Variant` via `BooleanValueDetails` |
| String variable | `decision.Variables.ToMap()["key"]` | `client.StringValue(ctx, "flag", "", evalCtx)` with `variableKey` in context |
| Integer variable | `decision.Variables.ToMap()["key"]` | `client.IntValue(ctx, "flag", 0, evalCtx)` with `variableKey` in context |
| Float variable | `decision.Variables.ToMap()["key"]` | `client.FloatValue(ctx, "flag", 0.0, evalCtx)` with `variableKey` in context |
| JSON variable | `decision.Variables.GetValue("key", &v)` | `client.ObjectValue(ctx, "flag", nil, evalCtx)` with `variableKey` in context |
| All variables map | `decision.Variables.ToMap()` | `client.ObjectValue(ctx, "flag", nil, evalCtx)` without `variableKey` |
| Track event | `optlyClient.Track(eventKey, user, eventTags)` | `client.Track(ctx, eventKey, evalCtx, details)` |
| Decision reasons | `user.Decide("flag", []decide.OptimizelyDecideOptions{decide.IncludeReasons}).Reasons` | `details.FlagMetadata["reasons"]` (always included) |
| Forced decisions | `user.SetForcedDecision(...)` | Not supported through OpenFeature |
| Multiple flag decisions | `user.DecideForKeys(...)` / `user.DecideAll(...)` | Not supported — evaluate one flag at a time |
| User profile service | `client.WithUserProfileService(...)` | Pass via `WithClientOptions(client.WithUserProfileService(...))` |
| Notification listeners | `client.GetNotificationCenter().AddHandler(...)` | Use `NewProviderWithClient` and access the client directly |

## Caveats

1. **`variableKey` is required for typed evaluations.** String, int, and float evaluations require a `variableKey` attribute in the evaluation context to select which of the feature's variables to return. Without it, the provider returns an error and the default value. Object evaluation is the exception — omitting `variableKey` returns the full variables map.

2. **Attributes must be scalar types.** The Optimizely SDK only supports `string`, `bool`, `int`, `int64`, and `float64` as user attributes. Maps, slices, and other complex types in the evaluation context are silently dropped.

3. **Decision reasons are always requested.** The provider calls `Decide` with `IncludeReasons` on every evaluation so that reasons are available in `FlagMetadata["reasons"]`. This adds minor overhead compared to native calls without reasons.

4. **Forced decisions and batch evaluation are not available.** OpenFeature's API evaluates one flag at a time and does not expose Optimizely's `SetForcedDecision`, `DecideForKeys`, or `DecideAll` methods. Use `NewProviderWithClient` and access the native client directly for these capabilities.

5. **Shutdown behavior depends on construction mode.** When using `NewProvider`, calling `openfeature.Shutdown()` closes the underlying Optimizely client. When using `NewProviderWithClient`, the caller retains lifecycle ownership — the provider's shutdown does not close the client.

6. **Unknown flag keys return `FLAG_NOT_FOUND`.** If a flag key doesn't exist in your Optimizely project, the provider returns the default value with a `FLAG_NOT_FOUND` error code.

## License

This project is licensed under the Apache 2.0 License. See [LICENSE](../../LICENSE) for details.
