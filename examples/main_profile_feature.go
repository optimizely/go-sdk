// to run the CPU profiling: go build -ldflags "-X main.RunCPUProfile=true" main.go && ./main
// to run the Mem profiling: go build -ldflags "-X main.RunMemProfile=true" main.go && ./main

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/notification"

	"github.com/pkg/profile"
)

// stressTest has everything that test app has. it is used to run profile
func stressTest() {
	/*
		For the test app, the biggest json file is used with 100 entities.
		DATAFILES_DIR has to be set to point to the path where 100_entities.json is located.
	*/

	var datafileDir = path.Join(os.Getenv("DATAFILES_DIR"), "100_entities.json")

	datafile, err := ioutil.ReadFile(datafileDir)
	if err != nil {
		log.Fatal(err)
	}

	optlyClient := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	user := entities.UserContext{
		ID: "test_user_1",
		Attributes: map[string]interface{}{
			"attr_5": "testvalue",
		},
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

	clientApp.IsFeatureEnabled("feature_5", user)
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
