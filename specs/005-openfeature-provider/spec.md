# Feature Specification: OpenFeature Provider

**Feature Branch**: `005-openfeature-provider`
**Created**: 2026-04-10
**Status**: Draft
**Input**: User description: "Add OpenFeature specification compatibility to this SDK. Purely additive, no breaking changes. Implement only the OpenFeature spec for which equivalent behavior already exists in the Optimizely SDK."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Boolean Flag Evaluation via OpenFeature (Priority: P1)

A developer who has standardized on OpenFeature as their feature flag abstraction layer wants to evaluate Optimizely feature flags using the OpenFeature client. They register the Optimizely provider with OpenFeature, then call boolean evaluation to check whether a feature is enabled for a given user. The provider translates the OpenFeature evaluation context (targeting key and attributes) into an Optimizely user context and returns the feature's enabled/disabled state as a boolean resolution.

**Why this priority**: Boolean flag evaluation is the most common OpenFeature operation and maps directly to the Optimizely SDK's core `Decide` functionality. Without this, no other OpenFeature integration is useful.

**Independent Test**: Can be fully tested by registering the provider with OpenFeature, calling boolean evaluation for a known flag key with a targeting key and attributes, and verifying the returned value matches the expected Optimizely decision.

**Acceptance Scenarios**:

1. **Given** the Optimizely provider is registered with OpenFeature and the SDK is initialized with a valid datafile, **When** a developer evaluates a boolean flag for a user who qualifies for the feature, **Then** the evaluation returns `true` with a reason indicating the match and a variant identifying the variation.
2. **Given** the Optimizely provider is registered and initialized, **When** a developer evaluates a boolean flag for a user who does not qualify, **Then** the evaluation returns `false` with an appropriate reason.
3. **Given** the Optimizely provider is registered and initialized, **When** a developer evaluates a flag key that does not exist, **Then** the evaluation returns the caller-supplied default value and a "flag not found" error indication.
4. **Given** the Optimizely provider is registered but the SDK is not yet ready, **When** a developer evaluates any flag, **Then** the evaluation returns the default value with a "provider not ready" error indication.
5. **Given** the evaluation context has no targeting key, **When** a developer evaluates a boolean flag, **Then** the evaluation returns the default value with a "targeting key missing" error indication.

---

### User Story 2 - Typed Variable Evaluation via OpenFeature (Priority: P2)

A developer wants to retrieve Optimizely feature variable values through the OpenFeature typed evaluation methods (string, integer, float, and object). The provider maps each OpenFeature evaluation type to the corresponding Optimizely feature variable type, returning the variable value along with resolution details.

**Why this priority**: Typed variable evaluation extends the core boolean evaluation to cover the full range of Optimizely feature variables, enabling developers to use OpenFeature for all flag-driven configuration without falling back to the native SDK.

**Independent Test**: Can be tested by configuring feature flags with variables of each type (string, integer, double, JSON) in the datafile, then calling the corresponding OpenFeature evaluation method and verifying the returned value and type match.

**Acceptance Scenarios**:

1. **Given** a feature flag with a string variable is configured, **When** a developer calls string evaluation with the flag key, **Then** the provider returns the variable's string value for the matched variation.
2. **Given** a feature flag with an integer variable is configured, **When** a developer calls integer evaluation with the flag key, **Then** the provider returns the variable's integer value.
3. **Given** a feature flag with a double variable is configured, **When** a developer calls float evaluation with the flag key, **Then** the provider returns the variable's float value.
4. **Given** a feature flag with a JSON variable is configured, **When** a developer calls object evaluation with the flag key, **Then** the provider returns the variable's parsed structure.
5. **Given** a feature flag exists but the requested variable type does not match the configured type, **When** a developer calls a typed evaluation, **Then** the evaluation returns the default value with a "type mismatch" or "parse error" indication.
6. **Given** a feature flag with multiple variables is configured, **When** a developer calls a scalar typed evaluation without a `variableKey` in the evaluation context, **Then** the evaluation returns the default value with an error indicating the variable key is required.
7. **Given** a feature flag with multiple variables is configured, **When** a developer calls object evaluation without a `variableKey`, **Then** the evaluation returns the full variables map as the resolved object.

---

### User Story 3 - Provider Lifecycle Management (Priority: P3)

A developer wants the Optimizely provider to participate in the OpenFeature provider lifecycle so that the underlying Optimizely client initializes when the provider is registered and shuts down cleanly when the provider is removed or the application exits. The provider emits lifecycle events (ready, error, stale, configuration changed) so that OpenFeature consumers can react to provider state transitions.

**Why this priority**: Lifecycle management ensures the provider behaves correctly in applications that rely on OpenFeature's provider state model for readiness checks and graceful shutdown. It is important but secondary to the core evaluation functionality.

**Independent Test**: Can be tested by registering the provider, waiting for the ready event, verifying evaluations succeed, then shutting down and verifying the provider transitions to not-ready state and the underlying client resources are released.

**Acceptance Scenarios**:

1. **Given** a developer registers the Optimizely provider with OpenFeature, **When** the underlying Optimizely client finishes initialization, **Then** the provider emits a "ready" event and the provider state transitions to ready.
2. **Given** the provider is in the ready state, **When** the underlying Optimizely client receives a datafile update, **Then** the provider emits a "configuration changed" event.
3. **Given** the provider was created with an SDK key and is in the ready state, **When** the application requests provider shutdown, **Then** the provider calls `Close()` on the underlying Optimizely client it created and transitions to not-ready.
4. **Given** the provider was created with a pre-initialized client and is in the ready state, **When** the application requests provider shutdown, **Then** the provider transitions to not-ready but does NOT call `Close()` on the client (caller retains lifecycle ownership).
5. **Given** the Optimizely client fails to initialize (e.g., invalid SDK key), **When** the provider initialization completes, **Then** the provider emits an "error" event and the provider state reflects the error.

---

### User Story 4 - Event Tracking via OpenFeature (Priority: P4)

A developer wants to track conversion events through the OpenFeature tracking interface. The provider translates OpenFeature tracking calls into Optimizely `Track` calls, mapping the OpenFeature evaluation context to an Optimizely user context and the tracking event name to an Optimizely event key.

**Why this priority**: Event tracking completes the feature experimentation loop — flags control the experience and tracking measures the outcome. However, many users may continue to use the native Optimizely `Track` method directly, so this is a lower priority convenience.

**Independent Test**: Can be tested by calling the OpenFeature track method with an event name and evaluation context, then verifying the Optimizely event processor receives the corresponding conversion event.

**Acceptance Scenarios**:

1. **Given** the Optimizely provider is registered and ready, **When** a developer calls track with an event name and evaluation context, **Then** the provider dispatches a conversion event to the Optimizely event processor with the correct event key and user context.
2. **Given** the evaluation context includes additional attributes, **When** a developer calls track, **Then** those attributes are passed as event tags to the Optimizely `Track` call.
3. **Given** the provider is not ready, **When** a developer calls track, **Then** the call is silently dropped or returns an error without crashing.

---

### Edge Cases

- What happens when the evaluation context contains attribute types not supported by Optimizely (e.g., nested objects, arrays)? The provider passes only supported attribute types and ignores unsupported ones.
- What happens when multiple OpenFeature providers are registered and the Optimizely provider is not the active one? The provider must not consume resources or process events when not active.
- What happens when the Optimizely client's datafile is updated while evaluations are in progress? Evaluations in flight use the config snapshot they started with; subsequent evaluations use the new config. This is existing SDK behavior.
- What happens when `Decide` returns a decision with `Enabled: false` but a valid variation key? Boolean evaluation returns `false`; typed variable evaluations return values from the matched variation regardless of enabled state, consistent with the Optimizely `Decide` behavior.
- What happens when the flag key passed to a typed evaluation (e.g., string) corresponds to a flag that has no variables? The evaluation returns the default value with an appropriate error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The provider MUST implement the OpenFeature provider interface, supporting boolean, string, integer, float, and object evaluation methods.
- **FR-002**: The provider MUST map the OpenFeature evaluation context's targeting key to the Optimizely user ID, and all other context attributes to Optimizely user attributes.
- **FR-003**: Boolean evaluation MUST return the `Enabled` field from the Optimizely `Decide` result for the given flag key.
- **FR-004**: String, integer, float, and object evaluations MUST return the corresponding typed variable value from the Optimizely `Decide` result's `Variables` for the given flag key. The caller specifies which variable to extract via a `variableKey` attribute in the evaluation context. If `variableKey` is omitted, object evaluation MUST return the full variables map; scalar typed evaluations MUST return an error.
- **FR-005**: The provider MUST return the Optimizely variation key as the OpenFeature `Variant` in resolution details.
- **FR-006**: The provider MUST map Optimizely decision outcomes to standard OpenFeature reason codes: `TARGETING_MATCH`, `SPLIT`, `DEFAULT`, `DISABLED`, `ERROR`, `UNKNOWN`. Raw Optimizely reason strings are not included in resolution details.
- **FR-007**: The provider MUST return appropriate OpenFeature error codes when evaluation fails: flag not found, targeting key missing, type mismatch, provider not ready.
- **FR-008**: The provider MUST implement the OpenFeature provider lifecycle (initialize and shutdown), delegating to the Optimizely client's creation and `Close()` methods.
- **FR-009**: The provider MUST emit OpenFeature provider events for state transitions: ready, error, configuration changed.
- **FR-010**: The provider MUST implement the OpenFeature tracking interface, delegating to the Optimizely `Track` method.
- **FR-011**: The provider MUST be purely additive — no changes to existing Optimizely SDK types, interfaces, or behavior. All existing public APIs MUST continue to function identically.
- **FR-012**: The provider MUST return provider metadata including a descriptive name (e.g., "Optimizely").
- **FR-013**: The provider MUST support two construction modes: (a) accept an SDK key and optional Optimizely factory options, creating and owning the underlying client, or (b) accept a pre-initialized `OptimizelyClient`, wrapping it without managing its lifecycle. When a pre-initialized client is provided, the provider MUST NOT call `Close()` on it during shutdown — the caller retains lifecycle ownership.
- **FR-014**: The provider MUST be safe for concurrent use from multiple goroutines.

### Key Entities

- **Provider**: The adapter that bridges the OpenFeature provider interface to the Optimizely SDK. Holds a reference to an Optimizely client and translates between OpenFeature and Optimizely data models.
- **Evaluation Context**: The OpenFeature representation of a user, carrying a targeting key (mapped to Optimizely user ID) and arbitrary attributes (mapped to Optimizely user attributes).
- **Resolution Detail**: The OpenFeature response type containing the resolved value, variant (Optimizely variation key), reason code, and any error information.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can evaluate all five OpenFeature value types (boolean, string, integer, float, object) against Optimizely feature flags without using any Optimizely-specific API.
- **SC-002**: All existing Optimizely SDK tests continue to pass with zero modifications, confirming no breaking changes.
- **SC-003**: The provider passes the OpenFeature provider conformance expectations: correct lifecycle transitions, proper error codes, and accurate resolution details.
- **SC-004**: Decision results (enabled state, variable values, variation keys) returned through the OpenFeature provider are identical to results returned through the native Optimizely `Decide` API for the same inputs.
- **SC-005**: The provider initializes and becomes ready within the same time bounds as a native Optimizely client initialization.
- **SC-006**: Adding the provider as a dependency does not increase the compiled binary size of applications that do not import it (i.e., it lives in its own package with no init-time side effects).

## Clarifications

### Session 2026-04-10

- Q: How does the caller specify which variable to extract for typed evaluations when a flag has multiple variables? → A: Caller passes a `variableKey` attribute in the evaluation context. If omitted, object evaluation returns the full variables map; scalar types return an error.
- Q: Where does the provider package live — inside the existing module or as a separate Go module? → A: Inside the existing module as a new package (e.g., `pkg/openfeature/`). OpenFeature Go SDK added as a module dependency.
- Q: Should the provider accept only an SDK key, or also a pre-initialized OptimizelyClient? → A: Both. SDK key (provider creates and owns client) or pre-initialized client (provider wraps it; caller manages lifecycle).
- Q: How granular should the Optimizely-to-OpenFeature reason code mapping be? → A: Map to standard OpenFeature reason codes only (TARGETING_MATCH, SPLIT, DEFAULT, DISABLED, ERROR, UNKNOWN). Raw Optimizely reasons not included in resolution details.
- Q: Which version of the OpenFeature Go SDK should the provider target? → A: Latest stable v1.x release.

## Assumptions

- The provider targets the latest stable v1.x release of the OpenFeature Go SDK (`github.com/open-feature/go-sdk`).
- The target users are Go developers who have adopted OpenFeature as a vendor-neutral feature flag abstraction and want to use Optimizely as their backend provider.
- The OpenFeature Go SDK (`github.com/open-feature/go-sdk`) is added as a dependency to the existing module's `go.mod`. The provider lives in a dedicated package within the module (e.g., `pkg/openfeature/`). Applications that do not import this package will not link the OpenFeature code, though the dependency will appear in the module graph.
- The provider uses the Optimizely `Decide` API (not the legacy `IsFeatureEnabled` / `GetFeatureVariable*` APIs) as the underlying evaluation mechanism, since `Decide` returns all needed information (enabled state, variables, variation key, reasons) in a single call.
- For typed evaluations (string, integer, float, object), the provider evaluates the flag via `Decide` and extracts the requested variable from the decision's `Variables`. The flag key serves as the Optimizely feature flag key. The caller specifies which variable to extract by passing a variable key in the evaluation context (e.g., `"variableKey": "price"`). If the variable key is omitted, object evaluation returns the full variables map; scalar typed evaluations (string, integer, float) return an error indicating the variable key is required.
- OpenFeature hooks at the provider level are out of scope for this initial implementation. The provider returns an empty hooks list.
- OpenFeature features that have no Optimizely equivalent (e.g., domain-scoped providers, named clients) are out of scope.
- This work targets a MINOR version release of the Optimizely Go SDK, as it is purely additive.
