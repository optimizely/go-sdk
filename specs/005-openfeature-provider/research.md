# Research: OpenFeature Provider

**Date**: 2026-04-10
**Feature**: 005-openfeature-provider

## OpenFeature Go SDK Interface Requirements

### Decision: Target OpenFeature Go SDK v1.17.x

**Rationale**: v1.17.2 is the latest stable release. The module path
is `github.com/open-feature/go-sdk` (no `/v2` suffix). The provider
interface has been stable across v1.x releases.

**Alternatives considered**: None viable — v1.x is the only stable
release line.

### Decision: Implement four opt-in interfaces

The provider MUST implement:

1. `FeatureProvider` (required) — 5 evaluation methods + Metadata +
   Hooks
2. `StateHandler` (opt-in) — Init/Shutdown lifecycle
3. `EventHandler` (opt-in) — event channel for state transitions
4. `Tracker` (opt-in) — conversion tracking

**Rationale**: These four interfaces cover all functionality mapped
in the spec (evaluation, lifecycle, events, tracking). The OpenFeature
SDK discovers capabilities via interface assertion — implementing
all four in a single struct is the standard pattern.

**Alternatives considered**: `ContextAwareStateHandler` adds
context-aware init/shutdown but adds complexity without clear benefit
for the Optimizely use case. Can be added later if needed.

## Provider Patterns from Existing Implementations

### Decision: Single-struct provider with compile-time checks

Follow the LaunchDarkly pattern: one `Provider` struct implements all
interfaces. Use compile-time interface assertions:

```go
var _ openfeature.FeatureProvider = (*Provider)(nil)
var _ openfeature.StateHandler = (*Provider)(nil)
var _ openfeature.EventHandler = (*Provider)(nil)
var _ openfeature.Tracker = (*Provider)(nil)
```

**Rationale**: Simpler than flagd's multi-service abstraction. The
Optimizely SDK already handles all complexity internally — the
provider is a thin adapter.

**Alternatives considered**: flagd-style service abstraction rejected
as over-engineering for a single-backend provider.

### Decision: Background goroutine for event forwarding

Follow the flagd pattern: `Init()` starts a background goroutine that
listens to the Optimizely notification center for `ProjectConfigUpdate`
notifications and forwards them as `ProviderConfigChange` events on
the OpenFeature event channel.

**Rationale**: The Optimizely SDK emits config updates via its
notification center. A goroutine bridge is the only way to forward
these as OpenFeature channel-based events.

### Decision: Dual construction modes

Support both SDK-key-based and pre-initialized-client construction,
following the LaunchDarkly pattern where `Shutdown()` conditionally
closes the client based on an `ownsClient` flag.

**Rationale**: Clarified in spec. Matches enterprise usage patterns
where clients are shared across subsystems.

## Reason Code Mapping

### Decision: Map based on decision state

| Optimizely Decision State | OpenFeature Reason |
| --- | --- |
| VariationKey set, Enabled=true | `TARGETING_MATCH` |
| VariationKey set, Enabled=false | `DISABLED` |
| VariationKey empty (no decision) | `DEFAULT` |
| Error during evaluation | `ERROR` |
| All other cases | `UNKNOWN` |

**Rationale**: Maps directly to OpenFeature semantics. No granularity
lost since raw Optimizely reasons are free-text and not structured.

## Error Code Mapping

### Decision: Map Optimizely errors to OpenFeature error codes

| Condition | OpenFeature Error Code |
| --- | --- |
| Flag key not in datafile | `FLAG_NOT_FOUND` |
| Targeting key missing from context | `TARGETING_KEY_MISSING` |
| Variable type mismatch | `TYPE_MISMATCH` |
| Variable parse failure | `PARSE_ERROR` |
| Client not initialized | `PROVIDER_NOT_READY` |
| Other errors | `GENERAL` |

**Rationale**: Direct semantic match. The Optimizely SDK surfaces
flag-not-found via empty decision with error reasons; client readiness
is checked before evaluation.

## Variable Extraction Strategy

### Decision: Use `variableKey` from FlattenedContext

For typed evaluations, extract `variableKey` from the flattened
context attributes. Use it to look up a specific variable in the
`Decide` result's `Variables` map.

- If `variableKey` is present: extract that variable, parse to
  requested type.
- If `variableKey` is absent and evaluation is ObjectEvaluation:
  return the full variables map.
- If `variableKey` is absent and evaluation is scalar: return
  default value with `GENERAL` error code.

**Rationale**: Clarified in spec session. Flexible, explicit, and
avoids ambiguity with multi-variable flags.

## Dependency Impact

### Decision: Add `github.com/open-feature/go-sdk` to go.mod

**Rationale**: The provider lives in `pkg/openfeature/` within the
existing module. The OpenFeature SDK dependency is Apache 2.0
licensed — compatible with the project license. Applications that
don't import `pkg/openfeature/` won't link the code, though it
appears in the module graph.

**Alternatives considered**: Separate Go module rejected during
clarification — adds multi-module complexity without sufficient
benefit.
