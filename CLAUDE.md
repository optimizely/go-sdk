# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

Optimizely Go SDK v2 — feature flag / A/B testing / feature experimentation client library. Module path is `github.com/optimizely/go-sdk/v2` (the `/v2` suffix must be preserved in all import paths). Requires Go 1.21+. CI lints on Go 1.24 and runs tests on both the latest Go and 1.21.

## Common commands

All commands are driven by the `Makefile` and operate on `./pkg/...` (the root `optimizely.go` is a thin wrapper and is intentionally excluded from most targets).

```bash
make install     # one-time: installs golangci-lint v1.64.2 into $GOPATH/bin
make test        # run unit tests in ./pkg/...
make cover       # run tests with -race and write profile.cov
make lint        # runs golangci-lint per .golangci.yml, excludes *_test.go files
make benchmark   # runs benchmarks (-bench=. -run=^a) in ./pkg/...
```

Running a single test or package:

```bash
# one package
go test ./pkg/decision/...

# one test (matches TestDecide* in pkg/client)
go test ./pkg/client -run TestDecide -v

# with race detector (matches CI)
go test -race ./pkg/client -run TestDecideForKeys
```

Integration tests live outside `./pkg` and are driven by `scripts/run-fsc-tests.sh`, which pulls the Full Stack Compatibility suite (Gherkin features + datafiles) and runs `go test ./tests/integration`. These require `FEATURES_PATH` and `DATAFILES_PATH` and are normally only run in CI (`.github/workflows/integration_test.yml`).

## Architecture

The SDK is a layered system that turns a remote JSON "datafile" into feature/experiment decisions for a given user. Understanding the flow across these layers is necessary before changing behavior in any one of them.

**Public entry points** (`optimizely.go`, `pkg/client/`):
- `optimizely.Client(sdkKey)` in `optimizely.go` is a convenience wrapper around `client.OptimizelyFactory{SDKKey: ...}.Client(...)`.
- `pkg/client/factory.go` (`OptimizelyFactory`) is where construction actually happens. It wires together every subsystem below via functional options (`OptionFunc`). Most real customization (custom config manager, event dispatcher, user profile service, ODP manager, CMAB cache, tracer) goes through factory options — not through the top-level `optimizely.Client` helper.
- `pkg/client/client.go` (`OptimizelyClient`) is the user-facing API: `Decide`, `DecideForKeys`, `DecideAll`, `Activate`, `Track`, `IsFeatureEnabled`, `GetFeatureVariable*`, etc. Most methods delegate to a `UserContext` (`optimizely_user_context.go`) and then down into the decision service.

**Config layer** (`pkg/config/`):
- `config.ProjectConfigManager` is the interface that produces a `ProjectConfig` (the in-memory view of the datafile). Two implementations ship:
  - `PollingProjectConfigManager` (`polling_manager.go`) — fetches and periodically re-fetches a datafile from the Optimizely CDN.
  - `StaticProjectConfigManager` (`static_manager.go`) — wraps a hard-coded datafile passed via `OptimizelyFactory.Datafile`.
- `pkg/config/datafileprojectconfig/` parses the raw datafile JSON into typed entities (`entities/`, `mappers/`). If you are changing how a datafile field is interpreted, the change almost always lives in `mappers/` and the `DatafileProjectConfig` type.

**Decision layer** (`pkg/decision/`) — the core of the SDK. This is a composition of small services that each implement `ExperimentService` or `FeatureService` and can short-circuit a decision:
- `CompositeService` (`composite_service.go`) is the top-level `decision.Service` consumed by the client. It owns a `CompositeFeatureService` (for feature flags / rollouts) and a `CompositeExperimentService` (for experiments).
- `CompositeExperimentService` runs a pipeline: forced-decision → whitelist → user-profile (persisted) → CMAB → bucketer. First service to return a decision wins.
- `CompositeFeatureService` evaluates feature tests first via the experiment pipeline above, then falls back to `RolloutService` for targeted delivery rollouts.
- Individual services worth knowing when debugging a decision:
  - `experiment_bucketer_service.go` + `bucketer/` — MurmurHash3-based traffic allocation.
  - `experiment_cmab_service.go` — Contextual Multi-Armed Bandit decisions; talks to the CMAB prediction service via `pkg/cmab/`.
  - `rollout_service.go` — targeted rollouts / feature rollouts (FR).
  - `holdout_service.go` — holdout gating applied before regular decisions.
  - `forced_decision_service.go`, `experiment_override_service.go`, `experiment_whitelist_service.go` — various override paths.
  - `persisting_experiment_service.go` — reads/writes through a `UserProfileService` so returning users get sticky bucketing.
- `evaluator/` handles audience condition-tree evaluation (AND/OR/NOT trees of attribute matchers).
- `reasons/` defines the enumerated reasons appended to a `DecisionReasons` (surfaced when users pass `decide.IncludeReasons`).

**Event layer** (`pkg/event/`):
- `Processor` (default: `BatchEventProcessor` in `processor.go`) buffers impression/conversion events via a `Queue` and flushes them through a `Dispatcher` (`dispatcher.go`, which POSTs to the Optimizely events endpoint).
- `factory.go` builds the default processor wired into `OptimizelyFactory`. Swap either piece via `client.WithEventProcessor` / `client.WithEventDispatcher`.

**ODP layer** (`pkg/odp/`): Optimizely Data Platform integration for audience segments.
- `odp_manager.go` coordinates a `segment/` manager (fetches qualified segments for a user) and an `event/` manager (sends ODP events). Enabled by default; disable via factory option `client.WithOdpDisabled(true)`.

**CMAB layer** (`pkg/cmab/`): Contextual Multi-Armed Bandit prediction client. `client.go` is the HTTP client against the prediction endpoint, `service.go` wraps it with a cache (`cache/`). Configured at the client level via `CmabConfig` in `pkg/client/factory.go` — the client-level `CmabConfig` intentionally mirrors the internal `cmab.Config` so the public surface can stay stable.

**Supporting packages**:
- `pkg/entities/` — typed domain model (`Experiment`, `Feature`, `Variation`, `UserContext`, `Audience`, etc.). Note `ExperimentType` constants: `ab`, `mab`, `cmab`, `td`, `fr`.
- `pkg/decide/` — `Options` flags (e.g. `IncludeReasons`, `DisableDecisionEvent`) and the `OptimizelyDecision` result type returned to callers.
- `pkg/notification/` + `pkg/registry/` — notification center; listeners for `Decision`, `Track`, `LogEvent`, `ProjectConfigUpdate`, etc. The registry gives each SDK key its own notification center.
- `pkg/logging/` — `OptimizelyLogProducer` used throughout; logger is keyed by SDK key + component name.
- `pkg/tracing/` — OpenTelemetry integration. Span names are exported as `client.SpanName*` constants.
- `pkg/metrics/` — pluggable metrics registry (wired via `WithMetricsRegistry`).
- `pkg/cache/`, `pkg/utils/`, `pkg/optimizelyjson/` — generic helpers.

### Per-SDK-key singletons

Several subsystems (notification center, logger) are keyed by the SDK key through `pkg/registry/`. Two `OptimizelyClient` instances created with the same SDK key will share those singletons. Keep this in mind when writing tests that touch notifications or when reasoning about cross-instance state leaks.

## Linting notes

The linter config (`.golangci.yml`) deliberately disables a few `revive` rules (`var-naming`, `exported`, `unused-parameter`, `superfluous-else`) because fixing them would break the public API (e.g. `Ids` → `IDs` rename) or because the code intentionally implements interfaces with unused parameters. Don't re-enable these or "fix" flagged instances as part of unrelated work. `gosec`'s "weak cryptographic primitive" warning is also excluded because the bucketer uses MurmurHash3 for traffic allocation, not security.

## Contribution conventions

- Every source file must carry the Optimizely Apache 2.0 copyright header (see `CONTRIBUTING.md` for the exact block). Update the year line when editing an existing file.
- PRs require accompanying unit tests; `make lint` and `make test` must both pass.
- The project follows trunk-based development on `master`; `master` is not always the most stable — releases are cut via Git tags.
