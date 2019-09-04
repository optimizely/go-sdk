// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/logging"
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

	app, err := optimizelyFactory.StaticClient()

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	enabled, _ := app.IsFeatureEnabled("mutext_feat", user)
	fmt.Printf("Is feature enabled? %v\n", enabled)

	fmt.Println()

	/************* ClientWithOptions - custom context  ********************/

	optimizelyFactory = &client.OptimizelyFactory{
		SDKKey: "4SLpaJA1r1pgE6T2CoMs9q",
	}
	ctx := context.Background()
	ctx, cancelManager := context.WithCancel(ctx) // user can set up any context
	clientOptions := client.Options{
		Context: ctx,
	}

	app, err = optimizelyFactory.ClientWithOptions(clientOptions)
	cancelManager() //  user can cancel anytime

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	enabled, _ = app.IsFeatureEnabled("mutext_feat", user)
	fmt.Printf("Is feature enabled? %v\n", enabled)

	time.Sleep(1000 * time.Millisecond)
	fmt.Println()

	/************* Client ********************/

	optimizelyFactory = &client.OptimizelyFactory{
		SDKKey: "4SLpaJA1r1pgE6T2CoMs9q",
	}

	app, err = optimizelyFactory.Client()
	app.Close() //  user can cancel anytime

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	enabled, _ = app.IsFeatureEnabled("mutext_feat", user)
	fmt.Printf("Is feature enabled? %v\n", enabled)

}
