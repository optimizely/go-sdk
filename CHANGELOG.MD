# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [2.1.0] - November 8, 2024

### Enhancements
- **Batch Processing for User Profile Service (UPS)**: Added support for batch processing in `decideAll` and `decideForKeys`, enabling more efficient handling of multiple decisions in the User Profile Service. ([#394](https://github.com/optimizely/go-sdk/pull/394))

- **Optimized Logging**: Reduced log verbosity in the Decision flow by shifting info-level logs to debug. This change will significantly reduce the volume of logs at the info level, improving readability and performance for applications with high decision-making activity. ([#390](https://github.com/optimizely/go-sdk/pull/390))

## [2.0.0] - January 22th, 2024

### New Features  

The 2.0.0 release introduces a new primary feature, [Advanced Audience Targeting]( https://docs.developers.optimizely.com/feature-experimentation/docs/optimizely-data-platform-advanced-audience-targeting) enabled through integration with [Optimizely Data Platform (ODP)](https://docs.developers.optimizely.com/optimizely-data-platform/docs) ([#350](https://github.com/optimizely/go-sdk/pull/350), [#353](https://github.com/optimizely/go-sdk/pull/353), [#354](https://github.com/optimizely/go-sdk/pull/354), [#355](https://github.com/optimizely/go-sdk/pull/355)).

You can use ODP, a high-performance [Customer Data Platform (CDP)]( https://www.optimizely.com/optimization-glossary/customer-data-platform/), to easily create complex real-time segments (RTS) using first-party and 50+ third-party data sources out of the box. You can create custom schemas that support the user attributes important for your business, and stitch together user behavior done on different devices to better understand and target your customers for personalized user experiences. ODP can be used as a single source of truth for these segments in any Optimizely or 3rd party tool.  

With ODP accounts integrated into Optimizely projects, you can build audiences using segments pre-defined in ODP. The SDK will fetch the segments for given users and make decisions using the segments. For access to ODP audience targeting in your Feature Experimentation account, please contact your Optimizely Customer Success Manager.

This version includes the following changes: 

* New API added to `OptimizelyUserContext`: 

	- `FetchQualifiedSegments()`: this API will retrieve user segments from the ODP server. The fetched segments will be used for audience evaluation. The fetched data will be stored in the local cache to avoid repeated network delays. 

	- When an `OptimizelyUserContext` is created, the SDK will automatically send an identify request to the ODP server to facilitate observing user activities. 

* New APIs added to `OptimizelyClient`: 

	- `SendOdpEvent()`: customers can build/send arbitrary ODP events that will bind user identifiers and data to user profiles in ODP.

For details, refer to our documentation pages:

* [Advanced Audience Targeting](https://docs.developers.optimizely.com/feature-experimentation/docs/optimizely-data-platform-advanced-audience-targeting)  

* [Server SDK Support](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/advanced-audience-targeting-for-server-side-sdks) 

* [Initialize Go SDK](https://docs.developers.optimizely.com/feature-experimentation/docs/initialize-sdk-go) 

* [OptimizelyUserContext Go SDK](https://docs.developers.optimizely.com/feature-experimentation/docs/optimizelyusercontext-go) 

* [Advanced Audience Targeting segment qualification methods](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/advanced-audience-targeting-segment-qualification-methods-go) 

* [Send Optimizely Data Platform data using Advanced Audience Targeting](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/send-odp-data-using-advanced-audience-targeting-go) 

### Breaking Changes 

* ODPManager in the SDK is enabled by default. Unless an ODP account is integrated into the Optimizely projects, most ODPManager functions will be ignored. If needed, ODPManager can be disabled when `OptimizelyClient` is instantiated. 
* Minimum golang version changed from 1.13 to 1.21.0. ([#369](https://github.com/optimizely/go-sdk/pull/369), [#380](https://github.com/optimizely/go-sdk/pull/380))
* Module name has been changed to `github.com/optimizely/go-sdk/v2`.

### Bug Fixes
* Logged correct error message in HTTP requests. ([#374](https://github.com/optimizely/go-sdk/pull/374))
* Stopped all tickers to prevent goroutine leak. ([#375](https://github.com/optimizely/go-sdk/pull/375)).
* Client is closed gracefully to dispatch queued events. ([#376](https://github.com/optimizely/go-sdk/pull/376))
* Convert warning into error log for SyncConfig failure. ([#383](https://github.com/optimizely/go-sdk/pull/383))

### Functionality Enhancement
* Github issue template is updated. ([#379](https://github.com/optimizely/go-sdk/pull/379))
* Helper function for passing notification center is added. ([#381](https://github.com/optimizely/go-sdk/pull/381))
* OpenTelemetry tracing support is added. ([#385](https://github.com/optimizely/go-sdk/pull/385))
  - New ClientOptionFunc `WithTracer()` is added to initiate `OptimizelyClient` with OpenTelemetry tracer.
  - New method `WithTraceContext()` is added to `OptimizelyClient` to set the trace context before calling APIs.

## [1.8.5] - October 5th, 2023

* Fixed a bug in the HTTP Requester logging. Now correct error message is logged. ([#374](https://github.com/optimizely/go-sdk/pull/374))

## [2.0.0-beta] - April 27th, 2023

### New Features  

The 2.0.0-beta release introduces a new primary feature, [Advanced Audience Targeting]( https://docs.developers.optimizely.com/feature-experimentation/docs/optimizely-data-platform-advanced-audience-targeting) enabled through integration with [Optimizely Data Platform (ODP)](https://docs.developers.optimizely.com/optimizely-data-platform/docs) ([#350](https://github.com/optimizely/go-sdk/pull/350), [#353](https://github.com/optimizely/go-sdk/pull/353), [#354](https://github.com/optimizely/go-sdk/pull/354), [#355](https://github.com/optimizely/go-sdk/pull/355)).

You can use ODP, a high-performance [Customer Data Platform (CDP)]( https://www.optimizely.com/optimization-glossary/customer-data-platform/), to easily create complex real-time segments (RTS) using first-party and 50+ third-party data sources out of the box. You can create custom schemas that support the user attributes important for your business, and stitch together user behavior done on different devices to better understand and target your customers for personalized user experiences. ODP can be used as a single source of truth for these segments in any Optimizely or 3rd party tool.  

With ODP accounts integrated into Optimizely projects, you can build audiences using segments pre-defined in ODP. The SDK will fetch the segments for given users and make decisions using the segments. For access to ODP audience targeting in your Feature Experimentation account, please contact your Optimizely Customer Success Manager.

This version includes the following changes: 

* New API added to `OptimizelyUserContext`: 

	- `FetchQualifiedSegments()`: this API will retrieve user segments from the ODP server. The fetched segments will be used for audience evaluation. The fetched data will be stored in the local cache to avoid repeated network delays. 

	- When an `OptimizelyUserContext` is created, the SDK will automatically send an identify request to the ODP server to facilitate observing user activities. 

* New APIs added to `OptimizelyClient`: 

	- `SendOdpEvent()`: customers can build/send arbitrary ODP events that will bind user identifiers and data to user profiles in ODP. 

For details, refer to our documentation pages:  

* [Advanced Audience Targeting](https://docs.developers.optimizely.com/feature-experimentation/docs/optimizely-data-platform-advanced-audience-targeting)  

* [Server SDK Support](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/advanced-audience-targeting-for-server-side-sdks) 

* [Initialize Go SDK](https://docs.developers.optimizely.com/feature-experimentation/docs/initialize-sdk-go) 

* [OptimizelyUserContext Go SDK](https://docs.developers.optimizely.com/feature-experimentation/docs/optimizelyusercontext-go) 

* [Advanced Audience Targeting segment qualification methods](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/advanced-audience-targeting-segment-qualification-methods-go) 

* [Send Optimizely Data Platform data using Advanced Audience Targeting](https://docs.developers.optimizely.com/feature-experimentation/v1.0/docs/send-odp-data-using-advanced-audience-targeting-go) 

### Breaking Changes 

* ODPManager in the SDK is enabled by default. Unless an ODP account is integrated into the Optimizely projects, most ODPManager functions will be ignored. If needed, ODPManager can be disabled when `OptimizelyClient` is instantiated. 
* Minimum golang version changed from 1.13 to 1.19. ([#369](https://github.com/optimizely/go-sdk/pull/369))

## [1.8.4] - March 10th, 2023

* We updated our README.md and other non-functional code to reflect that this SDK supports both Optimizely Feature Experimentation and Optimizely Full Stack. ([#366](https://github.com/optimizely/go-sdk/pull/366))

## [1.8.3] - October 12, 2022

### New Features
* Add option to set custom http.Client on HTTPRequester. ([#335](https://github.com/optimizely/go-sdk/pull/335))

### Bug Fix
* Fixes issue with visitor where dispatch timestamp was used instead of userEvent creation timestamp. ([#345](https://github.com/optimizely/go-sdk/pull/345))

## [1.8.2] - September 1, 2022

### Bug Fix
* Add additional validation for numeric attribute types when building an event object. ([#341](https://github.com/optimizely/go-sdk/pull/341))

## [1.8.1] - August 29, 2022

### Bug Fix
* Add validation for user attribute types when building an event object. Only allow primitive types ([#339](https://github.com/optimizely/go-sdk/pull/339))

## [1.8.0] - January 12, 2022

### New Features
* Add a set of new APIs for overriding and managing user-level flag, experiment and delivery rule decisions. These methods can be used for QA and automated testing purposes. They are an extension of the OptimizelyUserContext interface ([#324](https://github.com/optimizely/go-sdk/pull/324), [#326](https://github.com/optimizely/go-sdk/pull/326), [#328](https://github.com/optimizely/go-sdk/pull/328)):
	- setForcedDecision
	- getForcedDecision
	- removeForcedDecision
	- removeAllForcedDecisions

* For details, refer to our documentation pages: [OptimizelyUserContext](https://docs.developers.optimizely.com/full-stack/v4.0/docs/optimizelyusercontext-go) and [Forced Decision methods](https://docs.developers.optimizely.com/full-stack/v4.0/docs/forced-decision-methods-go).

## [1.7.0] - September 17, 2021

### New Features
* Add new public properties to `OptimizelyConfig`. ([#322](https://github.com/optimizely/go-sdk/pull/322), [#312](https://github.com/optimizely/go-sdk/pull/312))
	- SdkKey
 	- EnvironmentKey
	- Attributes
	- Audiences
	- Events
	- ExperimentRules and DeliveryRules to `OptimizelyFeature`
	- Audiences to `OptimizelyExperiment`
* For details, refer to our documentation page: [https://docs.developers.optimizely.com/full-stack/v4.0/docs/optimizelyconfig-go](https://docs.developers.optimizely.com/full-stack/v4.0/docs/optimizelyconfig-go).

### Deprecated

* `OptimizelyFeature.ExperimentsMap` of `OptimizelyConfig` is deprecated as of this release. Please use `OptimizelyFeature.ExperimentRules` and `OptimizelyFeature.DeliveryRules`. ([#322](https://github.com/optimizely/go-sdk/pull/322))

## [1.6.1] - March 2, 2021
## Bug Fix
- Fix issues for decision variables and reasons. [#309](https://github.com/optimizely/go-sdk/pull/309)

## [1.6.0] - February 16th, 2021
## New Features
Introducing a new primary interface for retrieving feature flag status, configuration and associated experiment decisions for users ([#304](https://github.com/optimizely/go-sdk/pull/304), [#303](https://github.com/optimizely/go-sdk/pull/303), [#299](https://github.com/optimizely/go-sdk/pull/299), [#298](https://github.com/optimizely/go-sdk/pull/298), [#297](https://github.com/optimizely/go-sdk/pull/297)). The new `OptimizelyUserContext` class is instantiated with `CreateUserContext` and exposes the following APIs to get `OptimizelyDecision`:

    - SetAttribute
    - Decide
    - DecideAll
    - DecideForKeys
    - TrackEvent

For details, refer to our documentation page: https://docs.developers.optimizely.com/full-stack/v4.0/docs/go-sdk.

## [1.5.1] - November 16th, 2020
## New Features
- Add "enabled" field to decision metadata structure ([#301](https://github.com/optimizely/go-sdk/pull/301))

## [1.5.0] - October 29th, 2020
## New Features
- Correct logic for rollout service ([#296](https://github.com/optimizely/go-sdk/pull/296))
- Add support for sending flag decisions along with decision metadata ([#292](https://github.com/optimizely/go-sdk/pull/292))

## [1.4.0] - October 2nd, 2020
## New Features
- Add MatchRegistry for custom match implementations. ([#286](https://github.com/optimizely/go-sdk/pull/286)) …

- Added and refactored audience evaluation logs. ([#280](https://github.com/optimizely/go-sdk/pull/280))

- Add datafile accessor to access config as json string. ([#290](https://github.com/optimizely/go-sdk/pull/290))
 
- Support version audience evaluation following semantic versioning implementation. ([#289](https://github.com/optimizely/go-sdk/pull/289))

- Support number 'less than or equal to' and 'greater than or equal to'. ([#289](https://github.com/optimizely/go-sdk/pull/289))
## Bug Fix
- Fixed a spelling mistake in go.mod file. ([#285](https://github.com/optimizely/go-sdk/pull/285))

## [1.3.0] - July 7th, 2020
## New Features
- Add support for JSON feature variables ([#250](https://github.com/optimizely/go-sdk/pull/250), [#251](https://github.com/optimizely/go-sdk/pull/251), [#253](https://github.com/optimizely/go-sdk/pull/253))
- Add support for authenticated datafiles ([#267](https://github.com/optimizely/go-sdk/pull/267))
- Expose ability to customize URL for event dispatcher. ([#268](https://github.com/optimizely/go-sdk/pull/268))

## Bug Fix
- Provide synchronization for Notification Center Cache ([#273](https://github.com/optimizely/go-sdk/pull/273))
- Fix unbounded memory allocation in event dispatcher ([#275](https://github.com/optimizely/go-sdk/pull/275))

## [1.3.0-beta3] - June 30th, 2020
## Bug Fix
- Provide synchronization for Notification Center Cache ([#273](https://github.com/optimizely/go-sdk/pull/273))

## [1.3.0-beta2] - June 17th, 2020
## Bug Fix
- Bring back old constructors for static manager ([#271](https://github.com/optimizely/go-sdk/pull/271))

## [1.3.0-beta1] - June 16th, 2020
## New Features
- Add support for JSON feature variables ([#250](https://github.com/optimizely/go-sdk/pull/250), [#251](https://github.com/optimizely/go-sdk/pull/251), [#253](https://github.com/optimizely/go-sdk/pull/253))
- Add support for authenticated datafiles ([#267](https://github.com/optimizely/go-sdk/pull/267))
- Expose ability to customize URL for event dispatcher. ([#268](https://github.com/optimizely/go-sdk/pull/268))

## [1.2.0] - May 21st, 2020
## New Features
- feat: support for multi-rule rollouts [#247](https://github.com/optimizely/go-sdk/pull/247)

## [1.1.3] - April 22th, 2020
## Bug Fix
- logger not set for httpDispatcher [#254](https://github.com/optimizely/go-sdk/pull/254)

## [1.1.2] - March 26th, 2020
## New Features
- refact: Update logging field keys to be consumable in structured logs [#246](https://github.com/optimizely/go-sdk/pull/246)

## [1.1.1] - March 25th, 2020
## New Features
- feat: Logging sdk key masked value with every log message.  The masking is by default and can be disabled or set to a custom mapping. [#242](https://github.com/optimizely/go-sdk/pull/242)

## [1.1.0] - March 6th, 2020
## New Features
- feat: Allow ClientName and Version to be set at compile time. [#227](https://github.com/optimizely/go-sdk/pull/227)
- feat: Add explicit json mappings.  [#228](https://github.com/optimizely/go-sdk/pull/228)
- refact(snapshot): Made decisions optional for snapshot entity. [#226](https://github.com/optimizely/go-sdk/pull/226)
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
