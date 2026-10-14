# Research: OpenFeature Provider Spec Compliance Gaps

**Date**: 2026-04-11  
**Branch**: `006-openfeature-spec-gaps`

## R1: TrackingEventDetails API Surface

**Decision**: Use `TrackingEventDetails.Value()` for numeric revenue and `TrackingEventDetails.Attributes()` for custom fields.

**Rationale**: `TrackingEventDetails` is a concrete struct (not an interface) in the OpenFeature Go SDK v1.14.1 with these methods:
- `Value() float64` — returns the numeric tracking value
- `Attributes() map[string]interface{}` — returns a copy of custom attributes
- `Attribute(key string) interface{}` — retrieves a single attribute
- `Add(key string, value interface{}) TrackingEventDetails` — builder method

**Alternatives considered**: None — these are the only accessors available.

## R2: Revenue Event Tag Convention

**Decision**: Use `"revenue"` as the event tag key for the numeric tracking value.

**Rationale**: The Optimizely SDK defines `const revenueKey = "revenue"` in `pkg/event/factory.go:38`. The `getRevenueValue()` function extracts this key from event tags. The event structs use `Revenue *int64` (pointer to distinguish 0 from absent). We will pass `Value()` as a float64 under the `"revenue"` key — the Optimizely event processor handles the float-to-int conversion internally.

**Alternatives considered**: Using `"value"` — rejected because the Optimizely backend expects `"revenue"` as the canonical key.

## R3: ProviderInitError Type

**Decision**: Return `&openfeature.ProviderInitError{ErrorCode: openfeature.ProviderFatalCode, Message: ...}` from `Init()` on failure.

**Rationale**: `ProviderInitError` is defined in `resolution_error.go` with fields `ErrorCode ErrorCode` and `Message string`. It implements the `error` interface. The OpenFeature Go SDK's `initializer()` function in `openfeature_api.go` uses `errors.As(err, &initErr)` to detect this type and reads `initErr.ErrorCode` to determine FATAL vs ERROR state. `ProviderFatalCode` is the constant `"PROVIDER_FATAL"`.

**Alternatives considered**: Wrapping with `fmt.Errorf` — rejected because the Go SDK uses `errors.As` which requires the concrete type in the error chain.

## R4: ProviderEventDetails ErrorCode Field

**Decision**: Set `ProviderEventDetails.ErrorCode` to the appropriate `ErrorCode` constant on all error events.

**Rationale**: `ProviderEventDetails` has an `ErrorCode ErrorCode` field (same type as resolution error codes). The Go SDK's `stateFromEvent()` checks this field for `ProviderFatalCode` to determine FATAL vs ERROR state. Currently our `emitEvent` helper only sets `Message`.

**Alternatives considered**: None — this is the required mechanism.

## R5: OptimizelyDecision Fields for FlagMetadata

**Decision**: Populate FlagMetadata with `flagKey`, `ruleKey`, and `reasons` from `OptimizelyDecision`.

**Rationale**: `OptimizelyDecision` exposes:
- `FlagKey string` — always populated
- `RuleKey string` — populated when a rule matches (experiment or rollout rule)
- `Reasons []string` — populated when `decide.IncludeReasons` is used
- `VariationKey string` — already exposed as `Variant` on `ProviderResolutionDetail`

`FlagMetadata` is `map[string]interface{}` with typed accessor methods (`GetString`, `GetBool`, `GetInt`, `GetFloat`).

**Alternatives considered**: Including `ExperimentID` — rejected because it is not exposed on the public `OptimizelyDecision` type.

## R6: Stale Provider Signaling Feasibility

**Decision**: Defer User Story 5 (PROVIDER_STALE). The Optimizely SDK notification center does not provide connectivity failure signals.

**Rationale**: The notification center defines 4 types: `Decision`, `Track`, `ProjectConfigUpdate`, `LogEvent`. None signal connectivity failures or staleness. `ProjectConfigUpdate` only fires on successful config refreshes. Detecting staleness would require either:
1. A timeout-based approach (no `ProjectConfigUpdate` for N intervals), or
2. Upstream SDK changes to add a failure notification type.

Option 1 is fragile (static configs never update and would always appear stale). Option 2 is out of scope.

**Alternatives considered**: Timer-based staleness detection — rejected as unreliable for static config managers.
