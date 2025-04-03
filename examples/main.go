// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"fmt"
	"time"

	optimizely "github.com/optimizely/go-sdk/v2"
	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/event"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// /************* CMAB Example ********************/

// func cmabExample() {
//     sdkKey := "RZKHh5HhUExLvpeieGZnD"
//     logging.SetLogLevel(logging.LogLevelDebug)

//     fmt.Println("\n/************* CMAB Example ********************/")

//     // Initialize client with CMAB support
//     optimizelyFactory := &client.OptimizelyFactory{
//         SDKKey: sdkKey,
//     }

//     optimizelyClient, err := optimizelyFactory.Client()
//     if err != nil {
//         fmt.Printf("Error instantiating client: %s\n", err)
//         return
//     }
//     defer optimizelyClient.Close()

//     // Create user context with attributes that might influence CMAB decisions
//     userContext := optimizelyClient.CreateUserContext("user123", map[string]interface{}{
//         "age":      28,
//         "location": "San Francisco",
//         "device":   "mobile",
//     })

//     // Get CMAB decision with reasons included
// 	cmabDecision, err := optimizelyClient.GetCMABDecision("cmab-rule-123", userContext, client.IncludeReasons)
//     if err != nil {
//         fmt.Printf("Error getting CMAB decision: %s\n", err)
//         return
//     }

//     // Display decision details
//     fmt.Printf("CMAB Decision for rule %s and user %s:\n", cmabDecision.RuleID, cmabDecision.UserID)
//     fmt.Printf("  Variant ID: %s\n", cmabDecision.VariantID)
//     fmt.Printf("  Attributes: %v\n", cmabDecision.Attributes)

//     if len(cmabDecision.Reasons) > 0 {
//         fmt.Println("  Reasons:")
//         for _, reason := range cmabDecision.Reasons {
//             fmt.Printf("    - %s\n", reason)
//         }
//     }

//     // Demonstrate cache usage
//     fmt.Println("\nGetting second decision (should use cache):")
// 	secondDecision, _ := optimizelyClient.GetCMABDecision("cmab-rule-123", userContext, client.IncludeReasons)
//     if len(secondDecision.Reasons) > 0 {
//         for _, reason := range secondDecision.Reasons {
//             fmt.Printf("    - %s\n", reason)
//         }
//     }

//     // Demonstrate cache invalidation
//     fmt.Println("\nInvalidating user cache and getting new decision:")
//     _ = optimizelyClient.InvalidateUserCMABCache(userContext.GetUserID())
// 	thirdDecision, _ := optimizelyClient.GetCMABDecision("cmab-rule-123", userContext, client.IncludeReasons)
//     if len(thirdDecision.Reasons) > 0 {
//         for _, reason := range thirdDecision.Reasons {
//             fmt.Printf("    - %s\n", reason)
//         }
//     }

//     // Demonstrate full cache reset
//     fmt.Println("\nResetting entire CMAB cache:")
//     _ = optimizelyClient.ResetCMABCache()
//     fmt.Println("Cache reset complete")
// }

func main() {
	sdkKey := "RZKHh5HhUExLvpeieGZnD"
	logging.SetLogLevel(logging.LogLevelDebug)

	/************* Bad SDK Key  ********************/

	if optimizelyClient, err := optimizely.Client("some_key"); err == nil {
		userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
			"country":      "Unknown",
			"likes_donuts": true,
		})
		decision := userContext.Decide("mutext_feat", nil)
		fmt.Printf("Is feature enabled? %v\n", decision.Enabled)
		if len(decision.Reasons[0]) > 0 {
			fmt.Println("A Valid 403 error received:", decision.Reasons[0])
		}
	}

	/************* Simple usage ********************/

	if optimizelyClient, err := optimizely.Client(sdkKey); err == nil {
		userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
			"country":      "US",
			"likes_donuts": false,
		})
		decision := userContext.Decide("mutext_feat", nil)
		fmt.Printf("Is feature enabled? %v\n", decision.Enabled)
	}

	// /************* StaticClient ********************/

	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: sdkKey,
	}

	optimizelyClient, err := optimizelyFactory.StaticClient()
	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
		"country":      "Unknown",
		"likes_donuts": true,
	})
	decision := userContext.Decide("mutext_feat", nil)
	fmt.Printf("Is feature enabled? %v\n", decision.Enabled)

	fmt.Println()
	optimizelyClient.Close() //  user can close dispatcher
	fmt.Println()

	/************* Client ********************/

	optimizelyFactory = &client.OptimizelyFactory{
		SDKKey: sdkKey,
	}

	optimizelyClient, err = optimizelyFactory.Client()

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	userContext = optimizelyClient.CreateUserContext("user1", map[string]interface{}{
		"country":      "Unknown",
		"likes_donuts": true,
	})
	decision = userContext.Decide("mutext_feat", nil)
	fmt.Printf("Is feature enabled? %v\n", decision.Enabled)

	optimizelyClient.Close() //  user can close dispatcher

	/************* Setting Polling Interval ********************/

	optimizelyClient, _ = optimizelyFactory.Client(
		client.WithPollingConfigManager(time.Second, nil),
		client.WithBatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)

	optimizelyClient.Close()

	// /************* Contextual Multi-Armed Bandit Example (CMAB) ********************/
    // cmabExample()
}
