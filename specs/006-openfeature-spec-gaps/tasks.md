# Tasks: OpenFeature Provider Spec Compliance Gaps

**Input**: Design documents from `/specs/006-openfeature-spec-gaps/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per constitution (Principle III: Test-Driven Development is NON-NEGOTIABLE). Tests MUST be written first and FAIL before implementation.

**Organization**: Tasks grouped by user story. User Story 5 (PROVIDER_STALE) is deferred per research.md R6.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: No setup needed — project structure and dependencies already exist. The `pkg/openfeature/` package is established. OpenFeature Go SDK v1.14.1 is already in go.mod.

No tasks in this phase.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Refactor the shared `emitEvent` helper to accept an error code parameter. This change is required by both US1 (error event error codes) and US3 (init error typing), so it must be done first.

- [x] T001 Write test asserting `emitEvent` populates `ProviderEventDetails.ErrorCode` when an error code is provided in pkg/openfeature/lifecycle_test.go
- [x] T002 Write test asserting `emitEvent` leaves `ErrorCode` empty when no error code is provided in pkg/openfeature/lifecycle_test.go
- [x] T003 Update `emitEvent` signature to accept `errorCode of.ErrorCode` and set it on `ProviderEventDetails` in pkg/openfeature/lifecycle.go
- [x] T004 Update all existing `emitEvent` call sites to pass the new parameter (empty string for non-error events, `of.ProviderFatalCode` for error events) in pkg/openfeature/lifecycle.go

**Checkpoint**: `emitEvent` now supports error codes. All existing tests still pass. New tests pass.

---

## Phase 3: User Story 1 - Error Event Error Codes (Priority: P1)

**Goal**: All `PROVIDER_ERROR` events emitted via `EventChannel()` carry a non-empty `ErrorCode` in `ProviderEventDetails`.

**Independent Test**: Trigger a provider init failure and verify the emitted error event has `ErrorCode == ProviderFatalCode`.

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T005 [US1] Write test asserting that when `Init()` fails, the `PROVIDER_ERROR` event emitted on `EventChannel()` has `ErrorCode == of.ProviderFatalCode` in pkg/openfeature/lifecycle_test.go
- [x] T006 [US1] Write test asserting that when `Init()` succeeds, the `PROVIDER_READY` event has empty `ErrorCode` in pkg/openfeature/lifecycle_test.go

### Implementation for User Story 1

- [x] T007 [US1] Verify `emitEvent` calls in `Init()` pass `of.ProviderFatalCode` on error path and empty string on success path in pkg/openfeature/lifecycle.go

**Checkpoint**: Error events carry error codes. `go test ./pkg/openfeature/... -run TestInit -v` passes.

---

## Phase 4: User Story 3 - Init Error Typing (Priority: P2)

**Goal**: `Init()` returns `*of.ProviderInitError` with `ErrorCode == ProviderFatalCode` on failure, enabling the OpenFeature SDK to set FATAL state.

**Independent Test**: Force init failure and assert the returned error is `*of.ProviderInitError` with correct `ErrorCode`.

### Tests for User Story 3

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T008 [US3] Write test asserting `Init()` failure returns `*of.ProviderInitError` (via `errors.As`) with `ErrorCode == of.ProviderFatalCode` in pkg/openfeature/lifecycle_test.go
- [x] T009 [US3] Write test asserting `Init()` success returns nil error (pre-initialized client mode) in pkg/openfeature/lifecycle_test.go

### Implementation for User Story 3

- [x] T010 [US3] Change `Init()` error return from `fmt.Errorf(...)` to `&of.ProviderInitError{ErrorCode: of.ProviderFatalCode, Message: err.Error()}` in pkg/openfeature/lifecycle.go

**Checkpoint**: Init errors are properly typed. `go test ./pkg/openfeature/... -run TestInit -v` passes.

---

## Phase 5: User Story 2 - Tracking Event Details Forwarding (Priority: P1)

**Goal**: `Track()` forwards `TrackingEventDetails.Value()` as `"revenue"` event tag and merges `TrackingEventDetails.Attributes()` into event tags, with tracking details overriding context attributes on key conflict.

**Independent Test**: Call `Track` with revenue and custom attributes, verify they arrive in the Optimizely `client.Track` call as event tags.

### Tests for User Story 2

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T011 [P] [US2] Write test asserting `Track` with a numeric value includes `"revenue"` in event tags passed to `client.Track` in pkg/openfeature/tracking_test.go
- [x] T012 [P] [US2] Write test asserting `Track` with custom attributes merges them into event tags in pkg/openfeature/tracking_test.go
- [x] T013 [P] [US2] Write test asserting tracking details override context attributes on key conflict in pkg/openfeature/tracking_test.go
- [x] T014 [P] [US2] Write test asserting `Track` with zero value (0.0) still includes `"revenue": 0.0` in event tags in pkg/openfeature/tracking_test.go
- [x] T015 [P] [US2] Write test asserting `Track` with empty details preserves backward-compatible behavior in pkg/openfeature/tracking_test.go

### Implementation for User Story 2

- [x] T016 [US2] Implement event tag merge logic in `Track()`: start with context attributes, add `"revenue"` from `details.Value()`, merge `details.Attributes()` (overrides context) in pkg/openfeature/tracking.go

**Checkpoint**: Tracking details flow through to Optimizely. `go test ./pkg/openfeature/... -run TestTrack -v` passes.

---

## Phase 6: User Story 4 - Flag Metadata Population (Priority: P3)

**Goal**: Successful flag evaluations include `FlagMetadata` with `flagKey`, `ruleKey` (if present), and `reasons` (if present) as a string slice.

**Independent Test**: Evaluate a known flag and assert `FlagMetadata` contains expected keys.

### Tests for User Story 4

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T017 [P] [US4] Write test asserting `BooleanEvaluation` result has `FlagMetadata["flagKey"]` populated in pkg/openfeature/evaluation_test.go
- [x] T018 [P] [US4] Write test asserting `FlagMetadata["ruleKey"]` is populated when a rule matches in pkg/openfeature/evaluation_test.go
- [x] T019 [P] [US4] Write test asserting `FlagMetadata["reasons"]` is a `[]string` when reasons are present in pkg/openfeature/evaluation_test.go
- [x] T020 [P] [US4] Write test asserting `FlagMetadata` is nil/empty on error paths (flag not found, provider not ready) in pkg/openfeature/evaluation_test.go

### Implementation for User Story 4

- [x] T021 [US4] Add `RuleKey string` and `Reasons []string` fields to `decisionResult` struct in pkg/openfeature/evaluation.go
- [x] T022 [US4] Populate `RuleKey` and `Reasons` from `OptimizelyDecision` in `evaluateWithContext` in pkg/openfeature/evaluation.go
- [x] T023 [US4] Update `toProviderDetail()` to build `of.FlagMetadata` map with `flagKey`, `ruleKey` (if non-empty), and `reasons` (if non-empty) in pkg/openfeature/evaluation.go

**Checkpoint**: FlagMetadata populated on successful evaluations. `go test ./pkg/openfeature/... -run TestObjectEvaluation -v` and `TestBooleanEvaluation` pass.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories.

- [x] T024 Run full test suite: `make test` — all tests pass with no regressions
- [x] T025 Run linter: `make lint` — no new lint violations (pre-existing copyloopvar issue unrelated)
- [x] T026 Run race detector: `go test -race ./pkg/openfeature/...` — no data races
- [x] T027 Verify copyright headers updated to 2026 on all modified files in pkg/openfeature/

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No tasks — already done
- **Foundational (Phase 2)**: No dependencies — can start immediately. BLOCKS Phases 3-6.
- **US1 Error Codes (Phase 3)**: Depends on Foundational (T001-T004)
- **US3 Init Error Typing (Phase 4)**: Depends on Foundational (T001-T004). Can run in parallel with Phase 3.
- **US2 Tracking Details (Phase 5)**: Depends on Foundational only. Can run in parallel with Phases 3-4.
- **US4 Flag Metadata (Phase 6)**: Depends on Foundational only. Can run in parallel with Phases 3-5.
- **Polish (Phase 7)**: Depends on all user stories being complete.

### User Story Dependencies

- **US1 (Error Event Codes)**: Depends on Foundational. No cross-story dependencies.
- **US3 (Init Error Typing)**: Depends on Foundational. Shares `Init()` code with US1 but modifies different aspects (error return type vs event emission).
- **US2 (Tracking Details)**: Depends on Foundational only (for consistency). Modifies `tracking.go` — fully independent of other stories.
- **US4 (Flag Metadata)**: Depends on Foundational only. Modifies `evaluation.go` — fully independent of other stories.

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD per constitution)
- Implementation follows tests
- Story complete when all tests pass

### Parallel Opportunities

- T011-T015 (US2 tests) can all run in parallel — they test different behaviors
- T017-T020 (US4 tests) can all run in parallel — they test different behaviors
- Phases 3, 4, 5, 6 can all run in parallel after Phase 2 completes (different files)

---

## Parallel Example: User Story 2 (Tracking)

```
# Launch all tracking tests together:
Task: "Write test asserting Track with numeric value in pkg/openfeature/tracking_test.go"
Task: "Write test asserting Track with custom attributes in pkg/openfeature/tracking_test.go"
Task: "Write test asserting tracking details override context attributes in pkg/openfeature/tracking_test.go"
Task: "Write test asserting Track with zero value in pkg/openfeature/tracking_test.go"
Task: "Write test asserting Track with empty details in pkg/openfeature/tracking_test.go"

# Then implement:
Task: "Implement event tag merge logic in Track() in pkg/openfeature/tracking.go"
```

---

## Implementation Strategy

### MVP First (US1 + US3: Error Handling)

1. Complete Phase 2: Foundational (`emitEvent` refactor)
2. Complete Phase 3: US1 (Error event codes)
3. Complete Phase 4: US3 (Init error typing)
4. **STOP and VALIDATE**: Error events carry codes, Init returns typed errors
5. The OpenFeature SDK state machine now works correctly

### Incremental Delivery

1. Foundational → `emitEvent` supports error codes
2. US1 + US3 → Error handling spec-compliant → Validate
3. US2 → Tracking details flow through → Validate
4. US4 → FlagMetadata populated → Validate
5. Polish → Full suite green, no regressions

### Sequential Strategy (Single Developer)

Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7

Each phase is a natural commit point.

---

## Notes

- User Story 5 (PROVIDER_STALE) is deferred — no tasks generated. See research.md R6.
- [P] tasks = different files or different test functions, no dependencies
- [Story] label maps task to specific user story for traceability
- Constitution requires TDD: all test tasks must be completed and fail before implementation tasks
- Commit after each phase checkpoint
