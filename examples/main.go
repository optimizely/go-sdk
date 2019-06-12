package main

import (
	"fmt"

	"github.com/optimizely/go-sdk/optimizely/client"
)

func main() {
	optimizelyFactory := &client.OptimizelyFactory{
		SDKKey: "ABC",
	}
	client := optimizelyFactory.Client()
	fmt.Printf("Is feature enabled? %v", client.IsFeatureEnabled("go_sdk", "mike", nil))
}
