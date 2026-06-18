# Data Model: OpenFeature Provider

**Date**: 2026-04-10
**Feature**: 005-openfeature-provider

## Entities

### Provider

The central adapter struct. Implements `FeatureProvider`,
`StateHandler`, `EventHandler`, and `Tracker`.

**Attributes**:

| Field | Type | Description |
| --- | --- | --- |
| client | *OptimizelyClient | Reference to the underlying Optimizely client |
| ownsClient | bool | Whether the provider created and owns the client (controls shutdown behavior) |
| sdkKey | string | SDK key for client creation (empty if pre-initialized client provided) |
| clientOptions | []OptionFunc | Factory options passed when provider creates the client |
| eventChan | chan Event | Channel for emitting OpenFeature provider events |
| ready | atomic.Bool | Whether the provider is in ready state |

**Lifecycle**:

```
NOT_READY → (Init succeeds) → READY
READY → (config update) → READY (emit ConfigChange event)
READY → (Shutdown) → NOT_READY
NOT_READY → (Init fails) → ERROR
```

**Relationships**:
- Owns or wraps exactly one `OptimizelyClient`
- Listens to the client's `NotificationCenter` for config updates
- Produces events on `eventChan` consumed by the OpenFeature SDK

### Context Mapping

Maps between OpenFeature's `FlattenedContext` and Optimizely's
user model.

| OpenFeature Field | Optimizely Field |
| --- | --- |
| `targetingKey` | UserID (string) |
| All other keys | Attributes (map[string]interface{}) |
| `variableKey` (reserved) | Used by provider to select variable; stripped before passing to Optimizely |

**Validation rules**:
- `targetingKey` MUST be present and non-empty, or evaluation
  returns `TARGETING_KEY_MISSING` error
- `variableKey` is optional; when present it MUST be a string
- Unsupported attribute types (nested objects, slices) are
  silently dropped

### Resolution Detail Mapping

Maps Optimizely `Decide` results to OpenFeature resolution details.

| OpenFeature Field | Source |
| --- | --- |
| Value | Depends on evaluation type (see below) |
| Variant | `OptimizelyDecision.VariationKey` |
| Reason | Mapped from decision state (see research.md) |
| ResolutionError | Mapped from error condition (see research.md) |
| FlagMetadata | Empty (not populated) |

**Value extraction by evaluation type**:

| Evaluation Type | Value Source |
| --- | --- |
| Boolean | `OptimizelyDecision.Enabled` |
| String | `Variables.GetValue(variableKey)` parsed as string |
| Int | `Variables.GetValue(variableKey)` parsed as int64 |
| Float | `Variables.GetValue(variableKey)` parsed as float64 |
| Object | If `variableKey` set: parsed JSON variable. If omitted: full `Variables.ToMap()` |

## State Transitions

### Provider State Machine

```
                 ┌──────────────┐
                 │  NOT_READY   │
                 └──────┬───────┘
                        │ Init()
                 ┌──────┴───────┐
            ┌────│  Initializing │────┐
            │    └──────────────┘    │
         success                   failure
            │                        │
    ┌───────▼──────┐        ┌───────▼──────┐
    │    READY     │        │    ERROR     │
    │              │        └──────────────┘
    │  ┌────────┐  │
    │  │Config  │  │
    │  │Update  │──│ (emit ConfigChange, stay READY)
    │  └────────┘  │
    └──────┬───────┘
           │ Shutdown()
    ┌──────▼───────┐
    │  NOT_READY   │
    └──────────────┘
```
