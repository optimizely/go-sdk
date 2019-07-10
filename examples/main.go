package main

import (
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/event"
	"time"
)

func main() {
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "ABC",
	}
	client := optimizelyFactory.Client()
	fmt.Printf("Is feature enabled? %v", client.IsFeatureEnabled("go_sdk", "mike", nil))

	processor := event.NewEventProcessor(100, 100)

	impression := event.UserEvent{}

	processor.ProcessEvent(impression)

	_, ok := processor.(*event.QueueingEventProcessor)

	if ok {
		time.Sleep(1000 * time.Millisecond)
		fmt.Println("\nending")
	}
}
