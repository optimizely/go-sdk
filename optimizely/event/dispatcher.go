package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/optimizely/go-sdk/optimizely/logging"
	"net/http"
)

// Dispatcher dispatches events
type Dispatcher interface {
	DispatchEvent(event LogEvent, callback func(success bool))
}

// HTTPEventDispatcher is the HTTP implementation of the Dispatcher interface
type HTTPEventDispatcher struct {
}

var dispatcherLogger = logging.GetLogger("EventDispatcher")

// DispatchEvent dispatches event with callback
func (*HTTPEventDispatcher) DispatchEvent(event LogEvent, callback func(success bool)) {

	jsonValue, _ := json.Marshal(event.event)
	resp, err := http.Post( event.endPoint, "application/json", bytes.NewBuffer(jsonValue))
	// also check response codes
	// resp.StatusCode == 400 is an error
	success := true

	if err != nil {
		dispatcherLogger.Error("http.Post failed:", err)
		success = false
	} else {
		if resp.StatusCode == 204 {
			success = true
		} else {
			fmt.Printf("http.Post invalid response %d", resp.StatusCode)
			success = false
		}
	}
	callback(success)

}
