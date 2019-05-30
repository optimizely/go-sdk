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

	}

	// test
	conditions := "[\"and\", [\"or\",  [\"or\", [\"or\", {\"type\": \"custom_attribute\", \"name\": \"string_attribute\", \"value\": \"exact_match\"}], \"or\", [\"and\",{\"sd\": 1}]]], [\"or\", [\"or\", {\"type\": \"custom_attribute1\", \"name\": \"string_attribute\", \"value\": \"exact_match\"}]]]"

	var v interface{}
	json.Unmarshal([]byte(conditions), &v)
	typedAudience := project_config.Audience{Conditions: v}

	typedAudience.PopulateTypedConditions()

	tree := typedAudience.ConditionTree

	//// first level
	fmt.Println("Tree:", tree.Root.Nodes[0].SimpleCondition)
	fmt.Println("Tree:", tree.Root.Nodes[1].Element)
	fmt.Println("Tree:", tree.Root.Nodes[2].Element)
	fmt.Println(len(tree.Root.Nodes))

	// second level
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[0].SimpleCondition)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Element)
	fmt.Println(len(tree.Root.Nodes[1].Nodes))

	fmt.Println("Tree:", tree.Root.Nodes[2].Nodes[0].SimpleCondition)
	fmt.Println("Tree:", tree.Root.Nodes[2].Nodes[1].Element)
	fmt.Println(len(tree.Root.Nodes[2].Nodes))

	// third level
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[0].SimpleCondition)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[1].Element)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[2].SimpleCondition)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[3].Element)
	fmt.Println(len(tree.Root.Nodes[1].Nodes[1].Nodes))

	fmt.Println("Tree:", tree.Root.Nodes[2].Nodes[1].Nodes[0].Element)
	fmt.Println("Tree:", tree.Root.Nodes[2].Nodes[1].Nodes[1].ComplexCondition)
	fmt.Println(len(tree.Root.Nodes[2].Nodes[1].Nodes))

	//4th level
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[1].Nodes[0].Element)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[1].Nodes[1].ComplexCondition)
	fmt.Println(len(tree.Root.Nodes[1].Nodes[1].Nodes[1].Nodes))

	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[3].Nodes[0].Element)
	fmt.Println("Tree:", tree.Root.Nodes[1].Nodes[1].Nodes[3].Nodes[1].ComplexCondition)
	fmt.Println(len(tree.Root.Nodes[1].Nodes[1].Nodes[3].Nodes))

}
