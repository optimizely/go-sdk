// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"fmt"
	"time"

	optimizely "github.com/optimizely/go-sdk"
	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
)

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
}
