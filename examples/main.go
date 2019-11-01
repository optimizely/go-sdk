// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/pkg/logging"
)

func main() {
	sdkKey := "4SLpaJA1r1pgE6T2CoMs9q"
	logging.SetLogLevel(logging.LogLevelDebug)
	user := entities.UserContext{
		ID: "mike ng",
		Attributes: map[string]interface{}{
			"country":      "Unknown",
			"likes_donuts": true,
		},
	}
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: sdkKey,
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
		SDKKey: sdkKey,
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

	optimizelyClient, _ = optimizelyFactory.Client(
		client.WithPollingConfigManager(sdkKey, time.Second, nil),
		client.WithCompositeDecisionService(sdkKey),
		client.WithBatchEventProcessor(event.DefaultBatchSize, event.DefaultEventQueueSize, event.DefaultEventFlushInterval),
	)
	optimizelyClient.Close()

	/************* Setting experiment overrides (a.k.a. "forced variations") ********************/
	overrideKey := decision.ExperimentOverrideKey{
		Experiment: "aaaa",
		UserID:     "Matt",
	}
	overrides := map[decision.ExperimentOverrideKey]string{
		overrideKey: "variation_1",
	}
	compositeService := decision.NewCompositeService(
		sdkKey,
		decision.WithExperimentOverridesMap(overrides),
	)
	optimizelyClient, _ = optimizelyFactory.Client(
		client.WithDecisionService(compositeService),
	)
	// Optimizely client now has "variation_1" forced for user "Matt" in experiment "aaaa"
	// The forced variation will work regardless of whether "aaaa" is an A/B test or a Feature Test.
}
