package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/logging"
)

func main() {
	logging.SetLogLevel(logging.LogLevelDebug)
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey:   "4SLpaJA1r1pgE6T2CoMs9q",
		Datafile: []byte("datafile_string"),
	}
	client, err := optimizelyFactory.Client()

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	user := entities.UserContext{
		ID: "mike ng",
		Attributes: map[string]interface{}{
			"country":      "Unknown",
			"likes_donuts": true,
		},
	}

	enabled, _ := client.IsFeatureEnabled("binary_feature", user)
	fmt.Printf("Is feature enabled? %v", enabled)

	processor := event.NewEventProcessor(100, 100)

	impression := event.UserEvent{}

	processor.ProcessEvent(impression)

	_, ok := processor.(*event.QueueingEventProcessor)

	if ok {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("\nending")
	}
}
