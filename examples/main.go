package main

import (
	"fmt"
	"github.com/optimizely/go-sdk/optimizely"
)

func main() {
	OPTIMIZELY_ACCOUNT_ID := "12345"
	_, err := optimizely.GetOptimizelyClient(OPTIMIZELY_ACCOUNT_ID)
	if err != nil {
		fmt.Print(err)
	}
}
