# Optimizely Go SDK - Complete Documentation Code Samples

**Date Extracted**: 2026-02-17  
**Source**: https://docs.developers.optimizely.com/feature-experimentation/docs/go-sdk  
**Total Code Samples**: 34  
**Purpose**: Verification against actual SDK source code for correctness

---

## Table of Contents

1. [Summary by Category](#summary-by-category)
2. [Key Observations](#key-observations)
3. [All Code Samples](#all-code-samples)

---

## Summary by Category

### Installation & Setup (2 samples)
- **#1**: Basic installation command (`go get`)
- **#2**: Install from CLI source with `go install`

### Client Initialization (8 samples)
- **#3**: Basic initialization with SDK Key (background sync)
- **#4**: Static client with hard-coded datafile
- **#5**: Static client with SDK key (one-time datafile fetch)
- **#6**: Custom initialization with polling, event processing, decide options
- **#7**: Authenticated datafile access with access token
- **#23**: Event batching basic example
- **#24**: Event batching advanced example with custom options
- **#33**: Client with experiment overrides store

### ODP (Optimizely Data Platform) Configuration (9 samples)
- **#8**: Complete ODP Manager setup with Event and Segment managers
- **#9**: EventApiManager interface definition
- **#10**: Custom EventApiManager with HTTP timeout configuration
- **#11**: SegmentAPIManager interface definition
- **#12**: Custom ODPEventManager with queue size and flush interval
- **#13**: Custom ODPSegmentManager with cache configuration
- **#14**: Custom cache interface definition
- **#15**: CustomSegmentsCache implementation with mutex
- **#16**: Using custom cache with ODP Manager

### Decide Methods & User Context (6 samples)
- **#17**: Complete `decide()` example with all features
- **#18**: `DecideAll()` method for all flags
- **#19**: `DecideForKeys()` method for specific flags
- **#20**: Manual CMAB (Contextual Multi-Armed Bandit) cache control
- **#21**: CMAB decision reasons extraction
- **#22**: OptimizelyUserContext type definition with all methods

### Event Handling & Notifications (5 samples)
- **#25**: LogEvent listener registration/unregistration
- **#26**: All notification listener types setup
- **#28**: TrackEvent with event properties and tags
- **#29**: Custom event dispatcher implementation
- **#30**: Using custom event dispatcher with client

### Logging & Debugging (2 samples)
- **#31**: Custom logger implementation
- **#32**: Setting log level

### Advanced Features (2 samples)
- **#33**: Experiment overrides store configuration
- **#34**: Getting forced variation from override store

### Complete Examples (1 sample)
- **#27**: Full end-to-end usage workflow

---

## Key Observations

### Import Paths
All code samples reference both v1 and v2 import paths:
- **v1**: `github.com/optimizely/go-sdk`
- **v2**: `github.com/optimizely/go-sdk/v2` (with `/v2` suffix)

### Common Packages Referenced
- `github.com/optimizely/go-sdk/pkg/client`
- `github.com/optimizely/go-sdk/pkg/decide`
- `github.com/optimizely/go-sdk/pkg/event`
- `github.com/optimizely/go-sdk/pkg/odp`
- `github.com/optimizely/go-sdk/pkg/odp/event`
- `github.com/optimizely/go-sdk/pkg/odp/segment`
- `github.com/optimizely/go-sdk/pkg/logging`
- `github.com/optimizely/go-sdk/pkg/decision`
- `github.com/optimizely/go-sdk/pkg/utils`
- `github.com/optimizely/go-sdk/pkg/entities`

### Key Types & Functions to Verify
- `optly.Client(sdkKey)` - main client constructor
- `client.OptimizelyFactory{}` - factory for advanced configuration
- `client.CreateUserContext()` - user context creation
- `user.Decide()`, `DecideAll()`, `DecideForKeys()` - decision methods
- `user.TrackEvent()` - event tracking
- `event.NewBatchEventProcessor()` - event batching
- `odp.NewOdpManager()` - ODP manager creation
- Various `WithXxx()` option functions

### Potential Issues to Check
1. **Sample #12 (line 197)**: Has a typo - ends with `)is` instead of just `)`
2. **Sample #20 (lines 393-395, 405-407)**: Comments appear to be split incorrectly (e.g., "// Always fetch fresh from CMAB" followed by "service")
3. All interface definitions should match actual SDK interfaces
4. All option functions (`WithXxx`) should exist in SDK
5. Type assertions and method signatures should be accurate

---

## All Code Samples

GO SDK DOCUMENTATION CODE SAMPLES
Total: 27 samples
================================================================================


### CODE SAMPLE #1
File: install-sdk-go.md
Section: Install the Go SDK
Language: go
--------------------------------------------------------------------------------
go get github.com/optimizely/go-sdk // for v2: go get github.com/optimizely/go-sdk/v2
--------------------------------------------------------------------------------

### CODE SAMPLE #2
File: install-sdk-go.md
Section: Install from CLI source
Language: go
--------------------------------------------------------------------------------
go get github.com/optimizely/go-sdk // for v2: go get github.com/optimizely/go-sdk/v2
cd $GOPATH/src/github.com/optimizely/go-sdk
go install
--------------------------------------------------------------------------------

### CODE SAMPLE #3
File: initialize-sdk-go.md
Section: Basic initialization with SDK Key
Language: go
--------------------------------------------------------------------------------
import optly "github.com/optimizely/go-sdk" // for v2: "github.com/optimizely/go-sdk/v2"

// Instantiates a client that syncs the datafile in the background
optlyClient, err := optly.Client("SDK_KEY_HERE")
if err != nil{
	// handle error
}
--------------------------------------------------------------------------------

### CODE SAMPLE #4
File: initialize-sdk-go.md
Section: Static client instance
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client"

optimizelyFactory := &client.OptimizelyFactory{
          Datafile: []byte("DATAFILE_JSON_STRING_HERE"),
}

// Instantiate a static client (no datafile polling)
staticOptlyClient, err := optimizelyFactory.StaticClient()
if err != nil {
	// handle error
}
--------------------------------------------------------------------------------

### CODE SAMPLE #5
File: initialize-sdk-go.md
Section: Static client instance
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client"

optimizelyFactory := &client.OptimizelyFactory{
          SDKKey: "[SDK_KEY_HERE]",
}

// Instantiate a static client that will pull down the datafile one time
staticOptlyClient, err := optimizelyFactory.StaticClient()
if err != nil {
	// handle error
}
--------------------------------------------------------------------------------

### CODE SAMPLE #6
File: initialize-sdk-go.md
Section: Custom initialization
Language: go
--------------------------------------------------------------------------------
import (
  "time"
 
  "github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client" 
)

optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "[SDK_KEY_HERE]",
	}

datafilePollingInterval := 2 * time.Second
eventBatchSize := 20
eventQueueSize := 1500
eventFlushInterval := 10 * time.Second
defaultDecideOptions := []decide.OptimizelyDecideOptions{
	decide.IgnoreUserProfileService,
}

// Instantiate a client with custom configuration
optimizelyClient, err := optimizelyFactory.Client(
  client.WithPollingConfigManager(datafilePollingInterval, nil),
  client.WithBatchEventProcessor(
    eventBatchSize,
    eventQueueSize,
    eventFlushInterval,
  ),
  client.WithDefaultDecideOptions(defaultDecideOptions),
)
if err != nil {
	// handle error
}
--------------------------------------------------------------------------------

### CODE SAMPLE #7
File: initialize-sdk-go.md
Section: Use authenticated datafile in a secure environment
Language: go
--------------------------------------------------------------------------------
// fetch the datafile from an authenticated endpoint
accessToken := `YOUR_DATAFILE_ACCESS_TOKEN`
sdkKey := `YOUR_SDK_KEY`
factory := client.OptimizelyFactory{SDKKey: sdkKey}
optimizelyClient, err := factory.Client(client.WithDatafileAccessToken(accessToken))
if err != nil {
	// handle error
}
--------------------------------------------------------------------------------

### CODE SAMPLE #8
File: initialize-sdk-go.md
Section: `ODPEventManager`
Language: go
--------------------------------------------------------------------------------
import (
	"github.com/optimizely/go-sdk/pkg/odp"         // for v2: "github.com/optimizely/go-sdk/v2/pkg/odp"
	"github.com/optimizely/go-sdk/pkg/odp/event"   // for v2: "github.com/optimizely/go-sdk/v2/pkg/odp/event"
	"github.com/optimizely/go-sdk/pkg/odp/segment" // for v2: "github.com/optimizely/go-sdk/v2/pkg/odp/segment"
)

// You must configure Real-Time Audiences for Feature Experimentation
// before being able to calling ODPApiManager, ODPEventManager, and ODPSegmentManager.
sdkKey := `YOUR_SDK_KEY`
defaultEventApiManager := event.NewEventAPIManager(sdkKey, nil)
odpEventManager := event.NewBatchEventManager(event.WithAPIManager(defaultEventApiManager))
defaultSegmentApiManager := segment.NewSegmentAPIManager(sdkKey, nil)
odpSegmentManager := segment.NewSegmentManager(sdkKey, segment.WithAPIManager(defaultSegmentApiManager))

odpManager := odp.NewOdpManager(sdkKey, 
		false,
		odp.WithEventManager(odpEventManager),     // Optional
		odp.WithSegmentManager(odpSegmentManager), // Optional
)
--------------------------------------------------------------------------------

### CODE SAMPLE #9
File: initialize-sdk-go.md
Section: Customize `EventApiManager`
Language: go
--------------------------------------------------------------------------------
// APIManager represents the event API manager.
type APIManager interface {
	SendOdpEvents(apiKey, apiHost string, events []Event) (canRetry bool, err error)
}
--------------------------------------------------------------------------------

### CODE SAMPLE #10
File: initialize-sdk-go.md
Section: Customize `EventApiManager`
Language: go
--------------------------------------------------------------------------------
eventDispatchTimeoutMillis := 20000 * time.Millisecond
defaultEventApiManager := event.NewEventAPIManager(sdkKey, utils.NewHTTPRequester(logging.GetLogger(sdkKey, "EventAPIManager"), utils.Timeout(eventDispatchTimeoutMillis)))
--------------------------------------------------------------------------------

### CODE SAMPLE #11
File: initialize-sdk-go.md
Section: Customize SegmentAPIManager
Language: go
--------------------------------------------------------------------------------
// APIManager represents the segment API manager.
type APIManager interface {
	FetchQualifiedSegments(apiKey, apiHost, userID string, segmentsToCheck []string) ([]string, error)
}
--------------------------------------------------------------------------------

### CODE SAMPLE #12
File: initialize-sdk-go.md
Section: Customize `ODPEventManager`
Language: go
--------------------------------------------------------------------------------
queueSize := 20000
// Note: if this is set to 0 then batchSize will be set to 1, otherwise batchSize will be default, which is 10.
flushIntervalInMillis := 10000 * time.Millisecond // 10,000 msecs = 10 secs 

odpEventManager := event.NewBatchEventManager(
	event.WithAPIManager(defaultEventApiManager),
	event.WithQueueSize(queueSize),
	event.WithFlushInterval(flushIntervalInMillis),
)is
--------------------------------------------------------------------------------

### CODE SAMPLE #13
File: initialize-sdk-go.md
Section: Customize `ODPSegmentManager`
Language: go
--------------------------------------------------------------------------------
cacheSize := 600
cacheTimeoutInSeconds := 600 * time.Second // 10 mins = 600 secs
odpSegmentManager := segment.NewSegmentManager(
	sdkKey,
	segment.WithAPIManager(defaultSegmentApiManager),
	segment.WithSegmentsCacheSize(cacheSize),
	segment.WithSegmentsCacheTimeout(cacheTimeoutInSeconds),
)

// Second method to set custom cache size and timeout using odpManager
odpManager := odp.NewOdpManager(
	sdkKey, false,
	odp.WithSegmentsCacheSize(cacheSize),
	odp.WithSegmentsCacheTimeout(cacheTimeoutInSeconds),
)
--------------------------------------------------------------------------------

### CODE SAMPLE #14
File: initialize-sdk-go.md
Section: Custom cache
Language: go
--------------------------------------------------------------------------------
// Cache is used for caching ODP segments
type Cache interface {
	Save(key string, value interface{})
	Lookup(key string) interface{}
	Reset()
}
--------------------------------------------------------------------------------

### CODE SAMPLE #15
File: initialize-sdk-go.md
Section: Custom cache
Language: go
--------------------------------------------------------------------------------
type CustomSegmentsCache struct {
	Cache map[string]interface{}
	lock  sync.Mutex
}

func (c *CustomSegmentsCache) Save(key string, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Cache[key] = value
}

func (c *CustomSegmentsCache) Lookup(key string) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.Cache[key]
}

func (c *CustomSegmentsCache) Reset() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Cache = map[string]interface{}{}
}
--------------------------------------------------------------------------------

### CODE SAMPLE #16
File: initialize-sdk-go.md
Section: Custom cache
Language: go
--------------------------------------------------------------------------------
customCache := &CustomSegmentsCache{
	Cache: map[string]interface{}{},
}

// Note: To use custom Cache user should not pass SegmentManager 
odpManager = odp.NewOdpManager(
	sdkKey, 
	false,
	odp.WithSegmentsCache(customCache),
)
--------------------------------------------------------------------------------

### CODE SAMPLE #17
File: decide-methods-go.md
Section: Example `decide`
Language: go
--------------------------------------------------------------------------------
package main

import (
	"fmt"

  optly "github.com/optimizely/go-sdk"      // for v2: "github.com/optimizely/go-sdk/v2"
  "github.com/optimizely/go-sdk/pkg/decide" // for v2: "github.com/optimizely/go-sdk/v2/pkg/decide"
)

func main() {
	optimizely_client, err := optly.Client("SDK_KEY_HERE") // Replace with your SDK key
	if err != nil {
		panic(err)
	}

	user := optimizely_client.CreateUserContext("user123", map[string]interface{}{"logged_in": true})
	decision := user.Decide("product_sort", []decide.OptimizelyDecideOptions{})

	// Did the decision fail with a critical error?
	if decision.VariationKey == "" {
		fmt.Printf("[decide] error: %v", decision.Reasons)
		return
	}

	// Flag enabled state
	enabled := decision.Enabled
	fmt.Println("Decision enabled: ", enabled)

	// String variable value
	var value1 string
	if err := decision.Variables.GetValue("sort_method", &value1); err != nil {
		panic(err)
	}
	// Or
	value2 := decision.Variables.ToMap()["sort_method"].(string)
	fmt.Println("Variable value: ", value2)

	// All variable values
	allVarValues := decision.Variables
	fmt.Println("All variable: ", allVarValues)

	// Variation key
	variationKey := decision.VariationKey
	fmt.Println("Variation Key: ", variationKey)

	// User for which the decision was made
	userContext := decision.UserContext

	// Flag decision reasons
	reasons := decision.Reasons
	fmt.Println("Decision reasons: ", reasons)
}
--------------------------------------------------------------------------------

### CODE SAMPLE #18
File: decide-methods-go.md
Section: Example `DecideAll`
Language: go
--------------------------------------------------------------------------------
// Make a decisions for all active (unarchived) flags for the user
decisions := user.DecideAll([]decide.OptimizelyDecideOptions{})
// Or only for enabled flags
decisions = user.DecideAll([]decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly})

flagKeys := []string{}
flagDecisions := []client.OptimizelyDecision{}
for k, v := range decisions {
  flagKeys = append(flagKeys, k)
  flagDecisions = append(flagDecisions, v)
}
decisionForFlag1 := decisions["flag_1"]
--------------------------------------------------------------------------------

### CODE SAMPLE #19
File: decide-methods-go.md
Section: Example `DecideForKeys`
Language: go
--------------------------------------------------------------------------------
decisions := user.DecideForKeys([]string{"flag_1", "flag_2"}, []decide.OptimizelyDecideOptions{decide.EnabledFlagsOnly})

decisionForFlag1 := decisions["flag_1"]
decisionForFlag2 := decisions["flag_2"]
--------------------------------------------------------------------------------

### CODE SAMPLE #20
File: decide-methods-go.md
Section: Manual cache control
Language: go
--------------------------------------------------------------------------------
import (
    optly "github.com/optimizely/go-sdk/v2"
    "github.com/optimizely/go-sdk/v2/pkg/decide"
)

optimizely_client, err := optly.Client("SDK_KEY_HERE")
if err != nil {
    panic(err)
}

// Example 1: Bypass cache for real-time decision
user := optimizely_client.CreateUserContext("user123",
map[string]interface{}{
    "age":      25,
    "location": "US",
})
decision := user.Decide("my-cmab-flag",
[]decide.OptimizelyDecideOptions{
    decide.IgnoreCMABCache, // Always fetch fresh from CMAB
service
})

// Example 2: Invalidate cache when user context changes
significantly
user.SetAttributes(map[string]interface{}{
    "age":      26,
    "location": "UK", // Context changed
})
decision = user.Decide("my-cmab-flag",
[]decide.OptimizelyDecideOptions{
    decide.InvalidateUserCMABCache, // Clear cached decision
for this user
})

// Example 3: Reset entire CMAB cache (use sparingly)
decision = user.Decide("my-cmab-flag",
[]decide.OptimizelyDecideOptions{
    decide.ResetCMABCache, // Clear all CMAB cache entries
})
--------------------------------------------------------------------------------

### CODE SAMPLE #21
File: decide-methods-go.md
Section: CMAB decision reasons
Language: go
--------------------------------------------------------------------------------
decision := user.Decide("my-cmab-flag",
[]decide.OptimizelyDecideOptions{
    decide.IncludeReasons,
})

// Print CMAB-related decision reasons
for _, reason := range decision.Reasons {
    fmt.Println(reason)
    // Examples:
    // "CMAB decision retrieved from cache."
    // "CMAB decision fetched from service."
    // "CMAB cache invalidated due to attribute change."
}
--------------------------------------------------------------------------------

### CODE SAMPLE #22
File: optimizelyusercontext-go.md
Section: OptimizelyUserContext definition
Language: go
--------------------------------------------------------------------------------
type OptimizelyUserContext struct {
  UserID     string                
  Attributes map[string]interface{}
}
 
// GetOptimizely returns optimizely client instance for Optimizely user context
func (o OptimizelyUserContext) GetOptimizely() *OptimizelyClient
 
// GetUserID returns userID for Optimizely user context
func (o OptimizelyUserContext) GetUserID() string
 
// GetUserAttributes returns user attributes for Optimizely user context
func (o OptimizelyUserContext) GetUserAttributes() map[string]interface{}
 
// SetAttribute sets an attribute for a given key.
func (o *OptimizelyUserContext) SetAttribute(key string, value interface{})
 
// Decide returns a decision result for a given flag key and a user context, which contains
// all data required to deliver the flag or experiment.
func (o *OptimizelyUserContext) Decide(key string, options []decide.OptimizelyDecideOptions) OptimizelyDecision
 
// DecideAll returns a key-map of decision results for all active flag keys with options.
func (o *OptimizelyUserContext) DecideAll(options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision
 
// DecideForKeys returns a key-map of decision results for multiple flag keys and options.
func (o *OptimizelyUserContext) DecideForKeys(keys []string, options []decide.OptimizelyDecideOptions) map[string]OptimizelyDecision
 
// TrackEvent generates a conversion event with the given event key if it exists and queues it up to be sent to the Optimizely
// log endpoint for results processing.
func (o *OptimizelyUserContext) TrackEvent(eventKey string, eventTags map[string]interface{}) (err error)
 
// OptimizelyDecisionContext
type OptimizelyDecisionContext struct {
  FlagKey string
  RuleKey string
}
 
// OptimizelyForcedDecision
type OptimizelyForcedDecision struct {
  VariationKey string
}
 
// SetForcedDecision sets the forced decision (variation key) for a given decision context (flag key and optional rule key).
// returns true if the forced decision has been set successfully.
func (o *OptimizelyUserContext) SetForcedDecision(context pkgDecision.OptimizelyDecisionContext, decision pkgDecision.OptimizelyForcedDecision) bool
 
// GetForcedDecision returns the forced decision for a given flag and an optional rule
func (o *OptimizelyUserContext) GetForcedDecision(context pkgDecision.OptimizelyDecisionContext) (pkgDecision.OptimizelyForcedDecision, error)
 
// RemoveForcedDecision removes the forced decision for a given flag and an optional rule.
func (o *OptimizelyUserContext) RemoveForcedDecision(context pkgDecision.OptimizelyDecisionContext) bool
 
// RemoveAllForcedDecisions removes all forced decisions bound to this user context.
func (o *OptimizelyUserContext) RemoveAllForcedDecisions() bool
 
//
// The following methods require Real-Time Audiences for Feature Experimentation. 
// See note following this code sample. 
//
 
// GetQualifiedSegments returns an array of segment names that the user is qualified for. 
// The result of **FetchQualifiedSegments()** is saved here. 
// Can be nil if not properly updated with FetchQualifiedSegments(). 
func (o *OptimizelyUserContext) GetQualifiedSegments() []string
 
// SetQualifiedSegments can read and write directly to the qualified segments array. 
// This lets you bypass the remote fetching process from ODP 
// or for utilizing your own fetching service.   
func (o *OptimizelyUserContext) SetQualifiedSegments(qualifiedSegments []string)
      
// FetchQualifiedSegments fetches all qualified segments for the user context.
// The segments fetched are saved in the **qualifiedSegments** array
// and can be accessed any time. 
func (o *OptimizelyUserContext) FetchQualifiedSegments(options []pkgOdpSegment.OptimizelySegmentOption) (success bool)
 
// FetchQualifiedSegmentsAsync fetches all qualified segments aysnchronously for the user context. 
// This method fetches segments in a separate go routine and invoke the provided 
// callback when results are available.
func (o *OptimizelyUserContext) FetchQualifiedSegmentsAsync(options []pkgOdpSegment.OptimizelySegmentOption, callback func(success bool))
 
// IsQualifiedFor returns true if the user is qualified for the given segment name
func (o *OptimizelyUserContext) IsQualifiedFor(segment string) bool
--------------------------------------------------------------------------------

### CODE SAMPLE #23
File: event-batching-go.md
Section: Basic example
Language: go
--------------------------------------------------------------------------------
import optly "github.com/optimizely/go-sdk" // for v2: "github.com/optimizely/go-sdk/v2"

// the default client will have a BatchEventProcessor with the default options
optlyClient, err := optly.Client("SDK_KEY_HERE")
--------------------------------------------------------------------------------

### CODE SAMPLE #24
File: event-batching-go.md
Section: Advanced Example
Language: go
--------------------------------------------------------------------------------
import (
  "time"
  
  "github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client"
  "github.com/optimizely/go-sdk/pkg/event"  // for v2: "github.com/optimizely/go-sdk/v2/pkg/event"
  "github.com/optimizely/go-sdk/pkg/utils"  // for v2: "github.com/optimizely/go-sdk/v2/pkg/utils"
)

optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "SDK_KEY",	
}

// You can configure the batch size and flush interval
eventProcessor := event.NewBatchEventProcessor(
  event.WithBatchSize(10), 
  event.WithFlushInterval(30 * time.Second),
)
optlyClient, err := optimizelyFactory.Client(
  client.WithEventProcessor(eventProcessor),
)
--------------------------------------------------------------------------------

### CODE SAMPLE #25
File: event-batching-go.md
Section: Register and Unregister a `LogEvent` listener
Language: go
--------------------------------------------------------------------------------
import (
	"fmt"

  "github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client"
  "github.com/optimizely/go-sdk/pkg/event"  // for v2: "github.com/optimizely/go-sdk/v2/pkg/event"
)

// Callback for log event notification
callback := func(notification event.LogEvent) {

  // URL to dispatch log event to
  fmt.Print(notification.EndPoint)
  // Batched event
  fmt.Print(notification.Event)
}

optimizelyFactory := &client.OptimizelyFactory{
  SDKKey: "SDK_KEY",
}
optimizelyClient, err := optimizelyFactory.Client()
if err != nil {
	// handle error
}

// Add callback for logEvent notification
id, err := optimizelyClient.EventProcessor.(*event.BatchEventProcessor).OnEventDispatch(callback)
if err != nil {
	// handle error
}
// Remove callback for logEvent notification
err = optimizelyClient.EventProcessor.(*event.BatchEventProcessor).RemoveOnEventDispatch(id)
--------------------------------------------------------------------------------

### CODE SAMPLE #26
File: set-up-notification-listener-go.md
Section: Set up each type of notification listener
Language: go
--------------------------------------------------------------------------------
// Create default Client
optimizelyFactory := &client.OptimizelyFactory{
  SDKKey: "[SDK_KEY_HERE]",
}
optimizelyClient, err := optimizelyFactory.Client()
if err != nil {
	// handle error
}


// SET UP DECISION NOTIFICATION LISTENER

onDecision := func(notification notification.DecisionNotification) {
  // Add a DECISION Notification Listener for type FLAG
  if string(notification.Type) == "flag" {
    // Access information about feature, for example, key and enabled status
    fmt.Print(notification.DecisionInfo["flagKey"])
    fmt.Print(notification.DecisionInfo["enabled"])
    fmt.Print(notification.DecisionInfo["decisionEventDispatched"])
  }
}
notificationID, err := optimizelyClient.DecisionService.OnDecision(onDecision)
if err != nil {
	// handle error
}

// REMOVE DECISION NOTIFICATION LISTENER

optimizelyClient.DecisionService.RemoveOnDecision(notificationID)

// SET UP LOG EVENT NOTIFICATION LISTENER

onLogEvent := func(eventNotification event.LogEvent) {
  // process the logEvent object here (send to analytics provider, audit/inspect data)
}
notificationID, err = optimizelyClient.EventProcessor.OnEventDispatch(onLogEvent)
if err != nil {
	// handle error
}

// REMOVE LOG EVENT NOTIFICATION LISTENER

optimizelyClient.EventProcessor.RemoveOnEventDispatch(notificationID)

// SET UP OPTIMIZELY CONFIG NOTIFICATION LISTENER

// listen to OPTIMIZELY_CONFIG_UPDATE to get updated data
// You will get notifications whenever the datafile is updated except for SDK initialization
// you will get notifications whenever the datafile is updated, except for the SDK initialization
onConfigUpdate := func(notification notification.ProjectConfigUpdateNotification) {
}
notificationID, err = optimizelyClient.ConfigManager.OnProjectConfigUpdate(onConfigUpdate)
if err != nil {
	// handle error
}

// REMOVE OPTIMIZELY CONFIG NOTIFICATION LISTENER

optimizelyClient.ConfigManager.RemoveOnProjectConfigUpdate(notificationID)

// SET UP TRACK LISTENER

onTrack := func(eventKey string, userContext entities.UserContext, eventTags map[string]interface{}, conversionEvent event.ConversionEvent) {
  // process the event here (send to analytics provider, audit/inspect data)
}

notificationID, err = optimizelyClient.OnTrack(onTrack)
if err != nil {
	// handle error
}

// REMOVE TRACK LISTENER

optimizelyClient.RemoveOnTrack(notificationID)
--------------------------------------------------------------------------------

### CODE SAMPLE #27
File: example-usage-go.md
Section: Example usage of the Go SDK
Language: go
--------------------------------------------------------------------------------
import optly "github.com/optimizely/go-sdk" // for v2: "github.com/optimizely/go-sdk/v2"

optimizely_client, err := optly.Client("SDK_KEY_HERE")
if err != nil {
	// handle the err
}
// create a user and decide a flag rule (such as an A/B test) for them
user := optimizely_client.CreateUserContext("user123", map[string]interface{}{"logged_in": true})
decision := user.Decide("product_sort", []decide.OptimizelyDecideOptions{})

var variationKey string
if variationKey = decision.VariationKey; variationKey == "" {
  fmt.Printf("[decide] error: %v", decision.Reasons)
  return
}

// execute code based on flag enabled state
enabled := decision.Enabled

if enabled {
  // get flag variable values
  var value1 string
  decision.Variables.GetValue("sort_method", &value1)
  // or:
  value2 := decision.Variables.ToMap()["sort_method"].(string)
}

// or execute code based on flag variation:
if variationKey == "control" {
  // Execute code for control variation
} else if variationKey == "treatment" {
  // Execute code for treatment variation
}

// Track a user event
user.TrackEvent("purchased", nil)
--------------------------------------------------------------------------------

### CODE SAMPLE #28
File: track-event-go.md
Section: Example
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/client" 
// for v2: "github.com/optimizely/go-sdk/v2/pkg/client"

optimizelyFactory := client.OptimizelyFactory{SDKKey: "sdk-key"}
client, err := optimizelyFactory.StaticClient()
if err != nil {
	// handle error here
}
user := client.CreateUserContext("user123", map[string]interface{}{"logged_in": true})

// event properties 
properties := map[string]interface{}{
	"category": "shoes",
  "color": "red",
}

tags := map[string]interface{}{
  "revenue": 10000, 
  "value": 100.00,
  "$opt_event_properties": properties,
}
//user.TrackEvent(eventkey (required), eventTag (optional)
if err := user.TrackEvent("my_purchase_event_key", tags); err != nil {
		// handle error here
}
--------------------------------------------------------------------------------

### CODE SAMPLE #29
File: configure-event-dispatcher-go.md
Section: Configure the Go SDK event dispatcher
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/event" // for v2: "github.com/optimizely/go-sdk/v2/pkg/event"

type CustomEventDispatcher struct {
}

// DispatchEvent dispatches event with callback
func (d *CustomEventDispatcher) DispatchEvent(event event.LogEvent) (bool, error) {
	dispatchedEvent := map[string]interface{}{
		"url":       event.EndPoint,
		"http_verb": "POST",
		"headers":   map[string]string{"Content-Type": "application/json"},
		"params":    event.Event,
	}
	return true, nil
}
--------------------------------------------------------------------------------

### CODE SAMPLE #30
File: configure-event-dispatcher-go.md
Section: Configure the Go SDK event dispatcher
Language: go
--------------------------------------------------------------------------------
import (
	"github.com/optimizely/go-sdk/pkg/client" // for v2: "github.com/optimizely/go-sdk/v2/pkg/client"
  "github.com/optimizely/go-sdk/pkg/event"  // for v2: "github.com/optimizely/go-sdk/v2/pkg/event"
)

optimizelyFactory := &client.OptimizelyFactory{
  SDKKey: "SDK_KEY_HERE",
}

customEventDispatcher := &CustomEventDispatcher{}

// Create an Optimizely client with the custom event dispatcher
optlyClient, e := optimizelyFactory.Client(client.WithEventDispatcher(customEventDispatcher))
--------------------------------------------------------------------------------

### CODE SAMPLE #31
File: customize-logger-go.md
Section: Custom logger implementation in the SDK
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/logging" // for v2: "github.com/optimizely/go-sdk/v2/pkg/logging"

type CustomLogger struct {
}

func (l *CustomLogger) Log(level logging.LogLevel, message string, fields map[string]interface{}) {
}

func (l *CustomLogger) SetLogLevel(level logging.LogLevel) {
}

customLogger := New(CustomLogger)

logging.SetLogger(customLogger)
--------------------------------------------------------------------------------

### CODE SAMPLE #32
File: customize-logger-go.md
Section: Setting the log level
Language: go
--------------------------------------------------------------------------------
import "github.com/optimizely/go-sdk/pkg/logging" // for v2: "github.com/optimizely/go-sdk/v2/pkg/logging"

// Set log level to Debug
logging.SetLogLevel(logging.LogLevelDebug)
--------------------------------------------------------------------------------

### CODE SAMPLE #33
File: get-forced-variation-go.md
Section: Example
Language: go
--------------------------------------------------------------------------------
overrideStore := decision.NewMapExperimentOverridesStore()
client, err := optimizelyFactory.Client(
       client.WithExperimentOverrides(overrideStore),
)
--------------------------------------------------------------------------------

### CODE SAMPLE #34
File: get-forced-variation-go.md
Section: Example
Language: go
--------------------------------------------------------------------------------
overrideKey := decision.ExperimentOverrideKey{ExperimentKey: "test_experiment", UserID: "test_user"}
variation, success := overrideStore.GetVariation(overrideKey)
--------------------------------------------------------------------------------
