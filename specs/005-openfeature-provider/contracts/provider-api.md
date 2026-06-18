# Provider Public API Contract

**Package**: `github.com/optimizely/go-sdk/v2/pkg/openfeature`

## Constructor Functions

```go
// NewProvider creates a provider that creates and owns an Optimizely
// client using the given SDK key. The provider manages the client
// lifecycle (init and shutdown).
func NewProvider(sdkKey string, opts ...ProviderOption) *Provider

// NewProviderWithClient creates a provider that wraps a
// pre-initialized OptimizelyClient. The provider does NOT own the
// client — the caller is responsible for closing it.
func NewProviderWithClient(client *client.OptimizelyClient) *Provider
```

## Provider Options (functional options pattern)

```go
type ProviderOption func(*providerConfig)

// WithClientOptions passes Optimizely factory options to the
// underlying client created by NewProvider. Ignored when using
// NewProviderWithClient.
func WithClientOptions(opts ...client.OptionFunc) ProviderOption
```

## Implemented Interfaces

The `Provider` struct implements these OpenFeature interfaces:

### FeatureProvider (required)

```go
func (p *Provider) Metadata() openfeature.Metadata
func (p *Provider) BooleanEvaluation(ctx context.Context, flag string, defaultValue bool, flatCtx openfeature.FlattenedContext) openfeature.BoolResolutionDetail
func (p *Provider) StringEvaluation(ctx context.Context, flag string, defaultValue string, flatCtx openfeature.FlattenedContext) openfeature.StringResolutionDetail
func (p *Provider) FloatEvaluation(ctx context.Context, flag string, defaultValue float64, flatCtx openfeature.FlattenedContext) openfeature.FloatResolutionDetail
func (p *Provider) IntEvaluation(ctx context.Context, flag string, defaultValue int64, flatCtx openfeature.FlattenedContext) openfeature.IntResolutionDetail
func (p *Provider) ObjectEvaluation(ctx context.Context, flag string, defaultValue any, flatCtx openfeature.FlattenedContext) openfeature.InterfaceResolutionDetail
func (p *Provider) Hooks() []openfeature.Hook
```

### StateHandler (lifecycle)

```go
func (p *Provider) Init(evaluationContext openfeature.EvaluationContext) error
func (p *Provider) Shutdown()
```

### EventHandler (eventing)

```go
func (p *Provider) EventChannel() <-chan openfeature.Event
```

### Tracker (tracking)

```go
func (p *Provider) Track(ctx context.Context, trackingEventName string, evaluationContext openfeature.EvaluationContext, details openfeature.TrackingEventDetails)
```

## Compile-Time Interface Assertions

```go
var _ openfeature.FeatureProvider = (*Provider)(nil)
var _ openfeature.StateHandler    = (*Provider)(nil)
var _ openfeature.EventHandler    = (*Provider)(nil)
var _ openfeature.Tracker         = (*Provider)(nil)
```

## Evaluation Context Conventions

### Reserved Attributes

| Key | Type | Purpose |
| --- | --- | --- |
| `targetingKey` | string | Maps to Optimizely User ID (required) |
| `variableKey` | string | Selects which variable to extract for typed evaluations (optional) |

All other attributes are passed through as Optimizely user
attributes.

## Error Behavior

All evaluation methods return the `defaultValue` on error, along
with a `ProviderResolutionDetail` containing the appropriate
`ResolutionError` and `Reason: openfeature.ErrorReason`.

The provider never panics. All error conditions produce structured
resolution errors.
