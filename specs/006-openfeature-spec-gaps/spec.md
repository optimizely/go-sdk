# Feature Specification: OpenFeature Provider Spec Compliance Gaps

**Feature Branch**: `006-openfeature-spec-gaps`  
**Created**: 2026-04-11  
**Status**: Draft  
**Input**: User description: "Address the remaining gaps in our implementation of the spec"

## Clarifications

### Session 2026-04-11

- Q: Should init failures be classified as FATAL or GENERAL (transient)? → A: All init failures are FATAL. The provider cannot function without a client, and the OpenFeature SDK handles retry/re-registration at a higher level.
- Q: When tracking event tag keys conflict between context attributes and tracking details, which wins? → A: Tracking details override context attributes (more specific to the event).
- Q: How should decision reasons be represented in FlagMetadata? → A: Single key `"reasons"` with value as a string slice (`[]string{...}`), preserving the original structure.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Error Event Error Codes (Priority: P1)

An OpenFeature consumer registers a handler for `PROVIDER_ERROR` events to understand why the provider failed. When the Optimizely provider encounters an initialization failure, the consumer expects the error event to carry the `PROVIDER_FATAL` error code so the OpenFeature SDK transitions the provider to the `FATAL` state. All init failures are fatal since the provider cannot function without a successfully created client.

**Why this priority**: Without error codes on error events, the OpenFeature SDK cannot distinguish FATAL from ERROR state. This breaks the spec's state machine (requirements 1.7.4, 1.7.5, 5.1.5) and prevents consumers from implementing correct error-handling strategies.

**Independent Test**: Can be tested by triggering provider error events and asserting the `ErrorCode` field is populated in the emitted `ProviderEventDetails`.

**Acceptance Scenarios**:

1. **Given** a provider error event is emitted during initialization, **When** the initialization fails for any reason, **Then** the event's `ProviderEventDetails.ErrorCode` is set to `PROVIDER_FATAL` (all init failures are fatal since the provider cannot function without a client).
2. **Given** the provider `Init()` method fails, **When** the error is returned to the OpenFeature SDK, **Then** the error is a `*ProviderInitError` with `ErrorCode` set to `PROVIDER_FATAL` so the SDK transitions the provider to FATAL state.

---

### User Story 2 - Tracking Event Details Forwarding (Priority: P1)

An OpenFeature consumer calls `client.Track("purchase", evalCtx, trackingDetails)` with a numeric value (e.g., revenue of $49.99) and custom attributes (e.g., `{"currency": "USD", "item_count": 3}`). The consumer expects these tracking details to be forwarded to the Optimizely SDK so they appear in Optimizely's event reporting as event tags.

**Why this priority**: Tracking details (revenue, custom fields) are core to conversion tracking and experimentation analytics. Silently dropping them renders the Track implementation incomplete for real-world use (requirements 6.2.1, 6.2.2).

**Independent Test**: Can be tested by calling `Track` with tracking details containing a numeric value and custom attributes, then verifying the Optimizely `client.Track` receives them as event tags.

**Acceptance Scenarios**:

1. **Given** a consumer calls Track with a numeric value, **When** the provider forwards to Optimizely, **Then** the numeric value appears in the event tags (e.g., as `"revenue"`).
2. **Given** a consumer calls Track with custom attributes on the tracking details, **When** the provider forwards to Optimizely, **Then** the custom attributes are merged into the event tags sent to Optimizely.
3. **Given** a consumer calls Track with no tracking details value (zero) and no custom attributes, **When** the provider forwards to Optimizely, **Then** only user-context attributes are passed as event tags (backward-compatible behavior).

---

### User Story 3 - Init Error Typing (Priority: P2)

An OpenFeature consumer registers the Optimizely provider with the OpenFeature SDK. If initialization fails, the OpenFeature SDK inspects the returned error to determine whether the provider should be in `ERROR` state (recoverable) or `FATAL` state (permanent). The consumer expects the provider to return properly typed errors so the SDK's state machine works correctly.

**Why this priority**: The OpenFeature Go SDK specifically checks for `*ProviderInitError` with `ErrorCode == ProviderFatalCode` to set FATAL state. Without this, all init failures are treated as recoverable ERROR, which may cause the SDK to retry initialization unnecessarily or mislead consumers about provider health.

**Independent Test**: Can be tested by forcing init failures and asserting the returned error is a `*ProviderInitError` with the correct `ErrorCode`.

**Acceptance Scenarios**:

1. **Given** the provider is initialized and client creation fails for any reason, **When** `Init()` returns an error, **Then** the error is a `*ProviderInitError` with `ErrorCode` set to `PROVIDER_FATAL`.
2. **Given** the provider is initialized with a pre-initialized client (NewProviderWithClient), **When** `Init()` succeeds, **Then** no error is returned and the provider transitions to READY state (no init failure path exists for pre-initialized clients).

---

### User Story 4 - Flag Metadata Population (Priority: P3)

An OpenFeature consumer evaluates a flag and inspects the resolution details for metadata about how the decision was made. The consumer expects to see contextual information such as the flag key, rule key, and decision reasons in the `FlagMetadata` field of the resolution detail.

**Why this priority**: Flag metadata is a SHOULD requirement (2.2.9) that provides useful debugging and observability information. While not strictly required, it enriches the OpenFeature experience and aids in debugging flag evaluation behavior.

**Independent Test**: Can be tested by evaluating a known flag and asserting that `FlagMetadata` contains expected keys (e.g., `flagKey`, `ruleKey`).

**Acceptance Scenarios**:

1. **Given** a flag is successfully evaluated, **When** the resolution detail is returned, **Then** `FlagMetadata` contains the `flagKey` from the Optimizely decision.
2. **Given** a flag is successfully evaluated with a matched variation, **When** the resolution detail is returned, **Then** `FlagMetadata` contains the `ruleKey` from the Optimizely decision (if available).
3. **Given** a flag is successfully evaluated with decision reasons enabled, **When** the resolution detail is returned, **Then** `FlagMetadata` contains the decision `reasons` as a string slice preserving the original Optimizely reason list.

---

### User Story 5 - Stale Provider State Signaling (Priority: P3)

An OpenFeature consumer monitors provider state and wants to know when the provider's configuration data may be outdated (e.g., the polling config manager has lost connectivity to the Optimizely CDN). The consumer expects a `PROVIDER_STALE` event so they can display warnings or fall back to cached decisions with appropriate caution.

**Why this priority**: Stale signaling is a MAY requirement (5.1.1) and depends on the Optimizely SDK's notification system exposing connectivity status. This is the lowest priority because the Optimizely SDK may not currently surface granular connectivity events, making implementation dependent on upstream capabilities.

**Independent Test**: Can be tested by simulating a config update failure and verifying that a `PROVIDER_STALE` event is emitted.

**Acceptance Scenarios**:

1. **Given** the provider has been initialized and is in READY state, **When** the underlying config manager signals that it cannot reach the CDN, **Then** the provider emits a `PROVIDER_STALE` event.
2. **Given** the provider is in STALE state, **When** the config manager successfully fetches a new datafile, **Then** the provider emits a `PROVIDER_CONFIGURATION_CHANGED` event (returning to READY state).

---

### Edge Cases

- What happens when `Track` is called with a `TrackingEventDetails` value of 0.0? The value should still be forwarded (0.0 is a valid revenue/metric value, distinct from "no value provided").
- What happens when `Track` custom attributes contain unsupported types (maps, slices)? They should be silently filtered, consistent with context attribute handling.
- What happens when `Init()` fails and both an error event and error return occur? Both should carry consistent error codes.
- What happens when `FlagMetadata` is populated but the decision has no variation (flag not found)? Metadata should not be populated in error paths.
- What happens when context attributes and tracking details share the same key? Tracking details take precedence (more specific to the event).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The provider MUST populate `ProviderEventDetails.ErrorCode` on all `PROVIDER_ERROR` events emitted via `EventChannel()`.
- **FR-002**: The provider MUST return `*ProviderInitError` (not a plain `error`) from `Init()` when initialization fails, with `ErrorCode` set to `PROVIDER_FATAL` (all init failures are fatal since the provider cannot function without a client).
- **FR-003**: The provider MUST forward the numeric value from `TrackingEventDetails` to the Optimizely `client.Track` call as an event tag.
- **FR-004**: The provider MUST forward custom attributes from `TrackingEventDetails` to the Optimizely `client.Track` call, merged into event tags. When keys conflict between context attributes and tracking details, tracking details take precedence.
- **FR-005**: The provider MUST NOT break existing behavior when `TrackingEventDetails` has no value or custom attributes (zero value / empty map).
- **FR-006**: The provider SHOULD populate `FlagMetadata` on successful resolution details with available decision context (flag key, rule key, reasons).
- **FR-007**: The provider SHOULD emit `PROVIDER_STALE` events when the underlying configuration source becomes unreachable or outdated, if the Optimizely SDK provides such notifications.
- **FR-008**: The provider MUST NOT emit `PROVIDER_STALE` if the Optimizely SDK does not surface connectivity/staleness signals (graceful degradation).

### Key Entities

- **ProviderEventDetails**: The event payload structure carrying `Message`, `ErrorCode`, `FlagChanges`, and `EventMetadata`. Currently only `Message` is populated.
- **ProviderInitError**: The typed error structure the OpenFeature Go SDK inspects to determine FATAL vs ERROR provider state. Contains `ErrorCode` and `Message` fields.
- **TrackingEventDetails**: The OpenFeature structure carrying a numeric `Value()` and custom attributes for conversion tracking events.
- **FlagMetadata**: A `map[string]interface{}` on `ProviderResolutionDetail` carrying provider-specific decision context (flag key, rule key, decision reasons).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All `PROVIDER_ERROR` events emitted by the provider carry a non-empty `ErrorCode` in `ProviderEventDetails`.
- **SC-002**: The provider passes the OpenFeature Go SDK's state machine requirements: all init failures result in FATAL state via properly typed `*ProviderInitError`.
- **SC-003**: Tracking calls with a numeric value and custom attributes result in those values arriving in Optimizely event tags with no data loss.
- **SC-004**: Successful flag evaluations include `FlagMetadata` with at least the `flagKey` field populated.
- **SC-005**: No existing tests or behaviors are broken by these changes (full backward compatibility).
- **SC-006**: All new behaviors have accompanying unit tests with coverage for both success and error paths.

## Assumptions

- The OpenFeature Go SDK v1.14.1 is the target version. The `ProviderInitError`, `ProviderEventDetails.ErrorCode`, and `FlagMetadata` types are all available in this version.
- The Optimizely `client.Track` method accepts arbitrary event tags as `map[string]interface{}`, allowing us to forward tracking details without upstream changes.
- The `TrackingEventDetails` interface in the OpenFeature Go SDK exposes `Value()` for the numeric value and an attribute accessor for custom fields.
- The Optimizely SDK's `OptimizelyDecision` type exposes `FlagKey`, `RuleKey`, and `Reasons` fields that can be forwarded as flag metadata.
- The Optimizely SDK's notification center may not currently provide granular connectivity/staleness notifications, making User Story 5 (PROVIDER_STALE) dependent on upstream capability. If not available, this story will be deferred.
- The `revenue` key is the conventional Optimizely event tag name for numeric tracking values.
