package main

import (
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
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
}
