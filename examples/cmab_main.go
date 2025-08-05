package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

func main() {
	testCMABWithMaster()
}

func testCMABWithMaster() {
	sdkKey := "JgzFaGzGXx6F1ocTbMTmn" // matjaz editor develrc flag!! Able to run the experiment for cmab field to be included in datafile

	// Enable debug logging to see CMAB activity
	logging.SetLogLevel(logging.LogLevelDebug)

	fmt.Printf("Testing CMAB with master branch (camelCase JSON tags)\n")
	fmt.Printf("Attempting to fetch datafile from develrc environment\n")

	// Create config manager with develrc URL template - match Python approach
	configManager := config.NewPollingProjectConfigManager(sdkKey,
		config.WithDatafileURLTemplate("https://dev.cdn.optimizely.com/datafiles/%s.json"))

	// Use the proper factory option to set config manager
	factory := &client.OptimizelyFactory{
		SDKKey: sdkKey,
	}

	client, err := factory.Client(
		client.WithConfigManager(configManager),
	)
	if err != nil {
		fmt.Println("Error initializing Optimizely client:", err)
		return
	}
	defer client.Close()

	// Wait for datafile to load
	fmt.Println("Waiting for datafile to load...")
	time.Sleep(2 * time.Second)

	// Match Python example: user123 with empty attributes initially
	userContext := client.CreateUserContext("user123", map[string]interface{}{})

	fmt.Println("Making decision for flag 'flag-matjaz-editor'...")
	decision := userContext.Decide("flag-matjaz-editor", nil)

	fmt.Printf("=== DECISION RESULT ===\n")
	fmt.Printf("Enabled: %v\n", decision.Enabled)
	fmt.Printf("Variation: %s\n", decision.VariationKey)
	fmt.Printf("Rule: %s\n", decision.RuleKey)
	fmt.Printf("Reasons: %v\n", decision.Reasons)

	// Check if this looks like a CMAB decision
	if len(decision.Reasons) > 0 {
		fmt.Printf("\n=== CMAB ANALYSIS ===\n")
		for i, reason := range decision.Reasons {
			fmt.Printf("Reason %d: %s\n", i+1, reason)
			// Look for CMAB-specific reasons
			if contains(reason, "cmab") || contains(reason, "prediction") || contains(reason, "bandit") {
				fmt.Printf("  ^^^ This indicates CMAB activity!\n")
			}
		}
	} else {
		fmt.Printf("\n=== CMAB ANALYSIS ===\n")
		fmt.Printf("No reasons provided - this suggests a simple rollout (no CMAB)\n")
	}

	fmt.Printf("\nNote: Check logs above for any CMAB/prediction endpoint calls\n")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}