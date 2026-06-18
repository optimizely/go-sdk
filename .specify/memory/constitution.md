<!--
  Sync Impact Report
  ==================================================================
  Version change: 1.1.0 -> 1.2.0
  Modified principles: None renamed
  Added sections:
    - VIII. Idiomatic Go (new Core Principle)
  Removed sections: None
  Templates requiring updates:
    - .specify/templates/plan-template.md: ✅ no changes needed
      (Constitution Check section is dynamic)
    - .specify/templates/spec-template.md: ✅ no changes needed
    - .specify/templates/tasks-template.md: ✅ no changes needed
  Follow-up TODOs: None
  ==================================================================
-->

# Optimizely Go SDK Constitution

## Core Principles

### I. Backward Compatibility

All changes to the public API surface MUST preserve backward
compatibility within a major version. This includes exported types,
functions, methods, constants, interfaces, and the module path
(`github.com/optimizely/go-sdk/v2`).

- Removing or renaming an exported symbol requires a major version
  bump and a documented migration path.
- Adding optional behavior MUST use functional options, new methods,
  or new types rather than changing existing signatures.
- Serialization formats (datafile JSON, event payloads, ODP payloads)
  MUST NOT drop fields that downstream consumers rely on without a
  major version boundary.
- Rationale: Enterprise customers pin SDK versions and integrate
  deeply. Breaking changes impose costly upgrade cycles and erode
  trust.

### II. Correctness & Determinism

Decision logic (bucketing, targeting, CMAB, rollouts, holdouts) MUST
produce identical results for the same inputs across all Optimizely
SDKs that share the Full Stack specification.

- MurmurHash3-based bucketing MUST match the reference implementation
  bit-for-bit.
- Audience evaluation (AND/OR/NOT condition trees, attribute matchers)
  MUST follow the documented evaluation semantics, including null and
  missing-attribute handling.
- Changes to decision logic MUST include test cases derived from the
  Full Stack Compatibility (FSC) suite or an equivalent
  cross-SDK-validated fixture.
- Rationale: Customers rely on consistent experiment assignment.
  A single bucketing divergence can silently corrupt experiment
  results and erode statistical validity.

### III. Test-Driven Development (NON-NEGOTIABLE)

All new code and bug fixes MUST follow a strict Test-Driven
Development (TDD) workflow. Tests are written first; implementation
follows in a red-green-refactor cycle.

- **Red**: Write tests that define the expected behavior BEFORE
  writing any implementation code. Tests MUST compile and FAIL
  (red) to confirm they are exercising the correct condition.
- **Green**: Write the minimum implementation necessary to make
  the failing tests pass. No speculative code beyond what the
  tests demand.
- **Refactor**: Once green, improve structure, naming, and
  duplication without changing behavior. Tests MUST remain green
  after refactoring.
- New decision-path logic MUST include table-driven tests covering
  nominal, boundary, and error cases — written before the logic
  they validate.
- Bug fixes MUST include a regression test that fails without the
  fix (red), then apply the fix to turn it green.
- Integration tests (FSC suite) are run in CI and MUST NOT be
  broken by any change to `./pkg/...`.
- Test helpers and mocks MUST NOT leak into non-test build tags.
- `make lint` and `make test` MUST pass before merge.
- Rationale: The SDK executes inside customer applications in
  production. Defects are not patchable server-side; they require
  customers to upgrade. TDD ensures every line of implementation
  is justified by a test, preventing untested code paths from
  reaching customers.

### IV. Performance & Resource Efficiency

The SDK runs inside customer applications, often on the hot path of
request handling. Resource consumption MUST be predictable and
bounded.

- No unbounded goroutine creation, unbounded channel growth, or
  unbounded in-memory caching.
- Polling intervals, batch sizes, and flush intervals MUST have
  sensible defaults and MUST be configurable.
- Synchronous decision methods (`Decide`, `IsFeatureEnabled`,
  `GetFeatureVariable*`) MUST NOT perform network I/O; they operate
  on an in-memory `ProjectConfig`.
- Benchmark regressions (`make benchmark`) MUST be investigated
  before merge when changes touch allocation-sensitive paths
  (bucketing, config parsing, event batching).
- Rationale: Latency or memory regressions in the SDK directly
  degrade customer application performance, often without immediate
  visibility.

### V. Cross-SDK Parity

Feature behavior and public API semantics MUST align with the
Optimizely SDK specification shared across language SDKs (JavaScript,
Python, Java, Swift, etc.).

- New features (e.g., CMAB, ODP, holdouts, feature rollouts) MUST
  implement the specification contract; Go-idiomatic adaptations are
  permitted for method signatures and error handling but MUST NOT
  alter observable behavior.
- Decision reasons, notification types, and event payload schemas
  MUST match the specification.
- Deviations from cross-SDK parity MUST be documented and approved
  by the SDK team.
- Rationale: Customers operate multi-platform stacks and expect
  consistent behavior regardless of which SDK processes a given
  user's decision.

### VI. Security & Data Stewardship

The SDK handles customer data (user attributes, event payloads,
SDK keys) and MUST treat it responsibly.

- SDK keys, user identifiers, and attribute values MUST NOT be
  logged at INFO level or above. DEBUG-level logging MAY include
  them.
- HTTP clients MUST enforce TLS for all outbound connections
  (CDN, event endpoint, ODP, CMAB).
- Dependencies MUST be kept current with respect to known
  vulnerabilities (Dependabot/Renovate alerts MUST be triaged
  within one release cycle).
- Cryptographic primitives used for bucketing (MurmurHash3) are
  explicitly non-security-critical; `gosec` suppressions for these
  are intentional and MUST NOT be removed.
- Rationale: Enterprise customers operate under regulatory
  frameworks (SOC 2, GDPR, HIPAA). The SDK MUST NOT become a
  compliance liability.

### VII. Observability & Debuggability

The SDK MUST provide sufficient instrumentation for enterprise
support, debugging, and operational monitoring.

- The notification center MUST emit events for decision, track,
  config update, and log-event lifecycle stages.
- Structured logging via `pkg/logging` MUST be used consistently;
  ad-hoc `fmt.Print` calls are prohibited in library code.
- OpenTelemetry tracing (`pkg/tracing`) MUST cover all public
  client methods; span names MUST be stable across patch releases.
- Error returns MUST include actionable context (which entity,
  which operation, what failed) rather than generic messages.
- Rationale: When an enterprise customer reports unexpected
  behavior, SDK support and the customer's SRE team need
  deterministic, structured signals to diagnose issues without
  requiring a custom SDK build.

### VIII. Idiomatic Go

All code MUST follow established Go conventions as defined by
Effective Go, the Go Code Review Comments wiki, and the Go
Proverbs. The SDK MUST read as natural Go to any experienced Go
developer.

- **Error handling**: Functions MUST return `error` rather than
  panic. Errors MUST be wrapped with context using `fmt.Errorf`
  and the `%w` verb to preserve the error chain. Sentinel errors
  MUST be declared as package-level `var` values, not ad-hoc
  string comparisons.
- **Interfaces**: Accept interfaces, return concrete types.
  Interfaces MUST be small (1-3 methods) and defined by the
  consumer, not the implementer, unless shared across packages.
- **Naming**: Follow Go naming conventions — MixedCaps, not
  underscores; acronyms fully capitalized (`ID`, `HTTP`, `URL`).
  Note: existing public API symbols (e.g., `Ids`) that predate
  this principle are exempt; `revive/var-naming` remains disabled
  to protect backward compatibility (see Principle I).
- **Package design**: Packages MUST have a clear, singular
  purpose. Avoid package names like `util`, `common`, or
  `helpers`. Import cycles are forbidden.
- **Zero values**: Types SHOULD be designed so their zero value
  is useful or at minimum safe (no nil-pointer panics on
  zero-value receivers).
- **Concurrency**: Use goroutines and channels idiomatically.
  Prefer `sync.Mutex` or `sync.RWMutex` for protecting shared
  state over channel-based locking when the semantics are simple
  mutual exclusion. Always document goroutine ownership and
  lifecycle.
- **Comments**: Exported symbols MUST have doc comments following
  `godoc` conventions (start with the symbol name). Unexported
  code SHOULD have comments only where the intent is non-obvious.
- Rationale: Idiomatic code is easier to review, maintain, and
  contribute to. As an open-source SDK, the codebase MUST be
  approachable to the broader Go community and MUST NOT require
  contributors to learn project-specific conventions that diverge
  from the language's norms.

## Enterprise SDK Constraints

- **Go version support**: The SDK MUST compile and pass tests on
  the minimum supported Go version (currently 1.21) and the latest
  stable Go release.
- **Dependency discipline**: Direct dependencies MUST be minimized.
  New external dependencies require justification and review for
  license compatibility (Apache 2.0 compatible).
- **Module path**: The `/v2` suffix in the module path MUST be
  preserved in all import paths and documentation.
- **Copyright header**: Every source file MUST carry the Optimizely
  Apache 2.0 copyright header. The year line MUST be updated when
  a file is modified.
- **Thread safety**: All public types that may be used concurrently
  (notably `OptimizelyClient`, config managers, event processor)
  MUST be safe for concurrent use. Data races detected by
  `go test -race` are treated as P0 defects.

## Development Workflow & Quality Gates

- **Trunk-based development**: `master` is the mainline. Feature
  branches are short-lived and merge via pull request.
- **CI gates**: All PRs MUST pass `make lint`, `make test`, and
  the FSC integration suite before merge.
- **Linter configuration**: `.golangci.yml` intentionally disables
  specific `revive` rules (`var-naming`, `exported`,
  `unused-parameter`, `superfluous-else`) to preserve public API
  stability. These MUST NOT be re-enabled as part of unrelated work.
- **Release process**: Releases are cut via Git tags from `master`.
  Each release MUST include a changelog entry. Semantic versioning
  is enforced: MAJOR for breaking changes, MINOR for new features,
  PATCH for bug fixes.
- **Code review**: Every PR requires at least one approving review
  from an SDK team member. Changes to decision logic or public API
  require review from a senior SDK engineer.

## Governance

This constitution is the authoritative reference for development
standards in the Optimizely Go SDK. It supersedes informal
conventions and ad-hoc practices.

- **Amendment procedure**: Amendments require a pull request that
  modifies this file, with review and approval from at least two
  SDK team members. The PR description MUST state the rationale
  and the version bump classification (MAJOR/MINOR/PATCH).
- **Versioning**: This constitution follows semantic versioning.
  MAJOR: principle removal or incompatible redefinition. MINOR: new
  principle or materially expanded guidance. PATCH: clarification,
  wording, or typo fix.
- **Compliance review**: During code review, reviewers SHOULD
  verify that the change aligns with the principles above.
  Non-compliance MUST be flagged and resolved before merge.
- **Guidance file**: See `CLAUDE.md` for runtime development
  guidance, architecture details, and common commands.

**Version**: 1.2.0 | **Ratified**: 2026-04-10 | **Last Amended**: 2026-04-10
