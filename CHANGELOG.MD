# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
Changes that have landed but are not yet released.	Changes that have landed but are not yet released.
 - None

## [1.0.0-beta1] - September 27th, 2019

### New Features
- We've added decision listeners and project config update listeners.
- Added optional configuration parameters to NewProcessor.

### Breaking Changes
- This method OptimizelyFactory.ClientWithOptions() has been removed in favor of OptimizelyFactory.Client(clientOptions ...OptionFunc)

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
