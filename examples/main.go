// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/notification"
)

func main() {

	logging.SetLogLevel(logging.LogLevelDebug)
	user := entities.UserContext{
		ID: "mike ng",
		Attributes: map[string]interface{}{
			"country":      "Unknown",
			"likes_donuts": true,
		},
	}
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "4SLpaJA1r1pgE6T2CoMs9q",
	}

	/************* StaticClient ********************/

	optimizelyClient, err := optimizelyFactory.StaticClient()

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	enabled, _ := optimizelyClient.IsFeatureEnabled("mutext_feat", user)
	fmt.Printf("Is feature enabled? %v\n", enabled)

	fmt.Println()
	optimizelyClient.Close() //  user can close dispatcher
	fmt.Println()
	/************* Client ********************/

	optimizelyFactory = &client.OptimizelyFactory{
		SDKKey: "4SLpaJA1r1pgE6T2CoMs9q",
	}

	optimizelyClient, err = optimizelyFactory.Client()

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	enabled, _ = optimizelyClient.IsFeatureEnabled("mutext_feat", user)
	fmt.Printf("Is feature enabled? %v\n", enabled)
	optimizelyClient.Close() //  user can close dispatcher

	/************* Setting Polling Interval ********************/

	notificationCenter := notification.NewNotificationCenter()

	optimizelyClient, _ = optimizelyFactory.Client(
		client.PollingConfigManager("4SLpaJA1r1pgE6T2CoMs9q", time.Second, nil, notificationCenter),
		client.CompositeDecisionService(notificationCenter),
		client.BatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)
	optimizelyClient.Close()
}
