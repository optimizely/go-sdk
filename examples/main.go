// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/v2/pkg/client"
	"github.com/optimizely/go-sdk/v2/pkg/config"
	"github.com/optimizely/go-sdk/v2/pkg/logging"
)

// SIMPLE EXAMPLE: Uncomment this to test simple production CDN approach
func main() {
	customTemplateExample()
}

func customTemplateExample() {
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

	fmt.Printf("\nNote: Check logs above for CMAB/prediction endpoint calls\n")
}

// import (
// 	"fmt"
// 	"time"

// 	optimizely "github.com/optimizely/go-sdk/v2"
// 	"github.com/optimizely/go-sdk/v2/pkg/client"
// 	"github.com/optimizely/go-sdk/v2/pkg/event"
// 	"github.com/optimizely/go-sdk/v2/pkg/logging"
// )

// func main() {
// 	sdkKey := "RZKHh5HhUExLvpeieGZnD"
// 	logging.SetLogLevel(logging.LogLevelDebug)

// 	/************* Bad SDK Key  ********************/

// 	if optimizelyClient, err := optimizely.Client("some_key"); err == nil {
// 		userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
// 			"country":      "Unknown",
// 			"likes_donuts": true,
// 		})
// 		decision := userContext.Decide("mutext_feat", nil)
// 		fmt.Printf("Is feature enabled? %v\n", decision.Enabled)
// 		if len(decision.Reasons[0]) > 0 {
// 			fmt.Println("A Valid 403 error received:", decision.Reasons[0])
// 		}
// 	}

// 	/************* Simple usage ********************/

// 	if optimizelyClient, err := optimizely.Client(sdkKey); err == nil {
// 		userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
// 			"country":      "US",
// 			"likes_donuts": false,
// 		})
// 		decision := userContext.Decide("mutext_feat", nil)
// 		fmt.Printf("Is feature enabled? %v\n", decision.Enabled)
// 	}

// 	// /************* StaticClient ********************/

// 	optimizelyFactory := &client.OptimizelyFactory{
// 		SDKKey: sdkKey,
// 	}

// 	optimizelyClient, err := optimizelyFactory.StaticClient()
// 	if err != nil {
// 		fmt.Printf("Error instantiating client: %s", err)
// 		return
// 	}

// 	userContext := optimizelyClient.CreateUserContext("user1", map[string]interface{}{
// 		"country":      "Unknown",
// 		"likes_donuts": true,
// 	})
// 	decision := userContext.Decide("mutext_feat", nil)
// 	fmt.Printf("Is feature enabled? %v\n", decision.Enabled)

// 	fmt.Println()
// 	optimizelyClient.Close() //  user can close dispatcher
// 	fmt.Println()

// 	/************* Client ********************/

// 	optimizelyFactory = &client.OptimizelyFactory{
// 		SDKKey: sdkKey,
// 	}

// 	optimizelyClient, err = optimizelyFactory.Client()

// 	if err != nil {
// 		fmt.Printf("Error instantiating client: %s", err)
// 		return
// 	}

// 	userContext = optimizelyClient.CreateUserContext("user1", map[string]interface{}{
// 		"country":      "Unknown",
// 		"likes_donuts": true,
// 	})
// 	decision = userContext.Decide("mutext_feat", nil)
// 	fmt.Printf("Is feature enabled? %v\n", decision.Enabled)

// 	optimizelyClient.Close() //  user can close dispatcher

// 	/************* Setting Polling Interval ********************/

// 	optimizelyClient, _ = optimizelyFactory.Client(
// 		client.WithPollingConfigManager(time.Second, nil),
// 		client.WithBatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
// 	)

// 	optimizelyClient.Close()
// }
