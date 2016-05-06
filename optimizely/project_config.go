package optimizely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var PROJECT_CONFIG_LINK_TEMPLATE = "https://cdn.optimizely.com/json/%v.json"

// FetchConfig retrieves a JSON file from the Optimizely CDN
// and returns it accordingly
func FetchProjectConfig(project_id int) bytes.Buffer {
	var b bytes.Buffer

	var project_config_url = fmt.Sprintf(PROJECT_CONFIG_LINK_TEMPLATE, project_id)
	resp, err := http.Get(project_config_url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code of %v when fetching project config", resp.StatusCode)
	}
	if err != nil {
		log.Printf("Error fetching project config: %v", err)
		return b
	}

	_, err = b.ReadFrom(resp.Body)
	if err != nil {
		log.Printf("Error reading JSON response into a buffer")
	}
	return b
}

// DeserializeConfigBuffer takes a buffer containing a JSON response
// from the Optimizely CDN and converts it to a ProjectConfig entity
// which gets returned
func DeserializeConfigBuffer(b bytes.Buffer) ProjectConfig {
	var project_config = ProjectConfig{}
	r := bytes.NewReader(b.Bytes())
	err := json.NewDecoder(r).Decode(&project_config)
	if err != nil {
		log.Print(err)
		return project_config
	}
	return project_config
}
