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
	impression, ok := event.(Impression)
	if ok {
		jsonValue, _ := json.Marshal(impression)
		resp, err := http.Post("https://logx.optimizely.com/v1/events", "application/json", bytes.NewBuffer(jsonValue))
		fmt.Println(resp)
		// also check response codes
		// resp.StatusCode == 400 is an error
		if err != nil {
			fmt.Println(err)
			callback(false)
		} else {
			callback(true)
		}
	}
}

