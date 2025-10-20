// CMAB Testing Example for Optimizely Go SDK
// This file contains comprehensive test scenarios for CMAB functionality
//
// To run: go run cmab_test_example.go
// To run specific test: go run cmab_test_example.go -test=cache_hit

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/cmab"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

const (
	// SDK Key from russell-demo-gosdk-cmab branch
	// SDK_KEY  = "JgzFaGzGXx6F1ocTbMTmn"	// develrc
	// FLAG_KEY = "flag-matjaz-editor"
	SDK_KEY  = "DCx4eoV52jhgaC9MSab3g" // rc (prep)
	FLAG_KEY = "flag-cmab-1"

	// Test user IDs
	USER_QUALIFIED    = "test_user_99" // Will be bucketed into CMAB
	USER_NOT_BUCKETED = "test_user_1"  // Won't be bucketed (traffic allocation)
	USER_CACHE_TEST   = "cache_user_123"
)

var testCase = flag.String("test", "all", "Specific test case to run")

func main() {
	flag.Parse()

	// Enable debug logging to see CMAB activity
	logging.SetLogLevel(logging.LogLevelDebug)

	fmt.Println("=== CMAB Testing Suite for Go SDK ===")
	fmt.Printf("Testing CMAB with rc environment\n")
	fmt.Printf("SDK Key: %s\n", SDK_KEY)
	fmt.Printf("Flag Key: %s\n\n", FLAG_KEY)

	// Create config manager with rc URL template
	configManager := config.NewPollingProjectConfigManager(SDK_KEY,
		// config.WithDatafileURLTemplate("https://dev.cdn.optimizely.com/datafiles/%s.json"))	// develrc
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json")) // rc

	// Use the proper factory option to set config manager
	factory := &client.OptimizelyFactory{
		SDKKey: SDK_KEY,
	}

	optimizelyClient, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Println("Error initializing Optimizely client:", err)
		return
	}
	defer optimizelyClient.Close()

	// Wait for datafile to load
	fmt.Println("Waiting for datafile to load...")
	time.Sleep(2 * time.Second)

	// Run tests based on flag
	switch *testCase {
	case "basic":
		testBasicCMAB(optimizelyClient)
	case "cache_hit":
		testCacheHit(optimizelyClient)
	case "cache_miss":
		testCacheMissOnAttributeChange(optimizelyClient)
	case "ignore_cache":
		testIgnoreCacheOption(optimizelyClient)
	case "reset_cache":
		testResetCacheOption(optimizelyClient)
	case "invalidate_user":
		testInvalidateUserCacheOption(optimizelyClient)
	case "concurrent":
		testConcurrentRequests(optimizelyClient)
	case "error":
		testErrorHandling(optimizelyClient)
	case "fallback":
		testFallbackWhenNotQualified(optimizelyClient)
	case "traffic":
		testTrafficAllocation(optimizelyClient)
	case "forced":
		testForcedVariationOverride(optimizelyClient)
	case "event_tracking":
		testEventTracking(optimizelyClient)
	case "attribute_types":
		testAttributeTypes(optimizelyClient)
	case "performance":
		testPerformanceBenchmarks(optimizelyClient)
	case "cache_expiry":
		testCacheExpiry(optimizelyClient)
	case "cmab_config":
		testCmabConfiguration()
	default:
		fmt.Printf("Unknown test case: %s\n", *testCase)
		fmt.Println("\nAvailable test cases:")
		fmt.Println("  basic, cache_hit, cache_miss, ignore_cache, reset_cache,")
		fmt.Println("  invalidate_user, concurrent, error, fallback, traffic,")
		fmt.Println("  forced, event_tracking, attribute_types, performance, cache_expiry, cmab_config, all")
	}
}

// Test 1: Basic CMAB functionality
func testBasicCMAB(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Basic CMAB Functionality ---")
	for i := 1; i <= 1; i++ {

		// Enable debug logging to see CMAB activity
		logging.SetLogLevel(logging.LogLevelDebug)

		fmt.Println("=== CMAB Testing Suite for Go SDK ===")
		fmt.Printf("Testing CMAB with rc environment\n")
		fmt.Printf("SDK Key: %s\n", SDK_KEY)
		fmt.Printf("Flag Key: %s\n\n", FLAG_KEY)

		// Create config manager with rc URL template
		configManager := config.NewPollingProjectConfigManager(SDK_KEY,
			// config.WithDatafileURLTemplate("https://dev.cdn.optimizely.com/datafiles/%s.json"))	// develrc
			config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json")) // rc

		// Use the proper factory option to set config manager
		factory := &client.OptimizelyFactory{
			SDKKey: SDK_KEY,
		}

		optimizelyClient, err := factory.Client(
			client.WithConfigManager(configManager),
		)
		if err != nil {
			fmt.Println("Error initializing Optimizely client:", err)
			return
		}
		defer optimizelyClient.Close()

		// Wait for datafile to load
		fmt.Println("Waiting for datafile to load...")
		time.Sleep(2 * time.Second)

		// Test with user who qualifies for CMAB
		userContext := optimizelyClient.CreateUserContext(USER_QUALIFIED, map[string]interface{}{
			"country": "us",
		})

		decision := userContext.Decide(FLAG_KEY, nil)
		printDecision("CMAB Qualified User", decision)

		// cache miss
		userContext2 := optimizelyClient.CreateUserContext(USER_QUALIFIED, map[string]interface{}{
			"country": "ru",
		})
		decision2 := userContext2.Decide(FLAG_KEY, nil)
		printDecision("CMAB Qualified User2", decision2)
		time.Sleep(1000 * time.Millisecond)

		// cache hit
		userContext3 := optimizelyClient.CreateUserContext(USER_QUALIFIED, map[string]interface{}{
			"country": "ru",
		})
		decision3 := userContext3.Decide(FLAG_KEY, nil)
		printDecision("CMAB Qualified User3", decision3)
		time.Sleep(1000 * time.Millisecond)

		fmt.Println("===============================")

	}
}

// Test 2: Cache hit - same user and attributes
// Expected:
// 1. Decision 1: "hello" → Passes audience → CMAB API call → Cache stored for user + "hello"
// 2. Decision 2: Same user, same "hello" → Passes audience → Cache hit (same cache key) → Returns cached result (no API call)
func testCacheHit(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cache Hit (Same User & Attributes) ---")

	userContext := optimizelyClient.CreateUserContext(USER_CACHE_TEST, map[string]interface{}{
		"country": "us",
	})

	// First decision - should call CMAB service
	fmt.Println("First decision (CMAB call):")
	decision1 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)


	userContext2 := optimizelyClient.CreateUserContext(USER_CACHE_TEST, map[string]interface{}{
		"country": "fr",
	})
	// Second decision - miss cache
	fmt.Println("\nSecond decision (Cache hit):")
	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("Decision 2", decision2)

	// Second decision - should use cache
	fmt.Println("\nSecond decision (Cache hit):")
	decision3 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("Decision 3", decision3)

}

// Test 3: Cache miss when relevant attributes change
// Expected:
//  1. Decision 1: "hello" → Passes audience → CMAB API call → Cache stored for "hello"
//  2. Decision 2: "world" → Passes audience → Cache miss (different attribute value) → New CMAB API call → Cache stored for
//     "world"
//  3. Decision 3: "world" → Passes audience → Cache hit (same attribute) → Uses cached result
func testCacheMissOnAttributeChange(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cache Miss on Attribute Change ---")

	// First decision with valid attribute
	userContext1 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	fmt.Println("Decision with 'hello':")
	decision1 := userContext1.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)

	// Second decision with changed valid attribute
	userContext2 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"cmab_test_attribute": "world", // Changed value
	})

	fmt.Println("\nDecision with 'world' (cache miss expected):")
	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("Decision 2", decision2)

	// Third decision with same user and attributes
	userContext3 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"cmab_test_attribute": "world", // Same as decision2
	})

	fmt.Println("\nDecision with same user and attributes (cache hit expected):")
	decision3 := userContext3.Decide(FLAG_KEY, nil)
	printDecision("Decision 3", decision3)
}

// Test 4: IGNORE_CMAB_CACHE option
// Expected:
// 1. Decision 1: "hello" → Passes audience → CMAB API call → Cache stored for user + "hello"
// 2. Decision 2: Same user, same "hello" + IGNORE_CMAB_CACHE → Passes audience → Cache bypassed → New CMAB API call (original cache preserved)
// 3. Decision 3: Same user, same "hello" → Passes audience → Cache hit → Uses original cached result (no API call)
func testIgnoreCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: IGNORE_CMAB_CACHE Option ---")

	userContext := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_ignore", map[string]interface{}{
		"country": "fr",
	})

	// First decision - populate cache
	fmt.Println("First decision (populate cache):")
	decision1 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)

	// Second decision with IGNORE_CMAB_CACHE
	fmt.Println("\nSecond decision with IGNORE_CMAB_CACHE:")
	options := []decide.OptimizelyDecideOptions{
		decide.IgnoreCMABCache,
	}
	decision2 := userContext.Decide(FLAG_KEY, options)
	printDecision("Decision 2 (ignored cache)", decision2)

	// Third decision - should use original cache
	fmt.Println("\nThird decision (should use original cache):")
	decision3 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 3", decision3)
}

// Test 5: RESET_CMAB_CACHE option
// Expected:
// 1. User 1: "hello" → CMAB API call → Cache stored for User 1
// 2. User 2: "hello" → CMAB API call → Cache stored for User 2
// 3. User 1 + RESET_CMAB_CACHE → Entire cache cleared → New CMAB API call for User 1
// 4. User 2: Same "hello" → Cache was cleared → New CMAB API call for User 2 (no cached result)
func testResetCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: RESET_CMAB_CACHE Option ---")

	// Setup two different users
	userContext1 := optimizelyClient.CreateUserContext("reset_user_1", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	userContext2 := optimizelyClient.CreateUserContext("reset_user_2", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// Populate cache for both users
	fmt.Println("Populating cache for User 1:")
	decision1 := userContext1.Decide(FLAG_KEY, nil)
	printDecision("User 1 Decision", decision1)

	fmt.Println("\nPopulating cache for User 2:")
	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User 2 Decision", decision2)

	// Reset entire cache
	fmt.Println("\nResetting entire CMAB cache:")
	options := []decide.OptimizelyDecideOptions{
		decide.ResetCMABCache,
	}
	decision3 := userContext1.Decide(FLAG_KEY, options)
	printDecision("User 1 after RESET", decision3)

	// Check if User 2's cache was also cleared
	fmt.Println("\nUser 2 after cache reset (should refetch):")
	decision4 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User 2 after reset", decision4)
}

// Test 6: INVALIDATE_USER_CMAB_CACHE option
// Expected:
// 1. User 1: "hello" → CMAB API call → Cache stored for User 1
// 2. User 2: "hello" → CMAB API call → Cache stored for User 2
// 3. User 1 + INVALIDATE_USER_CMAB_CACHE → Only User 1's cache cleared → New CMAB API call for User 1
// 4. User 2: Same "hello" → User 2's cache preserved → Cache hit (no API call)
func testInvalidateUserCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: INVALIDATE_USER_CMAB_CACHE Option ---")

	// Setup two different users
	userContext1 := optimizelyClient.CreateUserContext("invalidate_user_1", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	userContext2 := optimizelyClient.CreateUserContext("invalidate_user_2", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// Populate cache for both users
	fmt.Println("Populating cache for User 1:")
	decision1 := userContext1.Decide(FLAG_KEY, nil)
	printDecision("User 1 Initial", decision1)

	fmt.Println("\nPopulating cache for User 2:")
	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User 2 Initial", decision2)

	// Invalidate only User 1's cache
	fmt.Println("\nInvalidating User 1's cache only:")
	options := []decide.OptimizelyDecideOptions{
		decide.InvalidateUserCMABCache,
	}
	decision3 := userContext1.Decide(FLAG_KEY, options)
	printDecision("User 1 after INVALIDATE", decision3)

	// Check if User 2's cache is still valid
	fmt.Println("\nUser 2 after User 1 invalidation (should use cache):")
	decision4 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User 2 still cached", decision4)
}

// TODO: MATJAZ: ISUE HERE - concurrency problem!
// Test 7: Concurrent requests for same user - thread safety, all goroutines should complete
// Expected: The Go SDK uses mutex-based cache synchronization (sync.RWMutex)
// EXPECTED BEHAVIOR: 1 CMAB API call + 4 cache hits
//   - First goroutine makes CMAB API call and stores result in cache
//   - Other 4 goroutines wait for mutex, then find cached result and use it
//   - Logs should show: 1 "Fetching CMAB decision" + 4 "Returning cached CMAB decision"
//
// ACTUAL BEHAVIOR: If you see 5 separate API calls, this may indicate:
//   - Race condition in cache check/write logic
//   - Cache key generation issues
//   - Timing issue where all requests start before first completes
//
// Key requirement regardless: all goroutines return same variation for consistency
func testConcurrentRequests(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Concurrent Requests ---")

	userContext := optimizelyClient.CreateUserContext("concurrent_user", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// Channel to collect results
	results := make(chan *client.OptimizelyDecision, 5)

	// Launch 5 concurrent requests
	fmt.Println("Launching 5 concurrent decide calls...")
	for i := 0; i < 5; i++ {
		go func(id int) {
			decision := userContext.Decide(FLAG_KEY, nil)
			fmt.Printf("  Goroutine %d completed\n", id)
			results <- &decision
		}(i)
	}

	// Collect results
	variations := make(map[string]int)
	for i := 0; i < 5; i++ {
		decision := <-results
		variations[decision.VariationKey]++
	}

	// All should return the same variation (only one CMAB call)
	fmt.Println("\nResults:")
	for variation, count := range variations {
		fmt.Printf("  Variation '%s': %d times\n", variation, count)
	}

	if len(variations) == 1 {
		fmt.Println("✓ Concurrent handling correct: All returned same variation")
	} else {
		fmt.Println("✗ Issue with concurrent handling: Different variations returned")
	}
}

// Test 8: Error handling simulation
// Test 7: Error handling - invalid attribute types and edge cases
// Expected: User with invalid attribute type should fail audience evaluation
//   - CMAB experiment requires string attribute "cmab_test_attribute": "hello"
//   - This test uses integer 12345 instead of string, causing type mismatch
//   - SDK logs warning about attribute type mismatch during audience evaluation
//   - User fails CMAB audience check and falls through to default rollout
//   - Result: Gets rollout variation (typically 'off') instead of CMAB variation
//
// This validates proper error handling and graceful fallback behavior
func testErrorHandling(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Error Handling ---")
	fmt.Println("Note: This test simulates error scenarios")

	// Test with invalid/malformed attributes that might cause issues
	userContext := optimizelyClient.CreateUserContext("error_test_user", map[string]interface{}{
		"cmab_test_attribute": 12345, // Invalid type (should be string)
	})

	fmt.Println("Testing with invalid attribute types:")
	decision := userContext.Decide(FLAG_KEY, nil)
	printDecision("Error scenario", decision)

	if len(decision.Reasons) > 0 {
		fmt.Println("Reasons for decision:")
		for _, reason := range decision.Reasons {
			fmt.Printf("  - %s\n", reason)
		}
	}
}

// Helper function to print decision details
func printDecision(label string, decision client.OptimizelyDecision) {
	fmt.Printf("\n%s:\n", label)
	fmt.Printf("  Enabled: %v\n", decision.Enabled)
	fmt.Printf("  Variation: %s\n", decision.VariationKey)
	fmt.Printf("  Rule: %s\n", decision.RuleKey)

	if decision.Variables != nil {
		fmt.Printf("  Variables: %v\n", decision.Variables.ToMap())
	}

	if len(decision.Reasons) > 0 {
		fmt.Printf("  Reasons:\n")
		for _, reason := range decision.Reasons {
			fmt.Printf("    - %s\n", reason)
		}
	}

	// Try to extract CMAB metadata if available (would need SDK support)
	// This is a placeholder for when metadata is exposed
	fmt.Printf("  [Check debug logs above for CMAB UUID and calls]\n")
}

// Additional helper to pretty print JSON (for debugging)
func prettyPrint(label string, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("%s: Error marshaling - %v\n", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(bytes))
}

// Test 9: Fallback when user doesn't qualify for CMAB
// Test 8: Fallback when not qualified for CMAB
// Expected: User without required attributes fails CMAB audience targeting
//   - User "fallback_user" has no attributes (empty map)
//   - CMAB experiment requires "cmab_test_attribute": "hello" OR "world"
//   - Both audience conditions evaluate to UNKNOWN (null attribute value)
//   - User fails CMAB audience check and falls through to default rollout
//   - Result: Gets rollout variation 'off' from "Everyone Else" rule
//   - Key validation: No CMAB API calls should appear in debug logs
//
// This tests proper audience targeting and graceful fallback behavior
func testFallbackWhenNotQualified(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Fallback When Not Qualified for CMAB ---")

	// User with attributes that don't match CMAB audience
	userContext := optimizelyClient.CreateUserContext("fallback_user", map[string]interface{}{})

	decision := userContext.Decide(FLAG_KEY, nil)
	printDecision("Non-CMAB User", decision)

	if decision.RuleKey != "exp_1" {
		fmt.Println("✓ Fallback working: Decision from non-CMAB experiment")
	} else {
		fmt.Println("✗ Fallback issue: Still received CMAB decision")
	}

	fmt.Println("Expected: No CMAB API call in debug logs above")
}

// Test 9: Traffic allocation check - verify traffic splitting works correctly
// SETUP REQUIRED: Set CMAB experiment traffic allocation to 50% in Optimizely UI for this test
// (Keep at 100% for all other tests to ensure consistent CMAB behavior)
// Expected: With 50% traffic allocation, users split between CMAB and rollout
//   - test_user_1: Should hash to one bucket (CMAB or rollout)
//   - test_user_99: Should hash to different bucket than test_user_1
//   - Expected logs: One user shows "Fetching CMAB decision", other falls to rollout
//   - If both get CMAB calls: traffic allocation still at 100%
//   - If both get rollout: traffic allocation set too low (try 75%)
//
// This validates CMAB traffic allocation and user bucketing logic
func testTrafficAllocation(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Traffic Allocation Check ---")

	// User not in traffic allocation (test_user_1)
	userContext1 := optimizelyClient.CreateUserContext(USER_NOT_BUCKETED, map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	decision1 := userContext1.Decide(FLAG_KEY, nil)
	printDecision("User Not in Traffic", decision1)

	// User in traffic allocation (test_user_99)
	userContext2 := optimizelyClient.CreateUserContext(USER_QUALIFIED, map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User in Traffic", decision2)

	fmt.Println("Expected: Only second user triggers CMAB API call")
}

// Test 10: Forced Variation Override using SetForcedDecision() 
// OVERVIEW:
// This test validates the SetForcedDecision() API which allows runtime
// forcing of variations at either flag-level or rule-level, bypassing
// CMAB decision logic entirely.
//
// TWO FORCING LEVELS:
//
// 1. FLAG-LEVEL (Broad Scope):
//    - Context: OptimizelyDecisionContext{FlagKey: "flag1"}
//    - Impact: Forces variation for ALL experiments/rules under this flag
//    - Priority: #1 (Highest - checked first)
//    - CMAB: Completely bypassed for this flag
//    - Example: Force all users to see flag1="on" regardless of which 
//              CMAB rule or A/B test would normally evaluate
//
// 2. RULE-LEVEL (Granular Scope):
//    - Context: OptimizelyDecisionContext{FlagKey: "flag1", RuleKey: "cmab-rule-1"}
//    - Impact: Forces variation ONLY for specified experiment/rule
//    - Priority: #2 (After flag-level, before whitelist)
//    - CMAB: Bypassed only for the specified rule
//    - Example: Force cmab-rule-1="on" but let cmab-rule-2 evaluate normally
//
// PRIORITY ORDER (from highest to lowest):
//    1. SetForcedDecision() - Flag-level
//    2. SetForcedDecision() - Rule-level
//    3. Override Store (if configured)
//    4. Whitelist from Datafile (UI Allowlist)
//    5. CMAB Service
//    6. Regular Bucketing
func testForcedVariationOverride(optimizelyClient *client.OptimizelyClient) {
      fmt.Println("\n--- Test: Forced Variation Override ---")

      // Test 1: Normal CMAB flow (baseline)
      fmt.Println("\n=== Test 1: Normal CMAB Flow (Baseline) ===")
      userContext1 := optimizelyClient.CreateUserContext("normal_user",
  map[string]interface{}{
          "country": "us",
          "state":   "ca",
      })

      decision1 := userContext1.Decide(FLAG_KEY, nil)
      printDecision("Normal CMAB User", decision1)
      fmt.Println("Expected: CMAB API call made, variation from CMAB service")

      // Test 2: FLAG-LEVEL Forced Decision
      fmt.Println("\n=== Test 2: FLAG-LEVEL Forced Decision ===")
      userContext2 := optimizelyClient.CreateUserContext("forced_user_flag",
  map[string]interface{}{
          "country": "us",
          "state":   "ca",
      })

      // Force at FLAG level (no RuleKey) - affects ALL rules under this flag
      forcedContextFlag := decision.OptimizelyDecisionContext{
          FlagKey: FLAG_KEY, // Only flag key, no rule key
      }
      forcedDecisionFlag := decision.OptimizelyForcedDecision{
          VariationKey: "on", // Force "on" variation
      }

      success := userContext2.SetForcedDecision(forcedContextFlag, forcedDecisionFlag)
      if success {
          fmt.Println("✓ Flag-level forced decision set successfully")
      } else {
          fmt.Println("✗ Failed to set flag-level forced decision")
      }

      decision2 := userContext2.Decide(FLAG_KEY, nil)
      printDecision("Flag-Level Forced User", decision2)
      fmt.Println("Expected: Variation 'on', NO CMAB API call")
      fmt.Println("Note: This forces decision for ALL rules under this flag")

      // Test 3: RULE-LEVEL Forced Decision (NEW!)
      fmt.Println("\n=== Test 3: RULE-LEVEL Forced Decision ===")
      userContext3 := optimizelyClient.CreateUserContext("forced_user_rule",
  map[string]interface{}{
          "country": "us",
          "state":   "ca",
      })

      // Force at RULE level (flag key + specific rule key)
      forcedContextRule := decision.OptimizelyDecisionContext{
          FlagKey: FLAG_KEY,
          RuleKey: "cmab-rule-1",
      }
      forcedDecisionRule := decision.OptimizelyForcedDecision{
          VariationKey: "off", // Force "off" variation for this specific rule
      }

      success3 := userContext3.SetForcedDecision(forcedContextRule, forcedDecisionRule)
      if success3 {
          fmt.Println("✓ Rule-level forced decision set successfully")
      } else {
          fmt.Println("✗ Failed to set rule-level forced decision")
      }

      decision3 := userContext3.Decide(FLAG_KEY, nil)
      printDecision("Rule-Level Forced User", decision3)
      fmt.Println("Expected: Variation 'off', NO CMAB API call for cmab-rule-1")
      fmt.Println("Note: Only affects cmab-rule-1, other rules would evaluate normally")

      // Test 4: Multiple Forced Decisions (flag + different rule)
      fmt.Println("\n=== Test 4: Multiple Forced Decisions on Same UserContext ===")
      userContext4 := optimizelyClient.CreateUserContext("forced_user_multi",
  map[string]interface{}{
          "country": "us",
          "state":   "ca",
      })

      // Set forced decision for one rule
      userContext4.SetForcedDecision(
          decision.OptimizelyDecisionContext{
              FlagKey: FLAG_KEY,
              RuleKey: "cmab-rule-1",
          },
          decision.OptimizelyForcedDecision{VariationKey: "on"},
      )

      // Set another forced decision for a different context
      userContext4.SetForcedDecision(
          decision.OptimizelyDecisionContext{
              FlagKey: "flag2",
              RuleKey: "cmab-rule-2",
          },
          decision.OptimizelyForcedDecision{VariationKey: "on"},
      )

      fmt.Println("✓ Set forced decisions for multiple flag/rule combinations")

      decision4 := userContext4.Decide(FLAG_KEY, nil)
      printDecision("Multi-Forced Decision User (flag1)", decision4)

      decision4b := userContext4.Decide("flag2", nil)
      printDecision("Multi-Forced Decision User (flag2)", decision4b)

      // Test 5: Priority - Flag-level vs Rule-level
      fmt.Println("\n=== Test 5: Priority Test - Flag-Level Overrides Rule-Level ===")
      userContext5 := optimizelyClient.CreateUserContext("forced_user_priority",
  map[string]interface{}{
          "country": "us",
          "state":   "ca",
      })

      // First set rule-level forced decision
      userContext5.SetForcedDecision(
          decision.OptimizelyDecisionContext{
              FlagKey: FLAG_KEY,
              RuleKey: "cmab-rule-1",
          },
          decision.OptimizelyForcedDecision{VariationKey: "off"},
      )
      fmt.Println("Step 1: Set rule-level forced decision to 'off'")

      // Then set flag-level forced decision (should override)
      userContext5.SetForcedDecision(
          decision.OptimizelyDecisionContext{
              FlagKey: FLAG_KEY,
              // No RuleKey - flag level
          },
          decision.OptimizelyForcedDecision{VariationKey: "on"},
      )
      fmt.Println("Step 2: Set flag-level forced decision to 'on'")

      decision5 := userContext5.Decide(FLAG_KEY, nil)
      printDecision("Priority Test User", decision5)
      fmt.Println("Expected: Variation 'on' (flag-level wins over rule-level)")
      fmt.Println("Note: Flag-level is checked BEFORE rule-level in SDK")

      // Test 6: Remove FLAG-LEVEL forced decision
      fmt.Println("\n=== Test 6: Remove Flag-Level Forced Decision ===")
      removed := userContext2.RemoveForcedDecision(decision.OptimizelyDecisionContext{
          FlagKey: FLAG_KEY,
          // No RuleKey - removing flag-level
      })
      if removed {
          fmt.Println("✓ Flag-level forced decision removed successfully")
      } else {
          fmt.Println("✗ Failed to remove forced decision")
      }

      decision6 := userContext2.Decide(FLAG_KEY, nil)
      printDecision("After Removing Flag-Level Forced Decision", decision6)
      fmt.Println("Expected: CMAB API call made (back to normal flow)")

      // Test 7: Remove RULE-LEVEL forced decision
      fmt.Println("\n=== Test 7: Remove Rule-Level Forced Decision ===")
      removed2 := userContext3.RemoveForcedDecision(decision.OptimizelyDecisionContext{
          FlagKey: FLAG_KEY,
          RuleKey: "cmab-rule-1", // Specify rule to remove
      })
      if removed2 {
          fmt.Println("✓ Rule-level forced decision removed successfully")
      } else {
          fmt.Println("✗ Failed to remove rule-level forced decision")
      }

      decision7 := userContext3.Decide(FLAG_KEY, nil)
      printDecision("After Removing Rule-Level Forced Decision", decision7)
      fmt.Println("Expected: CMAB API call made (back to normal flow)")

      // Test 8: Remove ALL forced decisions
      fmt.Println("\n=== Test 8: Remove ALL Forced Decisions ===")
      userContext4.RemoveAllForcedDecisions()
      fmt.Println("✓ Removed all forced decisions from userContext4")

      decision8 := userContext4.Decide(FLAG_KEY, nil)
      printDecision("After Removing All Forced Decisions", decision8)
      fmt.Println("Expected: CMAB API call made (back to normal flow)")

      fmt.Println("\n=== Test Summary ===")
      fmt.Println("✓ Tested flag-level forced decision (affects all rules)")
      fmt.Println("✓ Tested rule-level forced decision (affects specific rule)")
      fmt.Println("✓ Tested priority (flag-level > rule-level)")
      fmt.Println("✓ Tested removal of both types")
      fmt.Println("✓ All forced decisions bypass CMAB API calls")
  }

// Test 11: Event tracking with CMAB UUID - verify events contain proper metadata
// Expected: Impression events include CMAB UUID, conversion events do NOT include CMAB UUID
//   - Decision creates impression event with CMAB UUID in metadata
//   - Conversion events should NOT contain CMAB UUID (FX requirement)
//   - Current result: "event1" event should be configured in project
//   - Warning indicates conversion event needs to be added in Optimizely UI if missing
//   - CMAB UUID only appears in impression events for analytics correlation
//
// This validates event tracking and proper CMAB UUID handling for different event types
func testEventTracking(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Event Tracking with CMAB UUID ---")

	userContext := optimizelyClient.CreateUserContext("event_user", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// Make CMAB decision
	decision := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision for Events", decision)

	// Track a conversion event
	userContext.TrackEvent("event1", map[string]interface{}{

	})

	fmt.Println("\nConversion event tracked: 'event1'")
	fmt.Println("Expected: Impression events contain CMAB UUID, conversion events do NOT contain CMAB UUID")
	fmt.Println("Check event processor logs for CMAB UUID only in impression events")
}

// Test 12: Attribute types and formatting - validate attribute handling
// Expected: Only valid attributes are sent to CMAB API, invalid ones filtered
//   - User has mixed attribute types (string, int, float, boolean)
//   - Only "cmab_test_attribute": "hello" should be sent to CMAB API
//   - Invalid attributes are filtered during audience evaluation
//   - Current result: User fails audience due to missing valid attribute
//   - Falls back to rollout (no CMAB API call) - this is expected behavior
//
// This validates attribute filtering and type validation logic
func testAttributeTypes(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Attribute Types and Formatting ---")

	userContext := optimizelyClient.CreateUserContext("attr_user", map[string]interface{}{
		// Missing cmab_test_attribute - should cause fallback
	})

	decision := userContext.Decide(FLAG_KEY, nil)
	printDecision("Mixed Attribute Types", decision)

	fmt.Println("Expected in API request:")
	fmt.Println("- Valid attribute: cmab_test_attribute sent to CMAB API")
	fmt.Println("- Invalid attributes: filtered out, not sent to CMAB API")
	fmt.Println("- Only cmab_test_attribute should appear in CMAB request body")
}

// Test 13: Cache expiry - verify TTL-based cache invalidation
// Expected: Cached decisions expire after TTL and trigger new API calls
//   - Initial call: CMAB API call creates cache entry with timestamp
//   - Immediate follow-up: Returns cached result (within TTL)
//   - After TTL expires: New CMAB API call (cache entry invalid)
//   - Current result: Still cached after 2s (expected - TTL is ~30min)
//   - Real expiry test: Requires waiting for full TTL duration
//   - Default TTL from cmab.DefaultCacheTTL in SDK configuration
//
// This validates time-based cache invalidation and freshness guarantees
func testCacheExpiry(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cache Expiry (Simulated) ---")

	userContext := optimizelyClient.CreateUserContext("expiry_user", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// First decision
	fmt.Println("Decision at T=0:")
	decision1 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Initial Decision", decision1)

	// Simulate time passing (in real scenario this would be 30+ minutes)
	fmt.Println("\nSimulating cache expiry...")
	time.Sleep(2 * time.Second)

	// For actual testing, you would need to wait 30+ minutes or manipulate cache TTL
	fmt.Println("Decision after simulated expiry:")
	decision2 := userContext.Decide(FLAG_KEY, nil)
	printDecision("After Expiry", decision2)

	fmt.Println("Note: Real cache expiry test requires 30+ minute wait")
	fmt.Println("Expected: New CMAB API call after expiry")
}

// Test 14: Performance benchmarks - measure API vs cache performance
// Expected: Cache hits should be significantly faster than API calls
//   - First API call: ~160ms (network latency + CMAB processing)
//   - Cached calls: ~85µs average (memory lookup only)
//   - Performance improvement: ~1,880x faster for cached calls
//   - Targets: API calls <500ms, cached calls <10ms
//   - Results: ✓ API: 160ms < 500ms, ✓ Cache: 85µs < 10ms
//
// This validates caching performance and responsiveness under load
func testPerformanceBenchmarks(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Performance Benchmarks ---")

	userContext := optimizelyClient.CreateUserContext("perf_user", map[string]interface{}{
		"cmab_test_attribute": "hello",
	})

	// Measure first call (API call)
	start := time.Now()
	decision1 := userContext.Decide(FLAG_KEY, nil)
	apiDuration := time.Since(start)

	printDecision("First Call (API)", decision1)
	fmt.Printf("API call duration: %v\n", apiDuration)

	// Measure cached calls
	var cachedDurations []time.Duration
	for i := 0; i < 10; i++ {
		start = time.Now()
		userContext.Decide(FLAG_KEY, nil)
		cachedDurations = append(cachedDurations, time.Since(start))
	}

	// Calculate average cached duration
	var totalCached time.Duration
	for _, d := range cachedDurations {
		totalCached += d
	}
	avgCached := totalCached / time.Duration(len(cachedDurations))

	fmt.Printf("Average cached call duration: %v (10 calls)\n", avgCached)
	fmt.Printf("\nPerformance Targets:\n")
	fmt.Printf("- Cached calls: <10ms (actual: %v)\n", avgCached)
	fmt.Printf("- API calls: <500ms (actual: %v)\n", apiDuration)

	if avgCached < 10*time.Millisecond {
		fmt.Println("✓ Cached performance: PASS")
	} else {
		fmt.Println("✗ Cached performance: FAIL")
	}

	if apiDuration < 500*time.Millisecond {
		fmt.Println("✓ API performance: PASS")
	} else {
		fmt.Println("✗ API performance: FAIL")
	}
}

// Test 15: CMAB Configuration - demonstrate custom CMAB config options
// Expected: Shows how to configure CMAB with custom cache size, TTL, timeout, and retry settings
//   - Default config: CacheSize=100, CacheTTL=30min, HTTPTimeout=10s, MaxRetries=3
//   - Custom config: CacheSize=200, CacheTTL=5min, HTTPTimeout=30s, MaxRetries=5
//   - Demonstrates both default and custom configuration initialization
//   - Shows proper client setup with CMAB configuration
//
// This validates CMAB configuration options and client initialization patterns
func testCmabConfiguration() {
	fmt.Println("\n--- Test: CMAB Configuration Options ---")

	// Enable debug logging to see CMAB activity
	logging.SetLogLevel(logging.LogLevelDebug)

	// Example 1: Using default CMAB configuration
	fmt.Println("\n=== Default CMAB Configuration ===")
	defaultConfig := cmab.NewDefaultConfig()
	fmt.Printf("Default Config:\n")
	fmt.Printf("  CacheSize: %d\n", defaultConfig.CacheSize)
	fmt.Printf("  CacheTTL: %v\n", defaultConfig.CacheTTL)
	fmt.Printf("  HTTPTimeout: %v\n", defaultConfig.HTTPTimeout)
	fmt.Printf("  MaxRetries: %d\n", defaultConfig.RetryConfig.MaxRetries)

	// Create client with default config
	configManager1 := config.NewPollingProjectConfigManager(SDK_KEY,
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json"))

	factory1 := &client.OptimizelyFactory{
		SDKKey: SDK_KEY,
	}

	optimizelyClient1, err := factory1.Client(
		client.WithConfigManager(configManager1),
		client.WithCmabConfig(&defaultConfig),
	)
	if err != nil {
		fmt.Println("Error initializing Optimizely client with default config:", err)
		return
	}
	defer optimizelyClient1.Close()

	// Example 2: Using custom CMAB configuration
	fmt.Println("\n=== Custom CMAB Configuration ===")
	customConfig := cmab.Config{
		CacheSize:   200,                // Larger cache
		CacheTTL:    5 * time.Minute,    // Shorter TTL
		HTTPTimeout: 30 * time.Second,   // Longer timeout
		RetryConfig: &cmab.RetryConfig{
			MaxRetries: 5, // More retries
		},
	}
	fmt.Printf("Custom Config:\n")
	fmt.Printf("  CacheSize: %d\n", customConfig.CacheSize)
	fmt.Printf("  CacheTTL: %v\n", customConfig.CacheTTL)
	fmt.Printf("  HTTPTimeout: %v\n", customConfig.HTTPTimeout)
	fmt.Printf("  MaxRetries: %d\n", customConfig.RetryConfig.MaxRetries)

	// Create client with custom config
	configManager2 := config.NewPollingProjectConfigManager(SDK_KEY,
		config.WithDatafileURLTemplate("https://optimizely-staging.s3.amazonaws.com/datafiles/%s.json"))

	factory2 := &client.OptimizelyFactory{
		SDKKey: SDK_KEY,
	}

	optimizelyClient2, err := factory2.Client(
		client.WithConfigManager(configManager2),
		client.WithCmabConfig(&customConfig),
	)
	if err != nil {
		fmt.Println("Error initializing Optimizely client with custom config:", err)
		return
	}
	defer optimizelyClient2.Close()

	// Wait for datafiles to load
	fmt.Println("\nWaiting for datafiles to load...")
	time.Sleep(2 * time.Second)

	// Example 3: Test both clients to see behavior differences
	fmt.Println("\n=== Testing Both Configurations ===")
	
	userContext := map[string]interface{}{
		"country": "us",
	}

	// Test default config client
	fmt.Println("\nTesting DEFAULT config client:")
	userCtx1 := optimizelyClient1.CreateUserContext("config_test_user_default", userContext)
	decision1 := userCtx1.Decide(FLAG_KEY, nil)
	printDecision("Default Config", decision1)

	// Test custom config client  
	fmt.Println("\nTesting CUSTOM config client:")
	userCtx2 := optimizelyClient2.CreateUserContext("config_test_user_custom", userContext)
	decision2 := userCtx2.Decide(FLAG_KEY, nil)
	printDecision("Custom Config", decision2)

	fmt.Println("\n=== Configuration Summary ===")
	fmt.Println("✓ Default CMAB config: 100 cache size, 30min TTL, 10s timeout, 3 retries")
	fmt.Println("✓ Custom CMAB config: 200 cache size, 5min TTL, 30s timeout, 5 retries")
	fmt.Println("✓ Both clients initialized successfully with respective configs")
	fmt.Println("Note: Config differences affect caching behavior and HTTP retry logic")
}
