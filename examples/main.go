package main

import (
	"github.com/optimizely/go-sdk/optimizely"
)

func main() {
	OPTIMIZELY_ACCOUNT_ID := 12345
	buffer := optimizely.FetchProjectConfig(OPTIMIZELY_ACCOUNT_ID)
	optimizely.DeserializeConfigBuffer(buffer)
}
