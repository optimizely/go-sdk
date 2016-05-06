package optimizely

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

var PROJECT_CONFIG_LINK_TEMPLATE = "https://cdn.optimizely.com/json/%v.json"

// FetchConfig retrieves a JSON file from the Optimizely CDN
// and returns it accordingly
func FetchProjectConfig(project_id string) (ProjectConfig, error) {
	var project_config = ProjectConfig{}
	var project_config_url = fmt.Sprintf(PROJECT_CONFIG_LINK_TEMPLATE, project_id)
	resp, err := http.Get(project_config_url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code of %v when fetching project config", resp.StatusCode)
		return project_config, errors.New(fmt.Sprintf("Error fetching config, Optimizely returned %v", resp.StatusCode))
	}
	if err != nil {
		log.Printf("Error fetching project config: %v", err)
		return project_config, errors.New("Error fetching config")
	}

	err = json.NewDecoder(resp.Body).Decode(&project_config)
	if err != nil {
		log.Printf("Error reading JSON response into a buffer")
		return project_config, errors.New("Cannot parse JSON file")
	}
	return project_config, nil
}
