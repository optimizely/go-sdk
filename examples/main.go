package main

import (
	"fmt"

	"time"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
)

func main() {
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "ABC",
	}
	client := optimizelyFactory.Client()

	user := entities.UserContext{
		ID:         "mike ng",
		Attributes: entities.UserAttributes{},
	}

	enabled, _ := client.IsFeatureEnabled("go_sdk", user)
	fmt.Printf("Is feature enabled? %v", enabled)

	processor := event.NewEventProcessor(100, 100)

	impression := event.UserEvent{}

	processor.ProcessImpression(impression)

	_, ok := processor.(*event.QueueingEventProcessor)

	if ok {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("\nending")
	}
}
