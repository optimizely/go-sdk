# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
Changes that have landed but are not yet released.	Changes that have landed but are not yet released.

## [1.0.1] - March 6th, 2020
## New Features
- feat: Allow ClientName and Version to be set at compile time. [#227](https://github.com/optimizely/go-sdk/pull/227)
- feat: Add explicit json mappings.  [#228](https://github.com/optimizely/go-sdk/pull/228)
- refact(snapshot): Made decisions optional for snapshot entity. [#226](https://github.com/optimizely/go-sdk/pull/226)
- refact: Updated polling_manager implementation and unit tests. [#224](https://github.com/optimizely/go-sdk/pull/224)
- feat: set specific error from default project config manager when CDN returns 403 response  [#232](https://github.com/optimizely/go-sdk/pull/232)
- feat: OnEvent should be part of the event processor interface [#233](https://github.com/optimizely/go-sdk/pull/233)
- feat: add valid datafile version check [#234](https://github.com/optimizely/go-sdk/pull/234)

## Bug Fixes
- fix: Properly instantiate event processor with dispatcher. [#239](https://github.com/optimizely/go-sdk/pull/239)

## Breaking Changes
- feat: Return variables as an interface{} from GetAllFeatureVariables. [#235](https://github.com/optimizely/go-sdk/pull/235)


## [1.0.0] - January 9th, 2020
This is the General Availability release of the Go SDK.

## New Features
- feat(onTrack): Added onTrack callback. [#198](https://github.com/optimizely/go-sdk/pull/198)
- feat: Add observability metrics for event queue. [#211](https://github.com/optimizely/go-sdk/pull/211)
- feat: Implement OptimizelyConfig API. [#216](https://github.com/optimizely/go-sdk/pull/216)
- feat: Add observability metrics for event queue. [#223](https://github.com/optimizely/go-sdk/pull/223)

## Bug Fixes
- fix: Revert signature changes to PollingProjectConfigManager. [#209](https://github.com/optimizely/go-sdk/pull/209)
- fix(notification): race detected send notification issue. [#222](https://github.com/optimizely/go-sdk/pull/222)

## Breaking Changes
- refac: Execution context. [#212](https://github.com/optimizely/go-sdk/pull/212)
- refac(config: Move config interfaces back into config package. [#213](https://github.com/optimizely/go-sdk/pull/213)

## [1.0.0-rc1] - December 4th, 2019

## New Features
- feat(PollingConfigManager): Implemented caching headers in PollingConfigManager. [#189](https://github.com/optimizely/go-sdk/pull/189)
- feat(api): Add top-level wrapper for more convenient usage. [#191](https://github.com/optimizely/go-sdk/pull/191)

### Bug Fixes
- fix(audience-evaluation): Fixed null-bubbling issues. [#177](https://github.com/optimizely/go-sdk/pull/177)
- fix: correct error returned when SDK key is invalid [#203](https://github.com/optimizely/go-sdk/pull/203)
- fix: Invalid datafile JSON should return an error [#205](https://github.com/optimizely/go-sdk/pull/205)

## [1.0.0-beta7] - November 14th, 2019

## New Features
- feat: Method for removing forced variations [#176](https://github.com/optimizely/go-sdk/pull/176)

### Bug Fixes
- fix: Update built-in map-based ExperimentOverrideStore implementation to be concurrently usable [#171](https://github.com/optimizely/go-sdk/pull/171)
- fix: Add a public constructor for MapExperimentOverridesStore [#174](https://github.com/optimizely/go-sdk/pull/174)
- fix: don't return error when there is a good projectConfig in polling manager [#179](https://github.com/optimizely/go-sdk/pull/179)
- fix(decision): Logs produced by the various decision services. [#180](https://github.com/optimizely/go-sdk/pull/180)
- fix(exists-condition) Fixed exists comparison for leaf condition. [#185](https://github.com/optimizely/go-sdk/pull/185)

## [1.0.0-beta6] - November 8th, 2019

## New Features
- Experiment override service - implement groups
[#164](https://github.com/optimizely/go-sdk/pull/164)
- Add User profile service
[#163](https://github.com/optimizely/go-sdk/pull/163)

### Bug Fixes
- Fix config managers so that they don't try and parse on error returned from CDN. [#170](https://github.com/optimizely/go-sdk/pull/170)
- When event batch size has been reached only start one batch event processing go routine.
- When queue size is met, log a message and do not add to the queue.
- Duration used was setting the time too far into the future by multiplying by second and then by milliseconds. Flush interval is now any duration, default is 30 seconds.  If you don't pass in a multiplier the duration created is in microseconds.
[#167](https://github.com/optimizely/go-sdk/pull/167)
- fixed parsing for audience conditions.
[#165] (https://github.com/optimizely/go-sdk/pull/165)
- Check nil to prevent panic. [#162] (https://github.com/optimizely/go-sdk/pull/162)
- fix: support audience ids. [#161] (https://github.com/optimizely/go-sdk/pull/161)

## [1.0.0-beta5] - October 30th, 2019

### Bug Fixes
- fix(ExperimentWhitelistService): Fix finding variation by key [#155](https://github.com/optimizely/go-sdk/pull/155)

## [1.0.0-beta4] - October 28th, 2019

### New Features
- feat(logger): Expand LogConsumer interface to have explicit field mapping. [#147](https://github.com/optimizely/go-sdk/pull/147)

### Bug Fixes
- refac(client): Revise how the client handles and returns errors. [#151](https://github.com/optimizely/go-sdk/pull/151)

## [1.0.0-beta3] - October 15th, 2019

### Breaking Changes
- Refactored EventProcessor and ConfigManager components to no longer require an ExecutionContext be passed in during construction. It will be injected into them by the client factory. [#130](https://github.com/optimizely/go-sdk/pull/130)
- Moved the entire `optimizely` package from `github.com/optimizely/go-sdk/optimizely` to `github.com/optimizely/go-sdk/pkg` to better follow convention and best practices. [#135](https://github.com/optimizely/go-sdk/pull/135)

## [1.0.0-beta2] - October 2nd, 2019

### New Features
- Introduces the `Activate` and `GetVariation` APIs for running A/B tests.
- Makes the top-level client components public: `DecisionService, ConfigManager, EventProcessor`.

## [1.0.0-beta1] - September 27th, 2019

### New Features
- Added decision listeners and project config update listeners.
- Added optional configuration parameters to NewProcessor.

### Breaking Changes
- This method `OptimizelyFactory.ClientWithOptions()` has been removed in favor of `OptimizelyFactory.Client(clientOptions ...OptionFunc)`

## [0.2.0] - September 11th, 2019
This release of the SDK introduces Feature Management capabilities for running both Feature Rollouts as well as Feature Tests using Optimizely Feature Management.

### New Features
- Introduces feature variable getters via `GetFeatureVariable*` for parameterizing your feature tests.
- Introduces the `Track` API call for sending conversion events to Optimizely.
- The `IsFeatureEnabled` API will now send impression events if there is a feature test attached to the feature being accessed.

### Breaking Changes
- Vendored packages have been removed in favor of declaring dependencies and their versions using `go.mod`.

## [0.1.0-beta] - August 23rd, 2019
This is the initial release of the SDK, which includes support for running Feature Rollouts using Optimizely Feature Management.

### New Features
- Introduces the `IsFeatureEnabled` API for determining whether the feature is enabled for a user or not. Please see our [README](README.MD).
