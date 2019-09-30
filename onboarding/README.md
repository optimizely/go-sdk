## Set up tutorial

Set up Daily Deal in less than 1 minute.


Run this line to create a directory for the app:

```
mkdir optimizely_demo; cd optimizely_demo;
```

Install required dependencies:

```
go get github.com/optimizely/go-sdk/...
```

Create an empty text file called `app.go`. In most systems, you can use the command:

```
touch app.go
```

Open `app.go` in any text editor.

Paste this script into `app.go` and save it:
<br />
(Note: you'll want to change the `https://cdn.optimizely.com/onboarding/c7eafb4cc4c411e3a4d40de1d3a4ae5d.json` to a path to your datafile)
```
// Package main //
package main

import (
	"fmt"
	"time"

	"github.com/optimizely/go-sdk/optimizely/logging"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/entities"
)

var datafileURL = "https://cdn.optimizely.com/onboarding/c7eafb4cc4c411e3a4d40de1d3a4ae5d.json"

// Deal struct holds the info of a deal
type Deal struct {
	Text      string
	Enabled   bool
	DebugText string
}

var debugTextOn = "[DEBUG: Feature ON]"
var debugTextOff = "[DEBUG: Feature OFF]"

func getDailyDeal(optimizelyClient *client.OptimizelyClient, userContext entities.UserContext) (deal Deal, err error) {

	var enabled bool
	debugText := debugTextOff
	enabled, err = optimizelyClient.IsFeatureEnabled("purchase_option", userContext)

	if err != nil {
		fmt.Printf("Error in Feature Enabled: %s", err)
		return
	}
	text := "Daily deal: A bluetooth speaker for $99!"
	if enabled {
		text, err = optimizelyClient.GetFeatureVariableString("purchase_option", "message", userContext)
		debugText = debugTextOn
		if _, err := userContext.GetStringAttribute("purchasedItem"); err != nil {
			optimizelyClient.Track("purchase_item", userContext, nil)
		}
	}

	return Deal{Text: text, Enabled: enabled, DebugText: debugText}, err
}

func main() {

	logging.SetLogLevel(logging.LogLevelDebug)
	optimizelyFactory := &client.OptimizelyFactory{}

	requester := utils.NewHTTPRequester(datafileURL)

	optimizelyClient, err := optimizelyFactory.Client(client.PollingConfigManagerRequester(requester, 5*time.Second, nil))

	if err != nil {
		fmt.Printf("Error instantiating client: %s", err)
		return
	}

	visitors := []entities.UserContext{
		{ID: "alice"},
		{ID: "bob"},
		{ID: "charlie"},
		{ID: "don"},
		{ID: "eli"},
		{ID: "fabio"},
		{ID: "gary"},
		{ID: "helen"},
		{ID: "ian"},
		{ID: "jill"},
	}

	f := func() {

		fmt.Println("\n\nWelcome to Daily Deal, we have great deals!")
		fmt.Println("Let's see what the visitors experience!")

		experiences := []string{}
		var numOnVariations int
		frequencyMap := map[string]int{}
		total := len(visitors)

		for i, visitor := range visitors {
			deal, _ := getDailyDeal(optimizelyClient, visitor)
			if deal.Enabled {
				numOnVariations++
			}
			experiences = append(experiences, fmt.Sprintf("Visitor #%d: %s %v", i, deal.DebugText, deal.Text))
			frequencyMap[deal.Text]++
		}

		/**** Print results *****/

		for _, exp := range experiences {
			fmt.Println(exp)
		}

		if numOnVariations > 0 {
			fmt.Printf("\n%d out of %d visitors (~%f) had the feature enabled\n",
				numOnVariations,
				total,
				float64(numOnVariations)/float64(total)*100.0)

		}

		for key, value := range frequencyMap {
			perc := float64(value) / float64(total) * 100.0
			fmt.Printf("\n%d visitors (~%f%%) got the experience: '%s'", value, perc, key)
		}
		fmt.Println()
	}

	t := time.NewTicker(20 * time.Second)
	f()

	for range t.C {
		f()
	}
}

```

### Run the app

Now you're ready to run Daily Deal for the first time.

In the terminal, run: `go run app.go`.

You should see: `Daily deal: A bluetooth speaker for $99!`

Great! Daily Deal is running. Now you're ready to try out a feature flag in the app.
