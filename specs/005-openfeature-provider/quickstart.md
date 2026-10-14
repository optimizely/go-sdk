# Quickstart: OpenFeature Provider for Optimizely Go SDK

## Installation

```bash
go get github.com/optimizely/go-sdk/v2
go get github.com/open-feature/go-sdk
```

## Basic Usage (SDK Key)

```go
package main

import (
    "context"
    "fmt"

    "github.com/open-feature/go-sdk/openfeature"
    optimizely "github.com/optimizely/go-sdk/v2/pkg/openfeature"
)

func main() {
    // Create the Optimizely provider with your SDK key
    provider := optimizely.NewProvider("YOUR_SDK_KEY")

    // Register with OpenFeature
    err := openfeature.SetProviderAndWait(provider)
    if err != nil {
        panic(err)
    }
    defer openfeature.Shutdown()

    // Get an OpenFeature client
    client := openfeature.NewClient("my-app")

    // Evaluate a boolean flag
    ctx := context.Background()
    evalCtx := openfeature.NewEvaluationContext("user-123", map[string]any{
        "plan": "enterprise",
    })

    enabled, err := client.BooleanValue(ctx, "my_feature_flag", false, evalCtx)
    fmt.Printf("Feature enabled: %v (err: %v)\n", enabled, err)
}
```

## Typed Variable Evaluation

```go
// String variable — specify which variable via "variableKey"
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]any{
    "variableKey": "banner_text",
})
text, _ := client.StringValue(ctx, "my_feature_flag", "default", evalCtx)

// Integer variable
evalCtx = openfeature.NewEvaluationContext("user-123", map[string]any{
    "variableKey": "max_items",
})
maxItems, _ := client.IntValue(ctx, "my_feature_flag", 10, evalCtx)

// Float variable
evalCtx = openfeature.NewEvaluationContext("user-123", map[string]any{
    "variableKey": "discount_rate",
})
rate, _ := client.FloatValue(ctx, "my_feature_flag", 0.0, evalCtx)

// Object — all variables (no variableKey needed)
evalCtx = openfeature.NewEvaluationContext("user-123", nil)
allVars, _ := client.ObjectValue(ctx, "my_feature_flag", nil, evalCtx)
```

## Pre-initialized Client

```go
import (
    optlyClient "github.com/optimizely/go-sdk/v2/pkg/client"
    optimizely "github.com/optimizely/go-sdk/v2/pkg/openfeature"
)

// Create and configure the Optimizely client yourself
factory := optlyClient.OptimizelyFactory{SDKKey: "YOUR_SDK_KEY"}
optlyInstance, err := factory.Client(
    optlyClient.WithBatchEventProcessor(100, 1000, 30*time.Second),
)
if err != nil {
    panic(err)
}
defer optlyInstance.Close() // You manage the lifecycle

// Wrap it in the OpenFeature provider
provider := optimizely.NewProviderWithClient(optlyInstance)
openfeature.SetProviderAndWait(provider)
defer openfeature.Shutdown()
```

## Event Tracking

```go
evalCtx := openfeature.NewEvaluationContext("user-123", map[string]any{
    "revenue": 49.99,
})
client.Track(ctx, "purchase_completed", evalCtx, openfeature.NewTrackingEventDetails(49.99))
```

## Validation Checklist

- [ ] Boolean evaluation returns correct enabled state
- [ ] String/int/float evaluation returns correct variable values
- [ ] Object evaluation without variableKey returns all variables
- [ ] Provider emits ready event after initialization
- [ ] Shutdown releases resources (SDK-key mode) or detaches
      (pre-initialized mode)
- [ ] Tracking dispatches events to Optimizely
- [ ] Existing Optimizely SDK usage is unaffected
