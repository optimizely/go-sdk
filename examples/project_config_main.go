package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"go-sdk/optimizely/project-config"
)

func main() {
	projectConfig := project_config.ProjectConfig{}

	parser := func(fileName string) {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = json.Unmarshal(file, &projectConfig)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(projectConfig)
	}

	parser("examples/test1.json")
	parser("examples/test.json")
}
