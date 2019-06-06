package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

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
	var m1, m2 runtime.MemStats

	runtime.ReadMemStats(&m1)
	start := time.Now()
	for i := 0; i < 1; i++ {

		conditions := "[\"and\", [\"or\",  [\"or\", [\"or\", {\"type\": \"custom_attribute\", \"name\": \"string_attribute\", \"value\": \"exact_match\"}], \"or\", [\"and\",{\"sd\": 1}]]], [\"or\", [\"or\", {\"type\": \"custom_attribute1\", \"name\": \"string_attribute\", \"value\": \"exact_match\"}]]]"

		// test conditions
		conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" } ] ] ]"
		conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"lt\", \"value\": true } ] ], [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"gt\", \"value\": 41 } ] ] ]"
		//conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" }, { \"type\": \"custom_attribute\", \"name\": \"s_bar\", \"match\": \"exact\", \"value\": \"bar\" } ] ] ]"

		var v interface{}
		json.Unmarshal([]byte(conditions), &v)
		typedAudience := project_config.Audience{Conditions: v}

		typedAudience.PopulateTypedConditions()

		tree := typedAudience.ConditionTree

		m := map[string]interface{}{}

		//m["s_foo"] = "foo"
		//m["s_bar"] = "dbar"
		//m["i__not42"] = 44.0
		m["i_42"] = 41.2

		b, _ := json.Marshal(m)
		project_config.Evaluate(tree.Root, m)
		//project_config.Evaluate(tree.Root, m)
		fmt.Println("conditions:", conditions)
		fmt.Println("input:", string(b))
		fmt.Println("evaluation:", project_config.Evaluate(tree.Root, m))

	}

	fmt.Println("\nTotal time:", time.Now().Sub(start))
	runtime.ReadMemStats(&m2)
	fmt.Println("Bytes allocated:", m2.TotalAlloc-m1.TotalAlloc)

}
