# Data Model: OpenFeature Provider Spec Compliance Gaps

**Date**: 2026-04-11  
**Branch**: `006-openfeature-spec-gaps`

## Modified Entities

### Provider (existing: `pkg/openfeature/provider.go`)

No structural changes. Existing fields are sufficient.

### emitEvent Helper (existing: `pkg/openfeature/lifecycle.go`)

**Current signature**: `emitEvent(eventType of.EventType, message string)`

**Change**: Add `errorCode` parameter to support populating `ProviderEventDetails.ErrorCode`.

**New signature**: `emitEvent(eventType of.EventType, message string, errorCode of.ErrorCode)`

### evaluationDetail (existing: `pkg/openfeature/evaluation.go`)

**Current fields**:
- `decision decisionResult`

**Change**: Extend `decisionResult` to include `RuleKey` and `Reasons` for FlagMetadata population.

**New decisionResult fields**:

| Field        | Type       | Source                          | Purpose                 |
|-------------|------------|----------------------------------|-------------------------|
| VariationKey | string     | OptimizelyDecision.VariationKey  | Existing                |
| Enabled      | bool       | OptimizelyDecision.Enabled       | Existing                |
| Variables    | variablesAccessor | OptimizelyDecision.Variables | Existing               |
| FlagKey      | string     | OptimizelyDecision.FlagKey       | Existing                |
| RuleKey      | string     | OptimizelyDecision.RuleKey       | New — for FlagMetadata  |
| Reasons      | []string   | OptimizelyDecision.Reasons       | New — for FlagMetadata  |

### toProviderDetail (existing: `pkg/openfeature/evaluation.go`)

**Change**: Populate `FlagMetadata` with decision context.

**FlagMetadata keys**:

| Key         | Type      | Source                   | Always present? |
|-------------|-----------|--------------------------|-----------------|
| `flagKey`   | string    | decisionResult.FlagKey   | Yes             |
| `ruleKey`   | string    | decisionResult.RuleKey   | Only if non-empty |
| `reasons`   | []string  | decisionResult.Reasons   | Only if non-empty |

### Track (existing: `pkg/openfeature/tracking.go`)

**Change**: Build event tags from both evaluation context attributes and tracking event details.

**Event tag merge logic**:

1. Start with evaluation context attributes as base event tags
2. Add `"revenue"` key from `TrackingEventDetails.Value()` (always, including 0.0)
3. Merge `TrackingEventDetails.Attributes()` — these override context attributes on key conflict
4. Pass merged map as event tags to `client.Track`

### Init (existing: `pkg/openfeature/lifecycle.go`)

**Change**: Return `*of.ProviderInitError` instead of plain `error` on failure.

**Error construction**:
```
ProviderInitError{
    ErrorCode: ProviderFatalCode,
    Message:   "<original error message>",
}
```

## Removed/Deferred Entities

### ProviderStale event emission

**Status**: Deferred. No upstream notification type supports connectivity failure detection. See research.md R6.
