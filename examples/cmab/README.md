# CMAB Testing Guide for Optimizely Go SDK

## Test Environment Setup

### Prerequisites
1. Go 1.18+ installed
2. Access to Optimizely RC (Prep) environment
3. Go SDK with CMAB support

### Project Configuration Required
**IMPORTANT**: Testers must create their own project:

1. **Create Project**: Set up a new project in Optimizely RC (Prep) environment
2. **Create CMAB Experiment**: Add a CMAB-enabled flag experiment
3. **Configure Attributes**: Set up exactly 2 custom attributes in OR condition:
   - `cmab_test_attribute` with values: `"hello"` OR `"world"`
   - This is required for cache miss tests to work properly
4. **Update Code**: Replace the SDK key and flag key in `main.go`:
   ```go
   SDK_KEY  = "YOUR_RC_SDK_KEY_HERE"    // Replace with your RC SDK key
   FLAG_KEY = "YOUR_CMAB_FLAG_KEY_HERE" // Replace with your CMAB flag key
   ```

### Environment Details
- **Environment**: RC (Prep) - `https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json`
- **CMAB Endpoint**: `https://prep.prediction.cmab.optimizely.com/`

## Running the Tests

### Basic Usage
```bash
# Run individual tests (from go-sdk root)
go run examples/cmab/main.go -test=basic
go run examples/cmab/main.go -test=cache_hit
go run examples/cmab/main.go -test=performance

# Available test cases:
# - basic            : Basic CMAB functionality
# - cache_hit        : Cache hit scenarios  
# - cache_miss       : Cache miss on attribute changes
# - ignore_cache     : IGNORE_CMAB_CACHE option
# - reset_cache      : RESET_CMAB_CACHE option
# - invalidate_user  : INVALIDATE_USER_CMAB_CACHE option
# - concurrent       : Concurrent request handling (has known race condition bug)
# - error            : Error handling scenarios
# - fallback         : Fallback when not qualified
# - traffic          : Traffic allocation (requires 50% allocation setup)
# - forced           : Forced variation override
# - event_tracking   : Event tracking with CMAB UUID
# - attribute_types  : Attribute validation
# - performance      : Performance benchmarks
# - cache_expiry     : Cache TTL behavior
```

## Test Scenarios Overview

### Core Tests (1-6)
- **Basic**: CMAB qualified vs non-qualified users
- **Cache Hit**: Same user/attributes returns cached result
- **Cache Miss**: Different attribute values trigger new API calls
- **Ignore Cache**: Bypasses cache without affecting it
- **Reset Cache**: Clears entire CMAB cache
- **Invalidate User**: Clears cache for specific user only

### Advanced Tests (7-14)  
- **Concurrent**: Thread safety (⚠️ **Known Issue**: Race condition bug)
- **Error Handling**: Invalid attribute types and graceful fallback
- **Fallback**: Users without required attributes
- **Traffic Allocation**: Requires 50% traffic allocation setup
- **Forced Variations**: Override behavior (if configured)
- **Event Tracking**: CMAB UUID in analytics events
- **Attribute Types**: Missing attribute validation
- **Performance**: API vs cache timing benchmarks
- **Cache Expiry**: TTL-based invalidation

## Testing Guidelines

### ⚠️ Important Notes for Testers
1. **Test 9 (Traffic)**: Temporarily set CMAB experiment traffic allocation to 50% in Optimizely UI for this test only. Keep at 100% for all other tests.

2. **Test 6 (Concurrent)**: Currently has a known race condition bug where concurrent requests return different variations instead of consistent results. Document this behavior.

3. **Beyond Basic Tests**: These test cases are **guidelines only**. In testing try:
   - Try different attribute combinations
   - Test various traffic allocation percentages  
   - Create different audience conditions
   - Test edge cases and error scenarios
   - Experiment with different caching scenarios
   - Validate behavior under different load conditions

### Expected Test Results
- **CMAB API calls**: Look for `"Fetching CMAB decision"` in debug logs
- **Cache hits**: Look for `"Returning cached CMAB decision"` in debug logs
- **Consistent variations**: Same user should get same variation (except during concurrent test bug)
- **Graceful fallback**: Invalid scenarios should fall back to rollout experiments

## Debugging and Validation

### Enable Debug Logging
```go
logging.SetLogLevel(logging.LogLevelDebug)
```

### Key Debug Log Patterns
- `"Fetching CMAB decision for rule X and user Y"` - New API call
- `"Returning cached CMAB decision for rule X and user Y"` - Cache hit
- `"CMAB request body: {...}"` - Actual API request payload
- `"CMAB raw response: {...}"` - API response
- `"User X not in audience for CMAB experiment Y"` - Audience failure

### Validation Checklist
- [ ] CMAB-qualified users get CMAB API calls
- [ ] Non-qualified users fall back to rollout (no CMAB calls)
- [ ] Cache hits return same variation without API calls
- [ ] Cache misses trigger new API calls with different attribute values
- [ ] Cache control options work as expected
- [ ] Performance benchmarks show cache is significantly faster than API calls
- [ ] Error scenarios handle gracefully with appropriate fallback behavior

## Integration Notes

### Decision Response Structure
```go
type OptimizelyDecision struct {
    VariationKey string                 // "on"/"off" variation
    Enabled      bool                   // Feature enabled state
    RuleKey      string                 // CMAB experiment key
    Variables    map[string]interface{} // Feature variables
    Reasons      []string               // Decision reasons
}
```

### CMAB-Specific Metadata
- **CMAB UUID**: Unique identifier in request/response for tracking
- **Event Integration**: CMAB UUID included in impression/conversion events
- **Cache Keys**: Generated from user ID + relevant attribute values

## Known Issues
1. **Concurrent Test (Test 6)**: Race condition causes inconsistent variations across concurrent requests for the same user - currently working on this
2. **Traffic Test (Test 9)**: Requires manual configuration of traffic allocation percentage
3. **Event Tracking**: "purchase" conversion event may need to be configured in Optimizely UI