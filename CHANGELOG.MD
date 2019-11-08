# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
Changes that have landed but are not yet released.	Changes that have landed but are not yet released.

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
