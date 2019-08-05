package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent, callback func(success bool))
}

// HTTPEventDispatcher represents HTTPEventDispatcher object
type HTTPEventDispatcher struct {
}

// DispatchEvent dispatches event with callback
func (*HTTPEventDispatcher) DispatchEvent(event LogEvent, callback func(success bool)) {
	// add to current batch or create new batch
	// does a batch have to contain a decision or can it just be impressions?

	jsonValue, _ := json.Marshal(event.event)
	resp, err := http.Post(event.endPoint, "application/json", bytes.NewBuffer(jsonValue))
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
