// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"encoding/json"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/notification"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/pkg/profile"
)

// stressTest has everything that test app has. it is used to run profile
func stressTest() {
	/*
		For the test app, the biggest json file is used with 100 entities.
		DATAFILES_DIR has to be set to point to the path where 100_entities.json is located.
	*/
	type isFeatureEnabledRequestParams struct {
		FeatureKey string                 `json:"feature_flag_key"`
		UserID     string                 `json:"user_id"`
		Attributes map[string]interface{} `json:"attributes"`
	}

	type Context struct {
		CustomEventDispatcher string                   `json:"custom_event_dispatcher"`
		RequestID             string                   `json:"request_id"`
		UserProfileService    string                   `json:"user_profile_service"`
		Datafile              string                   `json:"datafile"`
		DispatchedEvents      []map[string]interface{} `json:"dispatched_events"`
	}

	strBytes := []byte(` { "context": {
			"datafile": "100_entities.json",
				"custom_event_dispatcher": "ProxyEventDispatcher",
				"request_id": "4e3e37e3-c7ae-4cb6-bbb5-f6ef93c84d43",
				"user_profile_service": "NoOpService",
				"user_profiles": [],
				"with_listener": []
		},
		"user_id": "test_user_1",
			"feature_flag_key": "feature_5",
			"attributes": {
			"attr_5": "testvalue"
		}
	}`)

	var params isFeatureEnabledRequestParams
	err := json.Unmarshal(strBytes, &params)
	if err != nil {
		log.Fatal(err)
	}

	var requestBodyMap map[string]*json.RawMessage

	err = json.Unmarshal(strBytes, &requestBodyMap)
	if err != nil {
		log.Fatal(err)
	}

	var fscCtx Context
	err = json.Unmarshal(*requestBodyMap["context"], &fscCtx)
	if err != nil {
		log.Fatal(err)
	}

	var datafileDir = path.Join(os.Getenv("DATAFILES_DIR"), fscCtx.Datafile)

	datafile, err := ioutil.ReadFile(datafileDir)
	if err != nil {
		log.Fatal(err)
	}

	optlyClient := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	user := entities.UserContext{
		ID:         params.UserID,
		Attributes: params.Attributes,
	}

	// Creates a default, canceleable context
	notificationCenter := notification.NewNotificationCenter()
	decisionService := decision.NewCompositeService(notificationCenter)

	clientOptions := client.Options{
		DecisionService: decisionService,
	}
	clientApp, err := optlyClient.ClientWithOptions(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	clientApp.IsFeatureEnabled(params.FeatureKey, user)
}

var RunMemProfile = "false"
var RunCPUProfile = "false"

func main() {

	if RunMemProfile == "true" || RunCPUProfile == "true" {

		const RUN_NUMBER = 50
		if RunMemProfile == "true" {
			defer profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.MemProfileRate(1)).Stop()
		} else if RunCPUProfile == "true" {
			defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
		}

		for i := 0; i < RUN_NUMBER; i++ {
			stressTest()
		}
	}

}
