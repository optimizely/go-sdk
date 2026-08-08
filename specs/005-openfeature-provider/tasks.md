# Tasks: OpenFeature Provider

**Input**: Design documents from `/specs/005-openfeature-provider/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: TDD is NON-NEGOTIABLE per Constitution Principle III. All tests are written FIRST (red), then implementation follows (green), then refactor.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Package initialization, dependency, and project structure

- [x] T001 Create package directory `pkg/openfeature/` and add Optimizely Apache 2.0 copyright header template
- [x] T002 Add `github.com/open-feature/go-sdk` v1.17.x dependency to `go.mod` via `go get`
- [x] T003 Create `pkg/openfeature/provider.go` with Provider struct, `providerConfig`, `ProviderOption` type, `NewProvider` and `NewProviderWithClient` constructors, and `WithClientOptions` option function. Include compile-time interface assertions for `FeatureProvider`, `StateHandler`, `EventHandler`, and `Tracker`

**Checkpoint**: Package compiles, constructors callable, interface assertions verified at build time

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core mapping utilities that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 **RED**: Write tests for context mapping in `pkg/openfeature/context_test.go` — table-driven tests covering: valid targeting key extraction, missing targeting key error, variableKey extraction and stripping, attribute passthrough, unsupported attribute type filtering (nested objects, slices dropped)
- [x] T005 **GREEN**: Implement context mapping in `pkg/openfeature/context.go` — function to convert `openfeature.FlattenedContext` to Optimizely userID + attributes, extracting and stripping `variableKey`
- [x] T006 [P] **RED**: Write tests for reason/error mapping in `pkg/openfeature/reason_test.go` — table-driven tests covering: all five reason code mappings (TARGETING_MATCH, DISABLED, DEFAULT, ERROR, UNKNOWN) and all six error code mappings (FLAG_NOT_FOUND, TARGETING_KEY_MISSING, TYPE_MISMATCH, PARSE_ERROR, PROVIDER_NOT_READY, GENERAL)
- [x] T007 [P] **GREEN**: Implement reason/error mapping in `pkg/openfeature/reason.go` — functions to map Optimizely decision state to OpenFeature Reason and to map error conditions to OpenFeature ResolutionError

**Checkpoint**: Foundation ready — context and reason mapping tested and working. User story implementation can begin.

---

## Phase 3: User Story 1 - Boolean Flag Evaluation (Priority: P1)

**Goal**: Evaluate Optimizely feature flags as boolean values through the OpenFeature interface

**Independent Test**: Register provider with OpenFeature, call boolean evaluation for a known flag key, verify result matches expected Optimizely decision

### Tests for User Story 1

> **Write these tests FIRST. Ensure they FAIL before implementation.**

- [x] T008 [P] [US1] **RED**: Write tests for Metadata and Hooks in `pkg/openfeature/provider_test.go` — Metadata returns name "Optimizely", Hooks returns empty slice
- [x] T009 [P] [US1] **RED**: Write tests for BooleanEvaluation in `pkg/openfeature/evaluation_test.go` — table-driven tests covering: user qualifies (returns true, TARGETING_MATCH reason, variation key as variant), user does not qualify (returns false), flag not found (returns default, FLAG_NOT_FOUND error), provider not ready (returns default, PROVIDER_NOT_READY error), missing targeting key (returns default, TARGETING_KEY_MISSING error)

### Implementation for User Story 1

- [x] T010 [P] [US1] **GREEN**: Implement `Metadata()` and `Hooks()` in `pkg/openfeature/provider.go`
- [x] T011 [US1] **GREEN**: Implement `BooleanEvaluation` in `pkg/openfeature/evaluation.go` — extract context via `context.go`, call `Decide` on Optimizely client, map `Enabled` field to bool result, map decision state to reason via `reason.go`, handle all error cases
- [x] T012 [US1] **REFACTOR**: Review and refactor US1 implementation — extract shared evaluation helper if pattern emerges, ensure godoc comments on all exported symbols

**Checkpoint**: Boolean flag evaluation works end-to-end. Provider can be registered with OpenFeature and used for boolean flags.

---

## Phase 4: User Story 2 - Typed Variable Evaluation (Priority: P2)

**Goal**: Retrieve Optimizely feature variable values through OpenFeature typed evaluation methods (string, int, float, object)

**Independent Test**: Configure flags with variables of each type, call corresponding OpenFeature evaluation, verify values match

### Tests for User Story 2

> **Write these tests FIRST. Ensure they FAIL before implementation.**

- [x] T013 [P] [US2] **RED**: Write tests for StringEvaluation in `pkg/openfeature/evaluation_test.go` — table-driven tests covering: valid string variable with variableKey, missing variableKey returns error, variable not found returns default, type mismatch returns default with PARSE_ERROR
- [x] T014 [P] [US2] **RED**: Write tests for IntEvaluation in `pkg/openfeature/evaluation_test.go` — table-driven tests covering: valid int variable, missing variableKey, parse error for non-integer value
- [x] T015 [P] [US2] **RED**: Write tests for FloatEvaluation in `pkg/openfeature/evaluation_test.go` — table-driven tests covering: valid float variable, missing variableKey, parse error
- [x] T016 [P] [US2] **RED**: Write tests for ObjectEvaluation in `pkg/openfeature/evaluation_test.go` — table-driven tests covering: specific variable via variableKey (parsed JSON), full variables map when variableKey omitted, flag with no variables returns default with error

### Implementation for User Story 2

- [x] T017 [US2] **GREEN**: Implement `StringEvaluation` in `pkg/openfeature/evaluation.go` — extract variableKey from context, call Decide, extract variable from Variables map, parse as string, handle errors
- [x] T018 [US2] **GREEN**: Implement `IntEvaluation` in `pkg/openfeature/evaluation.go` — same pattern, parse as int64
- [x] T019 [US2] **GREEN**: Implement `FloatEvaluation` in `pkg/openfeature/evaluation.go` — same pattern, parse as float64
- [x] T020 [US2] **GREEN**: Implement `ObjectEvaluation` in `pkg/openfeature/evaluation.go` — if variableKey present, extract and parse JSON variable; if absent, return full Variables.ToMap()
- [x] T021 [US2] **REFACTOR**: Refactor typed evaluations — extract common variable extraction logic into shared helper, reduce duplication across String/Int/Float implementations

**Checkpoint**: All five OpenFeature evaluation types work. Developer can evaluate boolean flags and retrieve typed variables.

---

## Phase 5: User Story 3 - Provider Lifecycle Management (Priority: P3)

**Goal**: Provider participates in OpenFeature lifecycle (init, shutdown, events)

**Independent Test**: Register provider, wait for ready event, verify evaluations work, shutdown, verify cleanup

### Tests for User Story 3

> **Write these tests FIRST. Ensure they FAIL before implementation.**

- [x] T022 [P] [US3] **RED**: Write tests for Init in `pkg/openfeature/lifecycle_test.go` — SDK-key mode: Init creates client and sets ready state; pre-initialized client mode: Init sets ready immediately; Init failure: returns error
- [x] T023 [P] [US3] **RED**: Write tests for Shutdown in `pkg/openfeature/lifecycle_test.go` — SDK-key mode: Shutdown calls Close on owned client, sets not-ready; pre-initialized mode: Shutdown sets not-ready but does NOT call Close on client
- [x] T024 [P] [US3] **RED**: Write tests for EventChannel in `pkg/openfeature/lifecycle_test.go` — returns a readable channel; emits ProviderReady on successful init; emits ProviderError on failed init; emits ProviderConfigChange on datafile update

### Implementation for User Story 3

- [x] T025 [US3] **GREEN**: Implement `Init` in `pkg/openfeature/lifecycle.go` — SDK-key mode: create OptimizelyClient via factory with stored options, register notification center listener for config updates, set ready state, emit ProviderReady event; pre-initialized mode: set ready, emit ProviderReady; on failure: emit ProviderError, return error
- [x] T026 [US3] **GREEN**: Implement `Shutdown` in `pkg/openfeature/lifecycle.go` — if ownsClient: call client.Close(); set not-ready; close event channel goroutine
- [x] T027 [US3] **GREEN**: Implement `EventChannel` in `pkg/openfeature/lifecycle.go` — return read-only event channel; background goroutine listens to notification center ProjectConfigUpdate and forwards as ProviderConfigChange events
- [x] T028 [US3] **REFACTOR**: Review lifecycle implementation — ensure goroutine cleanup is deterministic, verify no channel leaks, add godoc comments

**Checkpoint**: Full lifecycle works. Provider initializes, emits events, and shuts down cleanly in both construction modes.

---

## Phase 6: User Story 4 - Event Tracking (Priority: P4)

**Goal**: Track conversion events through OpenFeature tracking interface

**Independent Test**: Call track with event name and context, verify Optimizely event processor receives the conversion event

### Tests for User Story 4

> **Write these tests FIRST. Ensure they FAIL before implementation.**

- [x] T029 [US4] **RED**: Write tests for Track in `pkg/openfeature/tracking_test.go` — table-driven tests covering: valid track call dispatches to Optimizely Track with correct event key and user context, attributes passed as event tags, provider not ready drops call without panic

### Implementation for User Story 4

- [x] T030 [US4] **GREEN**: Implement `Track` in `pkg/openfeature/tracking.go` — extract targeting key and attributes from EvaluationContext, call Optimizely client.Track with event name and user context, handle not-ready state gracefully
- [x] T031 [US4] **REFACTOR**: Review tracking implementation — ensure error handling is consistent with evaluation methods, add godoc comments

**Checkpoint**: All four user stories complete. Full OpenFeature provider functionality operational.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validation, quality, and release readiness

- [x] T032 [P] Run `go test -race ./pkg/openfeature/...` to verify no data races under concurrent access
- [x] T033 [P] Run `make test` to verify ALL existing SDK tests still pass (zero modifications, SC-002)
- [x] T034 [P] Run `make lint` to verify linter passes on new package
- [x] T035 Parity validation: write a test that evaluates the same flag via native `Decide` API and via the OpenFeature provider and asserts identical results (SC-004) in `pkg/openfeature/evaluation_test.go`
- [x] T036 Review all new files for copyright headers, godoc comments, and idiomatic Go compliance (Constitution Principles VII, VIII)
- [x] T037 Run quickstart.md validation — verify the code examples compile and work

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - US1 (Phase 3): Can start after Foundational
  - US2 (Phase 4): Can start after Foundational (independent of US1, but US1 establishes evaluation patterns)
  - US3 (Phase 5): Can start after Foundational (independent of US1/US2)
  - US4 (Phase 6): Can start after Foundational (independent of US1/US2/US3)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Depends on Phase 2 only. Establishes the core evaluation pattern — recommended to complete first.
- **User Story 2 (P2)**: Depends on Phase 2. Benefits from US1's evaluation helper pattern but is independently implementable.
- **User Story 3 (P3)**: Depends on Phase 2. Independent of US1/US2 — lifecycle is orthogonal to evaluation.
- **User Story 4 (P4)**: Depends on Phase 2. Independent of all other stories.

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD Red phase)
- Implementation makes tests pass (TDD Green phase)
- Refactor follows green (TDD Refactor phase)
- Story complete before moving to next priority

### Parallel Opportunities

- T004 and T006 can run in parallel (different files: context_test.go vs reason_test.go)
- T008 and T009 can run in parallel (different test files)
- T013, T014, T015, T016 can all run in parallel (different test sections, no dependencies)
- T022, T023, T024 can all run in parallel (different test scenarios)
- T032, T033, T034 can all run in parallel (independent validation commands)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (context + reason mapping)
3. Complete Phase 3: User Story 1 (boolean evaluation)
4. **STOP and VALIDATE**: Register provider with OpenFeature, evaluate boolean flag
5. This alone delivers value — developers can use OpenFeature for feature flags

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Boolean evaluation works → MVP!
3. Add User Story 2 → Full typed evaluation → Feature-complete evaluation
4. Add User Story 3 → Lifecycle management → Production-ready provider
5. Add User Story 4 → Event tracking → Complete provider
6. Polish → Race testing, linting, parity validation → Release-ready

### Recommended Execution Order

Sequential by priority (P1 → P2 → P3 → P4) since a single developer will benefit from the evaluation patterns established in US1 carrying through to US2.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- **RED/GREEN/REFACTOR** labels indicate TDD phase (Constitution Principle III)
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
