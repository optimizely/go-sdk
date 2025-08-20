# CMAB Testing Guide for Optimizely Go SDK

## Overview
This guide provides comprehensive test scenarios and example code for testing CMAB (Contextual Multi-Armed Bandit) functionality in the Optimizely Go SDK.

## Test Environment Setup

### Prerequisites
1. Go 1.18+ installed
2. Access to Optimizely develrc environment
3. Go SDK with CMAB support

### SDK Configuration
- **SDK Key**: `JgzFaGzGXx6F1ocTbMTmn` (TODO: provide own SDK KEY for each?)
- **Environment**: develrc (dev.cdn.optimizely.com)
- **Test Flag**: `flag-matjaz-editor`  (TODO: will be own flag, testers cerate own flag in project?)

## Running the Tests

### Basic Usage
```bash
# Run all tests
go run cmab_test_example.go

# Run specific test
go run cmab_test_example.go -test=cache_hit

# Available test cases:
# - basic         : Basic CMAB functionality
# - cache_hit     : Cache hit scenarios
# - cache_miss    : Cache miss on attribute changes
# - ignore_cache  : IGNORE_CMAB_CACHE option
# - reset_cache   : RESET_CMAB_CACHE option
# - invalidate_user : INVALIDATE_USER_CMAB_CACHE option
# - concurrent    : Concurrent request handling
# - error         : Error handling scenarios
# - all           : Run all tests
```

### With Profiling (Optional)
```bash
# CPU profiling
go build -ldflags "-X main.RunCPUProfile=true" cmab_test_example.go && ./cmab_test_example

# Memory profiling
go build -ldflags "-X main.RunMemProfile=true" cmab_test_example.go && ./cmab_test_example
```

## Test Scenarios

### 1. Core Functionality Tests

#### Basic CMAB Decision
- Tests user qualification for CMAB experiments
- Verifies fallthrough when audience conditions aren't met
- Validates variation assignment from CMAB service

#### Traffic Allocation
- Tests users not bucketed into experiment traffic
- Verifies proper fallthrough to next experiment

### 2. Caching Tests

#### Cache Hit Scenarios
- Same user with identical attributes returns cached decision
- Non-relevant attribute changes still use cache
- Validates CMAB UUID consistency

#### Cache Miss Scenarios
- Relevant attribute changes trigger new CMAB call
- New CMAB UUID generated for different attribute combinations

#### Cache Control Options
- **IGNORE_CMAB_CACHE**: Forces fresh prediction without affecting cache
- **RESET_CMAB_CACHE**: Clears entire CMAB cache for all users
- **INVALIDATE_USER_CMAB_CACHE**: Clears cache for specific user only

### 3. Advanced Tests

#### Concurrent Requests
- Multiple simultaneous decisions for same user
- Validates single CMAB call and consistent results

#### Error Handling
- CMAB service errors (500, timeout)
- Invalid response handling
- Network failures

## Expected Behaviors

### Decision Response Structure
```go
type OptimizelyDecision struct {
    VariationKey string              // Assigned variation
    Enabled      bool                // Feature flag state
    RuleKey      string              // Experiment/rule key
    Variables    map[string]interface{} // Feature variables
    Reasons      []string            // Decision reasons
    // CMAB metadata in event logs
}
```

### CMAB-Specific Indicators
1. **Debug Logs**: Look for CMAB/prediction endpoint calls
2. **Decision Reasons**: May include CMAB-related messages
3. **Event Metadata**: CMAB UUID in dispatched events

## Debugging Tips

### Enable Debug Logging
```go
logging.SetLogLevel(logging.LogLevelDebug)
```

### Key Log Messages to Watch
- "Fetching CMAB prediction for experiment..."
- "CMAB cache hit for user..."
- "CMAB cache miss, fetching new prediction..."
- "CMAB service error..."

### Common Issues and Solutions

| Issue | Possible Cause | Solution |
|-------|---------------|----------|
| No CMAB calls | User not qualified | Check audience conditions |
| Always cache miss | Attribute types incorrect | Ensure consistent types |
| No variation returned | CMAB service error | Check service availability |
| Inconsistent results | Race condition | Use proper synchronization |

## Integration with FSC Tests

The test scenarios align with the Fullstack Compatibility Suite (FSC) feature tests:
- Audience qualification (`cmab_decision.feature:14-89`)
- Traffic allocation (`cmab_decision.feature:91-167`)
- Successful predictions (`cmab_decision.feature:169-258`)
- Cache behavior (`cmab_decision.feature:260-534`)
- Cache options (`cmab_decision.feature:536-869`)
- Error handling (`cmab_decision.feature:871-1013`)

## Extending the Tests

### Adding New Test Cases
1. Add new test function following the pattern:
```go
func testNewScenario(optimizelyClient *client.OptimizelyClient) {
    fmt.Println("\n--- Test: Your Test Name ---")
    // Test implementation
}
```

2. Add to switch statement in main()
3. Include in runAllTests() if applicable

### Custom Attributes
Modify the attribute maps to test your specific use cases:
```go
attributes := map[string]interface{}{
    "category": "cmab",
    "age":      30,
    "country":  "US",
    // Add your custom attributes
    "subscription_level": "premium",
    "total_purchases": 150,
}
```

## Notes for Engineers

### Key Implementation Points
1. **Cache Key Generation**: Based on user ID + relevant attributes
2. **Attribute Filtering**: Only CMAB-configured attributes affect cache
3. **UUID Tracking**: Each CMAB decision has unique UUID for tracking
4. **Async Behavior**: CMAB calls should be non-blocking where possible

### Performance Considerations
- Cache reduces CMAB service calls significantly
- TTL configuration balances freshness vs. performance
- Batch multiple decisions when possible

### Testing Checklist
- [ ] Basic CMAB qualification and bucketing
- [ ] All cache scenarios (hit, miss, control options)
- [ ] Concurrent request handling
- [ ] Error scenarios and fallbacks
- [ ] Event tracking with CMAB metadata
- [ ] Performance under load
- [ ] TTL expiration behavior

## Support
For issues or questions:
- Review debug logs for detailed CMAB behavior
- Check CMAB service health and connectivity
- Verify datafile has CMAB experiments configured
- Ensure SDK version supports CMAB features