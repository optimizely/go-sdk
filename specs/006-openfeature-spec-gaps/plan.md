# Implementation Plan: OpenFeature Provider Spec Compliance Gaps

**Branch**: `006-openfeature-spec-gaps` | **Date**: 2026-04-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-openfeature-spec-gaps/spec.md`

## Summary

Close 5 OpenFeature specification compliance gaps in the Optimizely provider (`pkg/openfeature/`): add error codes to error events (5.1.5), return typed `ProviderInitError` from Init (1.7.5), forward tracking event details to Optimizely (6.2.1, 6.2.2), populate FlagMetadata on resolution details (2.2.9), and defer PROVIDER_STALE signaling (5.1.1 MAY — no upstream mechanism available). All changes are internal to `pkg/openfeature/` with no public API surface modifications.

## Technical Context

**Language/Version**: Go 1.21+ (minimum), tested on Go 1.24 (CI)  
**Primary Dependencies**: `github.com/open-feature/go-sdk` v1.14.1 (Apache 2.0), `github.com/optimizely/go-sdk/v2` (this repo)  
**Storage**: N/A  
**Testing**: `go test`, `make test`, `make lint` (golangci-lint v1.64.2)  
**Target Platform**: Cross-platform library (any OS Go supports)  
**Project Type**: Library (Go SDK)  
**Performance Goals**: No performance regression — all changes are on evaluation return paths (struct field population), not hot decision paths  
**Constraints**: Must compile on Go 1.21. Must not break existing tests. Must pass `make lint` and `make test`.  
**Scale/Scope**: 4 files modified, ~60 lines of implementation, ~120 lines of tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Backward Compatibility | PASS | No public API changes. All modifications are internal behavior improvements. |
| II. Correctness & Determinism | PASS | No decision logic changes. Only touches event/error reporting and metadata. |
| III. Test-Driven Development | PASS | TDD workflow: write tests first (red), then implement (green), then refactor. |
| IV. Performance & Resource Efficiency | PASS | No new goroutines, channels, or unbounded structures. Small struct field additions. |
| V. Cross-SDK Parity | PASS | Aligns with OpenFeature specification — not Optimizely cross-SDK parity. No conflict. |
| VI. Security & Data Stewardship | PASS | No new logging of sensitive data. Error messages contain operation context, not user data. |
| VII. Observability & Debuggability | PASS | Improves observability — FlagMetadata and error codes add structured debugging signals. |
| VIII. Idiomatic Go | PASS | Uses standard error wrapping patterns, typed errors, map construction. |
| Enterprise Constraints | PASS | Go 1.21 minimum preserved. No new dependencies. Apache 2.0 compatible. |

**Post-Phase 1 re-check**: All gates remain PASS. No complexity violations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/006-openfeature-spec-gaps/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 research findings
├── data-model.md        # Phase 1 data model changes
├── quickstart.md        # Phase 1 quickstart guide
├── contracts/
│   └── provider-contract.md  # Phase 1 behavioral contract
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
pkg/openfeature/
├── provider.go          # No changes needed
├── lifecycle.go         # Modified: Init error typing, emitEvent ErrorCode
├── lifecycle_test.go    # Modified: Init error type assertions
├── tracking.go          # Modified: Event tag merge with tracking details
├── tracking_test.go     # Modified: Revenue/attribute forwarding tests
├── evaluation.go        # Modified: decisionResult fields, FlagMetadata population
├── evaluation_test.go   # Modified: FlagMetadata assertions
├── context.go           # No changes needed
├── context_test.go      # No changes needed
├── reason.go            # No changes needed
└── reason_test.go       # No changes needed
```

**Structure Decision**: All changes are contained within the existing `pkg/openfeature/` package. No new files or packages are needed.

## Implementation Phases

### Phase 1: Error Event Error Codes + Init Error Typing (P1+P2)

**Files**: `lifecycle.go`, `lifecycle_test.go`

**Changes**:
1. Modify `emitEvent` to accept an `errorCode of.ErrorCode` parameter and set it on `ProviderEventDetails`
2. Update all `emitEvent` call sites: pass `of.ProviderFatalCode` for error events, empty string for non-error events
3. Modify `Init()` to return `&of.ProviderInitError{ErrorCode: of.ProviderFatalCode, Message: err.Error()}` instead of `fmt.Errorf`
4. Update `Init()` error event emission to include `ProviderFatalCode`

**Tests (write first)**:
- Assert `Init()` failure returns `*of.ProviderInitError` with `ErrorCode == of.ProviderFatalCode`
- Assert error events emitted via `EventChannel()` carry `ErrorCode == of.ProviderFatalCode`
- Assert non-error events (ProviderReady, ProviderConfigChange) carry empty `ErrorCode`

### Phase 2: Tracking Event Details Forwarding (P1)

**Files**: `tracking.go`, `tracking_test.go`

**Changes**:
1. Build event tags map: start with context attributes as base
2. Set `"revenue"` key from `details.Value()`
3. Merge `details.Attributes()` over context attributes (tracking details win on conflict)
4. Pass merged map to `client.Track`

**Tests (write first)**:
- Track with numeric value: assert `"revenue"` key present in event tags
- Track with custom attributes: assert attributes merged into event tags
- Track with conflicting keys: assert tracking details override context attributes
- Track with zero value (0.0): assert `"revenue"` key present with 0.0
- Track with empty details: assert backward-compatible behavior (context attributes only plus revenue=0)

### Phase 3: FlagMetadata Population (P3)

**Files**: `evaluation.go`, `evaluation_test.go`

**Changes**:
1. Add `RuleKey string` and `Reasons []string` fields to `decisionResult`
2. Populate new fields from `OptimizelyDecision` in `evaluateWithContext`
3. Update `toProviderDetail()` to build `FlagMetadata` map with `flagKey`, `ruleKey` (if non-empty), and `reasons` (if non-empty)

**Tests (write first)**:
- Boolean evaluation: assert `FlagMetadata["flagKey"]` populated
- Evaluation with matched rule: assert `FlagMetadata["ruleKey"]` populated
- Evaluation with reasons: assert `FlagMetadata["reasons"]` is `[]string`
- Evaluation with no rule match: assert `ruleKey` absent from FlagMetadata
- Error evaluation: assert FlagMetadata is nil/empty

## Complexity Tracking

No constitution violations. No complexity justifications needed.
