# SDK Tests for Data Science Team Requirements Validation
## Tests to ensure SDK provides everything DS team needs

---

## DS Requirement: SDK Sends Variation Requests with Required Fields

### Test DS-1: API Request Contains All Required Fields
**DS Requirement**: "SDKs sends variation requests with user id, attributes, cmab uuid, variations"

**Test Steps**:
1. Create user "ds_test_user" with attributes {age: 30, country: "US"}
2. Call decide() and intercept API request
3. Validate request structure

**Must Validate**:
```json
{
  "instances": [{
    "visitorId": "ds_test_user",           ✓ User ID present
    "experimentId": "2001",                ✓ Experiment ID present
    "cmabUUID": "<valid-uuid-v4>",        ✓ CMAB UUID present
    "attributes": [                        ✓ Attributes array present
      {
        "id": "age",                       ✓ Attribute ID
        "type": "custom_attribute",        ✓ Attribute type
        "value": 30                        ✓ Attribute value
      },
      {
        "id": "country",
        "type": "custom_attribute", 
        "value": "US"
      }
    ]
  }]
}
```

**Pass Criteria**: All fields present and properly formatted

---

## DS Requirement: CMAB UUID Creation and Management

### Test DS-2: UUID Generation
**DS Requirement**: "CMAB UUID creation in SDKs"

**Test Steps**:
1. Make CMAB decision
2. Extract UUID from decision
3. Validate UUID format (regex: ^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$)

**Pass Criteria**: Valid UUID v4 format

### Test DS-3: UUID Persistence in Cache
**DS Requirement**: "SDKs manages session timeout for 4 cases for CMAB"

**Test Steps**:
1. Make decision, note UUID
2. Make same decision again (cache hit)
3. Verify same UUID returned
4. Change attribute, make decision
5. Verify new UUID generated

**Pass Criteria**: 
- Same UUID for cached decisions
- New UUID when cache invalidated

---

## DS Requirement: Event Pipeline Integration

### Test DS-4: Decision Events Include CMAB UUID
**DS Requirement**: "SDKs adds CMAB UUID for events payload for logx endpoint"

**Test Steps**:
1. Configure event processor (batch_size: 1)
2. Make CMAB decision
3. Capture impression event
4. Verify event structure

**Must Contain**:
```json
{
  "visitors": [{
    "snapshots": [{
      "decisions": [{
        "metadata": {
          "cmab_uuid": "<uuid-from-decision>"  ✓ UUID in metadata
        }
      }]
    }]
  }]
}
```

**Pass Criteria**: cmab_uuid present in decision event metadata

### Test DS-5: Conversion Events Include CMAB UUID
**DS Requirement**: "Logx endpoint data moved to decision and conversion tables with UUID"

**Test Steps**:
1. Make CMAB decision, note UUID
2. Track conversion: userContext.TrackEvent("purchase")
3. Capture conversion event
4. Verify UUID present

**Pass Criteria**: Same CMAB UUID in conversion event

---

## DS Requirement: Caching Algorithm

### Test DS-6: Hashing Algorithm Implementation
**DS Requirement**: "SDKs implement a hashing algorithm to reduce number of requests"

**Test Steps**:
1. User with {age: 25, country: "CA", language: "EN"}
2. CMAB uses only age and country
3. Make decision (cache populated)
4. Change language to "FR" 
5. Make decision again

**Pass Criteria**:
- Cache hit when non-relevant attribute changes
- Cache based on relevant attributes only

### Test DS-7: Session Timeout - 30 Minutes
**DS Requirement**: "SDKs manages session timeout"

**Test Steps**:
1. Decision at T=0
2. Decision at T=29 minutes (should cache hit)
3. Decision at T=31 minutes (should cache miss)

**Pass Criteria**: 
- Cache valid for 30 minutes
- New API call after timeout

### Test DS-8: Session Timeout - 24 Hour Expiry
**DS Requirement**: "SDKs manages session timeout for 4 cases"

**Test Steps**:
1. Decision at T=0
2. Simulate 24+ hours passing
3. Make decision again

**Pass Criteria**: Cache expired, new API call

### Test DS-9: Attribute Change Re-bucketing
**DS Requirement**: "Verify CMAB UUID changes according to pre-defined conditions"

**Test Steps**:
1. Decision with {age: 30}
2. Decision with {age: 31}
3. Verify different UUIDs

**Pass Criteria**: New UUID when relevant attributes change

---

## DS Requirement: User Identification

### Test DS-10: User ID Provided
**DS Requirement**: "SDKs provide a user identification variable"

**Test Steps**:
1. Create users with different IDs
2. Make decisions
3. Verify each API request has correct userId

**Pass Criteria**: userId correctly passed in all requests

---

## DS Requirement: Experiment Configuration

### Test DS-11: SDK Gets Configuration from Datafile
**DS Requirement**: "SDKs get experiment configuration from Flags DBs"

**Test Steps**:
1. Load datafile with CMAB experiment
2. Verify SDK recognizes:
   - attributeIds for filtering
   - trafficAllocation for bucketing
   - variations available

**Pass Criteria**: SDK correctly parses CMAB configuration

---

## DS Requirement: Error Handling

### Test DS-12: Fallback on API Failure
**DS Requirement**: "SDKs serves random variations as default mechanism"

**Test Steps**:
1. Force API failure (network error)
2. Call decide()
3. Check response

**Pass Criteria**: 
- Returns null/false decision
- No variation assigned
- Error in decision reasons

### Test DS-13: Retry Mechanism
**DS Requirement**: "In case of error with CMAB endpoint, retry 3x"

**Test Steps**:
1. Mock API to fail twice, succeed third time
2. Call decide()
3. Count API calls

**Pass Criteria**: Exactly 3 API attempts before success/failure

---

## DS Requirement: Attribute Formatting

### Test DS-14: Attribute Types Classification
**DS Requirement**: Attributes sent with proper type classification

**Test Steps**:
1. User with custom attributes (age, country)
2. User with standard attributes ($opt_bot_filtering)
3. Verify API request

**Expected Format**:
- Custom: `{"id": "age", "type": "custom_attribute", "value": 30}`
- Standard: `{"id": "$opt_bot_filtering", "type": "standard_attribute", "value": false}`

**Pass Criteria**: Correct type classification

### Test DS-15: Only Relevant Attributes Sent
**DS Requirement**: Only attributes used in CMAB sent to API

**Test Steps**:
1. CMAB configured with attributeIds: ["age", "country"]
2. User has: {age: 30, country: "US", hobby: "music", job: "engineer"}
3. Check API request

**Pass Criteria**: Only age and country in request

---

## DS Requirement: Performance

### Test DS-16: Latency Measurement
**DS Requirement**: "Measure latency in SDKs"

**Test Steps**:
1. Measure 100 cached decisions
2. Measure 10 API decisions
3. Calculate averages

**Pass Criteria**:
- Report latency metrics
- Cached: <10ms
- API: <500ms

---

## Summary Checklist for DS Team

### ✅ API Request Validation
- [ ] Test DS-1: All required fields present
- [ ] Test DS-10: User ID provided
- [ ] Test DS-14: Attribute types correct
- [ ] Test DS-15: Attribute filtering works

### ✅ UUID Management
- [ ] Test DS-2: UUID generation
- [ ] Test DS-3: UUID persistence
- [ ] Test DS-9: UUID changes on re-bucketing

### ✅ Event Pipeline
- [ ] Test DS-4: Decision events have UUID
- [ ] Test DS-5: Conversion events have UUID

### ✅ Cache & Performance
- [ ] Test DS-6: Hashing algorithm
- [ ] Test DS-7: 30-minute timeout
- [ ] Test DS-8: 24-hour expiry
- [ ] Test DS-16: Latency metrics

### ✅ Error Handling
- [ ] Test DS-12: Fallback behavior
- [ ] Test DS-13: 3x retry logic

### ✅ Configuration
- [ ] Test DS-11: Datafile parsing

---

## Test Execution for DS Validation

Run these tests in order to validate each DS requirement:

```go
// Example test runner
func ValidateForDS() {
    results := map[string]bool{
        "API_Fields": TestDS1_APIRequest(),
        "UUID_Generation": TestDS2_UUIDFormat(),
        "Event_Pipeline": TestDS4_DecisionEvents() && TestDS5_ConversionEvents(),
        "Cache_Algorithm": TestDS6_Hashing(),
        "Session_Management": TestDS7_Timeout() && TestDS8_Expiry(),
        "Error_Handling": TestDS12_Fallback() && TestDS13_Retry(),
    }
    
    // Report to DS team
    fmt.Println("DS Integration Validation:")
    for requirement, passed := range results {
        status := "❌"
        if passed {
            status = "✅"
        }
        fmt.Printf("%s %s\n", status, requirement)
    }
}
```

This ensures SDK provides everything the DS team needs for their model training and prediction service.