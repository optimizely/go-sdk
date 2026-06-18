# Provider Contract: OpenFeature Interface Compliance

**Date**: 2026-04-11  
**Branch**: `006-openfeature-spec-gaps`

## Interface Implementations (unchanged)

The provider implements 4 OpenFeature Go SDK interfaces. These do not change — only the behavior behind them improves.

```
openfeature.FeatureProvider  — Metadata, Hooks, *Evaluation methods
openfeature.StateHandler     — Init, Shutdown
openfeature.EventHandler     — EventChannel
openfeature.Tracker          — Track
```

## Behavioral Contract Changes

### Init() error return

**Before**: Returns `fmt.Errorf("openfeature provider init: %w", err)` (plain error)  
**After**: Returns `&openfeature.ProviderInitError{ErrorCode: openfeature.ProviderFatalCode, Message: err.Error()}`

**Consumer impact**: The OpenFeature SDK will now correctly transition to FATAL state on init failure instead of ERROR state. Consumers registered for `PROVIDER_ERROR` events will see `ErrorCode == "PROVIDER_FATAL"` in event details.

### EventChannel() events

**Before**: `ProviderEventDetails{Message: "..."}` (ErrorCode always zero-value)  
**After**: `ProviderEventDetails{Message: "...", ErrorCode: of.ProviderFatalCode}` on error events

**Consumer impact**: Error event handlers can now inspect `ErrorCode` to distinguish error severity.

### Track() event tags

**Before**: `client.Track(eventName, userContext, contextAttributes)` — tracking details ignored  
**After**: `client.Track(eventName, userContext, mergedEventTags)` where mergedEventTags = contextAttributes + revenue + trackingDetails.Attributes()

**Consumer impact**: Revenue and custom tracking attributes now flow through to Optimizely analytics. Tracking details override context attributes on key conflict.

### Resolution details FlagMetadata

**Before**: `ProviderResolutionDetail{Reason: ..., Variant: ...}` (FlagMetadata nil)  
**After**: `ProviderResolutionDetail{Reason: ..., Variant: ..., FlagMetadata: map[string]interface{}{"flagKey": ..., "ruleKey": ..., "reasons": ...}}`

**Consumer impact**: Consumers can inspect FlagMetadata for debugging and observability. All values are accessible via FlagMetadata's typed accessor methods (GetString, GetInt, etc.).

## Backward Compatibility

All changes are additive:
- Init() still returns `error` interface — the concrete type changes but the interface is the same
- EventChannel() events still have the same structure — ErrorCode was always present but zero-valued
- Track() still calls client.Track — event tags are now richer, not different
- Resolution details still have the same structure — FlagMetadata was always present but nil
- No public API signatures change
- No new exported types or functions are introduced
