package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"go-sdk/optimizely/project-config"
)

func main() {

	parser := func(fileName string) project_config.ProjectConfig {
		projectConfig := project_config.ProjectConfig{}
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return project_config.ProjectConfig{}
		}
		err = json.Unmarshal(file, &projectConfig)
		if err != nil {
			fmt.Println(err)
			return project_config.ProjectConfig{}
		}
		return projectConfig
	}

	projectConfig := parser("examples/test1.json")

	typedAudiences := projectConfig.TypedAudiences

	for _, audience := range typedAudiences {

		if err := audience.PopulateTypedConditions(); err != nil {
			fmt.Println(err)
		}
		fmt.Println(audience.Criteria)
		fmt.Println(audience.LogicalOperators)

	}
}
