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
	"log"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/decide"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

const (
	// SDK Key from russell-demo-gosdk-cmab branch
	SDK_KEY = "JgzFaGzGXx6F1ocTbMTmn"
	FLAG_KEY = "flag-matjaz-editor"
	
	// Test user IDs
	USER_QUALIFIED = "test_user_99"    // Will be bucketed into CMAB
	USER_NOT_BUCKETED = "test_user_1"  // Won't be bucketed (traffic allocation)
	USER_CACHE_TEST = "cache_user_123"
)

var testCase = flag.String("test", "all", "Specific test case to run")

func main() {
	flag.Parse()
	
	// Enable debug logging to see CMAB activity
	logging.SetLogLevel(logging.LogLevelDebug)
	
	fmt.Println("=== CMAB Testing Suite for Go SDK ===")
	fmt.Printf("SDK Key: %s\n", SDK_KEY)
	fmt.Printf("Testing with develrc environment\n\n")
	
	// Initialize client
	optimizelyClient := initializeClient()
	if optimizelyClient == nil {
		log.Fatal("Failed to initialize Optimizely client")
	}
	defer optimizelyClient.Close()
	
	// Wait for datafile
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
	case "all":
		runAllTests(optimizelyClient)
	default:
		fmt.Printf("Unknown test case: %s\n", *testCase)
	}
}

func initializeClient() *client.OptimizelyClient {
	// Create config manager with develrc URL template
	configManager := config.NewPollingProjectConfigManager(SDK_KEY,
		config.WithDatafileURLTemplate("https://dev.cdn.optimizely.com/datafiles/%s.json"))
	
	factory := &client.OptimizelyFactory{
		SDKKey: SDK_KEY,
	}
	
	optimizelyClient, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Printf("Error initializing client: %v\n", err)
		return nil
	}
	
	return optimizelyClient
}

func runAllTests(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n=== Running All Tests ===\n")
	
	testBasicCMAB(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testCacheHit(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testCacheMissOnAttributeChange(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testIgnoreCacheOption(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testResetCacheOption(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testInvalidateUserCacheOption(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	testConcurrentRequests(optimizelyClient)
	time.Sleep(1 * time.Second)
	
	fmt.Println("\n=== All Tests Completed ===")
}

// Test 1: Basic CMAB functionality
func testBasicCMAB(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Basic CMAB Functionality ---")
	
	// Test with user who qualifies for CMAB
	userContext := optimizelyClient.CreateUserContext(USER_QUALIFIED, map[string]interface{}{
		"category": "cmab",
		"age":      50,
		"country":  "BD",
	})
	
	decision := userContext.Decide(FLAG_KEY, nil)
	printDecision("CMAB Qualified User", decision)
	
	// Test with user who doesn't meet audience conditions
	userContextNotQualified := optimizelyClient.CreateUserContext(USER_NOT_BUCKETED, map[string]interface{}{
		"category": "not-cmab",
		"age":      50,
		"country":  "BD",
	})
	
	decisionNotQualified := userContextNotQualified.Decide(FLAG_KEY, nil)
	printDecision("Non-CMAB User (falls through)", decisionNotQualified)
}

// Test 2: Cache hit - same user and attributes
func testCacheHit(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cache Hit (Same User & Attributes) ---")
	
	userContext := optimizelyClient.CreateUserContext(USER_CACHE_TEST, map[string]interface{}{
		"category": "cmab",
		"age":      30,
		"country":  "US",
	})
	
	// First decision - should call CMAB service
	fmt.Println("First decision (CMAB call):")
	decision1 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)
	
	// Second decision - should use cache
	fmt.Println("\nSecond decision (Cache hit):")
	decision2 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 2", decision2)
	
	// Verify same variation returned
	if decision1.VariationKey == decision2.VariationKey {
		fmt.Println("✓ Cache working: Same variation returned")
	} else {
		fmt.Println("✗ Cache issue: Different variations")
	}
}

// Test 3: Cache miss when relevant attributes change
func testCacheMissOnAttributeChange(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Cache Miss on Attribute Change ---")
	
	// First decision with age=25
	userContext1 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"category": "cmab",
		"age":      25,
		"country":  "US",
	})
	
	fmt.Println("Decision with age=25:")
	decision1 := userContext1.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)
	
	// Second decision with age=26 (relevant attribute change)
	userContext2 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"category": "cmab",
		"age":      26,  // Changed
		"country":  "US",
	})
	
	fmt.Println("\nDecision with age=26 (cache miss expected):")
	decision2 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("Decision 2", decision2)
	
	// Third decision with non-relevant attribute change
	userContext3 := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_attr", map[string]interface{}{
		"category": "cmab",
		"age":      26,
		"country":  "US",
		"language": "EN",  // Non-relevant attribute
	})
	
	fmt.Println("\nDecision with added language attribute (cache hit expected):")
	decision3 := userContext3.Decide(FLAG_KEY, nil)
	printDecision("Decision 3", decision3)
}

// Test 4: IGNORE_CMAB_CACHE option
func testIgnoreCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: IGNORE_CMAB_CACHE Option ---")
	
	userContext := optimizelyClient.CreateUserContext(USER_CACHE_TEST+"_ignore", map[string]interface{}{
		"category": "cmab",
		"age":      35,
		"country":  "UK",
	})
	
	// First decision - populate cache
	fmt.Println("First decision (populate cache):")
	decision1 := userContext.Decide(FLAG_KEY, nil)
	printDecision("Decision 1", decision1)
	
	// Second decision with IGNORE_CMAB_CACHE
	fmt.Println("\nSecond decision with IGNORE_CMAB_CACHE:")
	options := []decide.OptimizelyDecideOption{
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
func testResetCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: RESET_CMAB_CACHE Option ---")
	
	// Setup two different users
	userContext1 := optimizelyClient.CreateUserContext("reset_user_1", map[string]interface{}{
		"category": "cmab",
		"age":      40,
		"country":  "CA",
	})
	
	userContext2 := optimizelyClient.CreateUserContext("reset_user_2", map[string]interface{}{
		"category": "cmab",
		"age":      45,
		"country":  "AU",
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
	options := []decide.OptimizelyDecideOption{
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
func testInvalidateUserCacheOption(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: INVALIDATE_USER_CMAB_CACHE Option ---")
	
	// Setup two different users
	userContext1 := optimizelyClient.CreateUserContext("invalidate_user_1", map[string]interface{}{
		"category": "cmab",
		"age":      50,
		"country":  "DE",
	})
	
	userContext2 := optimizelyClient.CreateUserContext("invalidate_user_2", map[string]interface{}{
		"category": "cmab",
		"age":      55,
		"country":  "FR",
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
	options := []decide.OptimizelyDecideOption{
		decide.InvalidateUserCMABCache,
	}
	decision3 := userContext1.Decide(FLAG_KEY, options)
	printDecision("User 1 after INVALIDATE", decision3)
	
	// Check if User 2's cache is still valid
	fmt.Println("\nUser 2 after User 1 invalidation (should use cache):")
	decision4 := userContext2.Decide(FLAG_KEY, nil)
	printDecision("User 2 still cached", decision4)
}

// Test 7: Concurrent requests for same user
func testConcurrentRequests(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Concurrent Requests ---")
	
	userContext := optimizelyClient.CreateUserContext("concurrent_user", map[string]interface{}{
		"category": "cmab",
		"age":      60,
		"country":  "JP",
	})
	
	// Channel to collect results
	results := make(chan *decide.OptimizelyDecision, 5)
	
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
func testErrorHandling(optimizelyClient *client.OptimizelyClient) {
	fmt.Println("\n--- Test: Error Handling ---")
	fmt.Println("Note: This test simulates error scenarios")
	
	// Test with invalid/malformed attributes that might cause issues
	userContext := optimizelyClient.CreateUserContext("error_test_user", map[string]interface{}{
		"category": "cmab",
		"age":      "not_a_number",  // Invalid type
		"country":  "",               // Empty value
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
func printDecision(label string, decision decide.OptimizelyDecision) {
	fmt.Printf("\n%s:\n", label)
	fmt.Printf("  Enabled: %v\n", decision.Enabled)
	fmt.Printf("  Variation: %s\n", decision.VariationKey)
	fmt.Printf("  Rule: %s\n", decision.RuleKey)
	
	if len(decision.Variables) > 0 {
		fmt.Printf("  Variables: %v\n", decision.Variables)
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