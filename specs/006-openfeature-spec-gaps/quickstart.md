# Quickstart: OpenFeature Provider Spec Compliance Gaps

**Branch**: `006-openfeature-spec-gaps`

## What's changing

This feature closes 5 spec compliance gaps in the OpenFeature provider (`pkg/openfeature/`):

1. **Error events now carry ErrorCode** — enables the OpenFeature SDK to distinguish FATAL from ERROR provider state
2. **Init failures return typed errors** — `*ProviderInitError` with `ProviderFatalCode` instead of plain `error`
3. **Track forwards revenue and custom attributes** — `TrackingEventDetails.Value()` and `.Attributes()` are merged into Optimizely event tags
4. **Resolution details include FlagMetadata** — `flagKey`, `ruleKey`, and `reasons` populated on successful evaluations
5. **PROVIDER_STALE deferred** — Optimizely SDK notification center doesn't support connectivity failure signals

## Files modified

```
pkg/openfeature/lifecycle.go       — Init error typing, emitEvent ErrorCode
pkg/openfeature/tracking.go        — Event tag merge with tracking details
pkg/openfeature/evaluation.go      — FlagMetadata population, decisionResult fields
pkg/openfeature/lifecycle_test.go   — Init error type assertions
pkg/openfeature/tracking_test.go    — Revenue and attribute forwarding tests
pkg/openfeature/evaluation_test.go  — FlagMetadata assertions
```

## How to verify

```bash
# Run all OpenFeature tests
go test ./pkg/openfeature/... -v -count=1

# Run full test suite (no regressions)
make test

# Lint
make lint
```

## Key decisions

- All init failures are FATAL (provider can't function without a client)
- Tracking details override context attributes on key conflict
- Revenue is passed as `"revenue"` event tag (Optimizely convention)
- Decision reasons stored as `[]string` in FlagMetadata under `"reasons"` key
- PROVIDER_STALE deferred — no upstream mechanism to detect connectivity failures
