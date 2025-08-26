# CMAB Go SDK Bug Bash Test Cases
## Comprehensive SDK Test Suite with DS Validation

*This file contains all tests needed for CMAB bug bash testing. Tests marked with 🔬 also validate Data Science team requirements.*

---

## Test 1: Basic CMAB Decision Flow 🔬
**Purpose**: Verify SDK correctly handles CMAB experiments  
**Bug Bash**: Basic functionality, API integration  
**DS Requirements**: API request format, UUID generation

**Steps**:
1. Initialize SDK with CMAB-enabled datafile
2. Create user "test_user_99" with attributes: {category: "cmab", age: 50, country: "BD"}
3. Call `userContext.Decide("flag_with_cmab")`
4. Verify decision comes from CMAB experiment
5. Check debug logs for CMAB API call
6. **DS Validation**: Capture API request and verify format

**Expected**:
- Variation returned from CMAB prediction API
- CMAB UUID generated (valid v4 format)
- Decision includes rule_key: "exp_1"
- **DS Check**: API request contains:
  ```json
  {
    "instances": [{
      "visitorId": "test_user_99",
      "experimentId": "2001",
      "cmabUUID": "<valid-uuid-v4>",
      "attributes": [
        {"id": "age", "type": "custom_attribute", "value": 50},
        {"id": "country", "type": "custom_attribute", "value": "BD"}
      ]
    }]
  }
  ```

**DS Validation Points**:
- ✅ SDK sends request with userId, experimentId, attributes, cmabUUID
- ✅ CMAB UUID creation (v4 format)
- ✅ Proper attribute formatting

---

## Test 2: Cache Hit Scenario 🔬
**Purpose**: Verify SDK caches CMAB decisions  
**Bug Bash**: Cache functionality works correctly  
**DS Requirements**: UUID persistence, caching algorithm

**Steps**:
1. Create user "cache_user" with attributes: {category: "cmab", age: 25}
2. Call decide() - first call
3. Note the returned variation and UUID
4. Call decide() again - second call  
5. Compare variations and UUIDs
6. **DS Validation**: Verify cache key format in logs

**Expected**:
- Same variation returned both times
- First call: CMAB API request in logs
- Second call: No API request (cache hit)
- Same CMAB UUID in both decisions
- **DS Check**: Cache key format should be "10-cache_user-2001" (length-userId-experimentId)

**DS Validation Points**:
- ✅ SDK implements caching to reduce API calls
- ✅ UUID persistence for cached decisions
- ✅ Cache key algorithm implementation

---

## Test 3: Cache Miss on Attribute Change 🔬
**Purpose**: Verify cache invalidation when attributes change  
**Bug Bash**: Cache invalidation works correctly  
**DS Requirements**: Session management, attribute filtering

**Steps**:
1. User "test_user" with {category: "cmab", age: 30, country: "US"}
2. Call decide() - returns variation A, note UUID
3. Update user with {category: "cmab", age: 31, country: "US"}
4. Call decide() - should make new API call, note UUID
5. Add non-relevant attribute {language: "EN"}
6. Call decide() - should use cache
7. **DS Validation**: Verify only relevant attributes in API requests

**Expected**:
- New API call when age changes (relevant attribute)
- Cache hit when language added (non-relevant)
- Different CMAB UUID after attribute change (steps 2 vs 4)
- Same CMAB UUID when non-relevant attribute added (steps 4 vs 6)
- **DS Check**: API requests only contain age, country (not language)

**DS Validation Points**:
- ✅ Attribute filtering (only relevant attributes affect cache)
- ✅ Session management for attribute changes
- ✅ MurmurHash3 attribute hashing for cache invalidation

---

## Test 4: IGNORE_CMAB_CACHE Option
**Purpose**: Verify cache bypass option  

**Steps**:
```go
// First call - populate cache
decision1 := userContext.Decide("flag_with_cmab", nil)

// Second call - ignore cache
options := []decide.OptimizelyDecideOptions{decide.IgnoreCMABCache}
decision2 := userContext.Decide("flag_with_cmab", options)

// Third call - normal (should use original cache)
decision3 := userContext.Decide("flag_with_cmab", nil)
```

**Expected**:
- Second call makes new API request despite cache
- New CMAB UUID for second call
- Third call uses original cached decision

**Validates DS Requirements**:
- ✅ Cache control for testing/debugging

---

## Test 5: RESET_CMAB_CACHE Option  
**Purpose**: Verify global cache reset  

**Steps**:
1. Populate cache for users A, B, C
2. User A calls decide() with `ResetCMABCache` option
3. All users call decide() again

**Expected**:
- All users make new API calls after reset
- New CMAB UUIDs for all users
- Complete cache cleared

**Validates DS Requirements**:
- ✅ Cache management capabilities

---

## Test 6: INVALIDATE_USER_CMAB_CACHE Option
**Purpose**: Verify per-user cache invalidation  

**Steps**:
1. Populate cache for users A and B
2. User A calls decide() with `InvalidateUserCMABCache`
3. Both users call decide() again

**Expected**:
- User A: New API call, new UUID
- User B: Cache hit, same UUID
- Only specific user's cache cleared

**Validates DS Requirements**:
- ✅ User-specific cache control

---

## Test 7: Decision Event Tracking 🔬
**Purpose**: Verify CMAB UUID in events  
**Bug Bash**: Event tracking functionality  
**DS Requirements**: UUID in event metadata

**Steps**:
1. Configure event processor with batch_size: 1
2. Make CMAB decision, note UUID
3. Capture dispatched impression event
4. Track a conversion: `userContext.TrackEvent("purchase")`
5. Capture conversion event
6. **DS Validation**: Verify UUID consistency across events

**Expected Event Structure**:
```json
{
  "visitors": [{
    "snapshots": [{
      "decisions": [{
        "experiment_id": "2001",
        "variation_id": "5002",
        "metadata": {
          "flag_key": "flag_with_cmab",
          "rule_key": "exp_1",
          "cmab_uuid": "550e8400-e29b-41d4-a716-446655440000"
        }
      }]
    }]
  }]
}
```

**Expected**:
- Both impression and conversion events contain same CMAB UUID
- UUID matches the one from decision response
- **DS Check**: Events linkable via UUID for model training

**DS Validation Points**:
- ✅ SDK adds CMAB UUID to event payload
- ✅ Events trackable through pipeline via UUID
- ✅ UUID consistency across impression and conversion events

---

## Test 8: Fallback When Not Qualified
**Purpose**: Verify behavior when user doesn't qualify  

**Steps**:
1. User with {category: "not-cmab"} (fails audience condition)
2. Call decide("flag_with_cmab")
3. Check logs for CMAB activity

**Expected**:
- No CMAB API call
- Decision from next experiment or rollout
- No CMAB UUID in decision
- Falls through to exp_2 or rollout

**Validates DS Requirements**:
- ✅ Efficient API usage (no unnecessary calls)

---

## Test 9: Traffic Allocation Check
**Purpose**: Verify user bucketing into CMAB traffic  

**Steps**:
1. User "test_user_1" (not in traffic allocation)
2. Call decide() 
3. User "test_user_99" (in traffic allocation)
4. Call decide()

**Expected**:
- test_user_1: No CMAB call, falls through
- test_user_99: CMAB API called
- Traffic allocation respected

**Validates DS Requirements**:
- ✅ Standard bucketing before CMAB call

---

## Test 10: API Error Handling
**Purpose**: Verify SDK handles API failures gracefully  

**Steps**:
1. Simulate API timeout/500 error
2. Call decide()
3. Check decision and reasons

**Expected**:
- SDK retries 3 times (check logs)
- Returns null/false decision
- Reason: "Failed to fetch CMAB data for experiment exp_1"
- No events dispatched

**Validates DS Requirements**:
- ✅ Error handling with retry logic
- ✅ Graceful degradation

---

## Test 11: Concurrent Requests
**Purpose**: Verify thread-safe cache handling  

**Steps**:
```go
// Launch 10 concurrent decide calls
for i := 0; i < 10; i++ {
    go func() {
        decision := userContext.Decide("flag_with_cmab", nil)
    }()
}
```

**Expected**:
- Only 1 CMAB API call
- All return same variation
- Same CMAB UUID for all
- No race conditions

**Validates DS Requirements**:
- ✅ Performance optimization
- ✅ Thread safety

---

## Test 12: Forced Variation Override
**Purpose**: Verify forced variations bypass CMAB  

**Steps**:
1. Set forced variation for user
2. Call decide() for CMAB flag
3. Check logs

**Expected**:
- Forced variation returned
- No CMAB API call
- No CMAB UUID in decision

**Validates DS Requirements**:
- ✅ Standard SDK override behavior maintained

---

## Test 13: Cache Expiry (Time-based)
**Purpose**: Verify 30-minute cache TTL  

**Steps**:
1. Make CMAB decision at T=0
2. Wait 29 minutes, call decide()
3. Wait 2 more minutes (T=31), call decide()

**Expected**:
- T=29min: Cache hit, same UUID
- T=31min: Cache miss, new API call, new UUID

**Validates DS Requirements**:
- ✅ Session timeout management (30-min default)

---

## Test 14: Attribute Types in API Request
**Purpose**: Verify correct attribute type classification  

**Steps**:
1. Create user with various attribute types:
   - Custom: age=50, country="US"
   - Standard: $opt_bot_filtering=false
2. Call decide() and capture API request

**Expected Request Format**:
```json
{
  "instances": [{
    "attributes": [
      {"id": "age", "type": "custom_attribute", "value": 50},
      {"id": "country", "type": "custom_attribute", "value": "US"},
      {"id": "$opt_bot_filtering", "type": "standard_attribute", "value": false}
    ]
  }]
}
```

**Validates DS Requirements**:
- ✅ Proper attribute formatting for DS model

---

## Test 15: Performance Benchmarks
**Purpose**: Measure SDK latency  

**Steps**:
1. Measure cached decision time (100 calls)
2. Measure API decision time (10 calls with cache reset)
3. Calculate averages

**Expected Performance**:
- Cached decision: <10ms
- API decision: <500ms  
- With retry (simulated failure): <2000ms

**Validates DS Requirements**:
- ✅ Latency measurement
- ✅ Performance monitoring

---

## Quick Test Execution Guide

### Minimal Test Set (Core Functionality)
1. Test 1: Basic CMAB flow
2. Test 2: Cache hit
3. Test 7: Event tracking
4. Test 10: Error handling

### Full Test Set Order
1. **Setup**: Tests 1, 8, 9 (Basic flow)
2. **Caching**: Tests 2, 3, 4, 5, 6, 13 
3. **Events**: Test 7
4. **Errors**: Test 10
5. **Performance**: Tests 11, 15
6. **Overrides**: Test 12
7. **API Format**: Test 14

### Required Test Data
```go
// Users
USER_QUALIFIED := "test_user_99"     // In traffic & audience
USER_NOT_QUALIFIED := "test_user_1"  // Not in traffic
USER_CACHE_TEST := "cache_user_123"  

// Attributes
CMAB_ATTRIBUTES := map[string]interface{}{
    "category": "cmab",
    "age": 50,
    "country": "BD",
}
```

---

---

# ADDITIONAL DATA SCIENCE SPECIFIC TESTS 🔬
## Implementation Compliance Tests for DS Team

*These tests validate specific technical implementation details that DS team requires but don't naturally fit into SDK functionality testing.*

---

## DS Test A: 3x Retry Mechanism Validation
**Purpose**: Verify exactly 3 retry attempts on API failure  
**DS Requirement**: "In case of error with CMAB endpoint, retry 3x"  
**Why Separate**: Bug bash tests general error handling; DS needs exact retry count

**Steps**:
1. Mock CMAB API to fail twice, succeed on third attempt
2. Call decide() and count retry attempts in debug logs
3. Mock API to fail all 3 times, verify final failure behavior
4. Verify total request count in both scenarios

**Expected**:
- Scenario 1: Exactly 3 API calls, then success
- Scenario 2: Exactly 3 API calls, then final failure with error message
- No more than 3 attempts in any failure scenario

**DS Validation**: Reliable API communication with predictable retry behavior

---

## DS Test B: Cache Key Algorithm Validation
**Purpose**: Verify cache key generation follows TDD specification  
**DS Requirement**: Cache key format `length(userId) + '-' + userId + '-' + experimentId`  
**Why Separate**: Bug bash tests cache works; DS needs specific key format

**Steps**:
1. Create users with different ID lengths:
   - User "test123" (7 chars) with experiment "2001"
   - User "a" (1 char) with experiment "2001"  
   - User "very_long_username" (18 chars) with experiment "2001"
2. Enable detailed cache logging
3. Make decisions and capture cache key formats from logs

**Expected Cache Keys**:
- User "test123": `"7-test123-2001"`
- User "a": `"1-a-2001"`  
- User "very_long_username": `"18-very_long_username-2001"`

**DS Validation**: Consistent cache key generation for debugging/monitoring

---

## DS Test C: MurmurHash3 Attribute Hashing
**Purpose**: Verify MurmurHash3 implementation for attribute cache invalidation  
**DS Requirement**: `attributesHash = MurmurHash3(JSON.stringify(filteredAttributes))`  
**Why Separate**: Bug bash tests cache invalidation; DS needs specific hash algorithm

**Steps**:
1. Create user with attributes: {age: 50, country: "BD"}
2. Calculate expected MurmurHash3 value for JSON.stringify({age: 50, country: "BD"})
3. Make CMAB decision (populates cache with hash)
4. Enable detailed cache logging to see attribute hash
5. Verify calculated hash matches cached hash

**Expected**:
- Attributes JSON: `{"age":50,"country":"BD"}`
- MurmurHash3 calculation matches cached hash value
- Hash changes when attributes change

**DS Validation**: Consistent attribute hashing across SDK versions

---

## DS Test D: Standard vs Custom Attribute Classification
**Purpose**: Verify correct attribute type classification in API requests  
**DS Requirement**: Proper "custom_attribute" vs "standard_attribute" typing  
**Why Separate**: Bug bash tests attributes work; DS needs exact type classification

**Steps**:
1. Create user with mixed attribute types:
   - Custom: `age=35, country="CA", premium=true`
   - Standard: `$opt_bot_filtering=false, $opt_ip="192.168.1.1"`
2. Call decide() and capture API request
3. Verify attribute type classification in request payload

**Expected API Request Format**:
```json
{
  "instances": [{
    "attributes": [
      {"id": "age", "type": "custom_attribute", "value": 35},
      {"id": "country", "type": "custom_attribute", "value": "CA"},
      {"id": "premium", "type": "custom_attribute", "value": true},
      {"id": "$opt_bot_filtering", "type": "standard_attribute", "value": false},
      {"id": "$opt_ip", "type": "standard_attribute", "value": "192.168.1.1"}
    ]
  }]
}
```

**DS Validation**: Correct attribute classification for DS model processing

---

## DS Test E: API Response Processing Validation
**Purpose**: Verify SDK correctly processes all DS API response scenarios  
**DS Requirement**: "API serves personalized experience or random variations"  
**Why Separate**: Bug bash tests basic responses; DS needs comprehensive response handling

**Steps**:
1. **Valid Response**: Mock `{"predictions": [{"variation_id": "5002"}]}`
2. Call decide(), verify variation "5002" returned
3. **Empty Predictions**: Mock `{"predictions": []}`
4. Call decide(), verify null decision with appropriate error
5. **Invalid Format**: Mock `{"invalid": "response"}`
6. Call decide(), verify null decision with appropriate error
7. **Network Error**: Mock network timeout
8. Call decide(), verify retry logic and final error handling

**Expected**:
- Valid response: Return correct variation
- Empty predictions: Return null decision, log error
- Invalid format: Return null decision, log parsing error  
- Network error: 3 retries, then null decision with timeout error

**DS Validation**: Robust response processing for reliable DS API integration

---

# TEST SUMMARY
## Complete Test List with DS Coverage

### **Bug Bash Tests:** *(🔬 = Also validates DS requirements)*
1. **Basic CMAB Decision Flow** 🔬 *(DS: API request format, UUID generation)*
2. **Cache Hit Scenario** 🔬 *(DS: UUID persistence, caching algorithm)*
3. **Cache Miss on Attribute Change** 🔬 *(DS: session management, attribute filtering)*
4. **IGNORE_CMAB_CACHE Option** *(Bug bash only)*
5. **RESET_CMAB_CACHE Option** *(Bug bash only)*
6. **INVALIDATE_USER_CMAB_CACHE Option** *(Bug bash only)*
7. **Decision Event Tracking** 🔬 *(DS: UUID in event metadata)*
8. **Fallback When Not Qualified** *(DS: efficient API usage)*
9. **Traffic Allocation Check** *(DS: proper bucketing logic)*
10. **API Error Handling** *(DS: fallback behavior)*
11. **Concurrent Requests** *(DS: performance optimization)*
12. **Forced Variation Override** *(DS: standard SDK behavior)*
13. **Cache Expiry** *(DS: 30-minute session timeout)*
14. **Attribute Types** *(DS: attribute formatting)*
15. **Performance Benchmarks** *(DS: latency measurement)*

### **Additional DS Implementation Compliance Tests:** 🔬
- **DS Test A**: 3x Retry Mechanism Validation
- **DS Test B**: Cache Key Algorithm Validation  
- **DS Test C**: MurmurHash3 Attribute Hashing
- **DS Test D**: Standard vs Custom Attribute Classification
- **DS Test E**: API Response Processing Validation

---

## Final Coverage Summary

**Total Tests**: 15 Bug Bash + 5 DS Implementation = 20 Tests  
**DS Requirements Coverage**: ✅ Complete (10 bug bash tests + 5 DS tests)  
**Bug Bash Coverage**: ✅ Complete (15 comprehensive tests)  
**Ready For**: Production deployment and DS team integration

### **Test Execution Strategy:**
1. **Run Bug Bash Tests 1-15**: Validates all SDK functionality
2. **Run DS Tests A-E**: Validates technical implementation compliance  
3. **DS Team Sign-off**: Confirms SDK provides all required data correctly