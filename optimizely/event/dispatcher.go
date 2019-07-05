package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Dispatcher interface {
	DispatchEvent(event interface{}, callback func(success bool))
}

type HttpEventDispatcher struct {
}

func (*HttpEventDispatcher) DispatchEvent(event interface{}, callback func(success bool)) {
	impression, ok := event.(EventBatch)
	// add to current batch or create new batch
	// does a batch have to contain a decision or can it just be impressions?
	if ok {
		jsonValue, _ := json.Marshal(impression)
		resp, err := http.Post("https://logx.optimizely.com/v1/events", "application/json", bytes.NewBuffer(jsonValue))
		fmt.Println(resp)
		fmt.Println(string(jsonValue))
		// also check response codes
		// resp.StatusCode == 400 is an error
		success := true

		if err != nil {
			fmt.Println(err)
			success = false
		} else {
			if resp.StatusCode == 204 {
				success = true
			} else {
				fmt.Printf("invalid response %d", resp.StatusCode)
				success = false
			}
		}
		callback(success)
	}
}

