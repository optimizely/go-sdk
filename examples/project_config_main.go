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
		conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"s_foo\", \"match\": \"exact\", \"value\": \"foo\" }, { \"type\": \"custom_attribute\", \"name\": \"s_bar\", \"match\": \"exact\", \"value\": \"bar\" } ] ] ]"
		conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"lt\", \"value\": 43 } ,  \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_43\", \"match\": \"gt\", \"value\": 41 }, { \"type\": \"custom_attribute\", \"name\": \"i_47\", \"match\": \"gt\", \"value\": 41 }, { \"type\": \"custom_attribute\", \"name\": \"i_48\", \"match\": \"gt\", \"value\": 41 } ] ]]]"

		conditions = "[ \"and\", [ \"or\", [ \"not\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"lt\", \"value\": 43 } ,  [\"pawel\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_43\", \"match\": \"gt\", \"value\": 41 }, " +
			"{ \"type\": \"custom_attribute\", \"name\": \"i_47\", \"match\": \"gt\", \"value\": 41 }, { \"type\": \"custom_attribute\", \"name\": \"i_48\", \"match\": \"gt\", \"value\": 41 }, [ \"or\", [ \"not\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"lt\", \"value\": 43 } ]] ] ]]]]"

		//conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"exact\", \"value\": 23.0} ] ] ]"

		//conditions = "[ \"and\", [ \"or\", [ \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"exact\", \"value\": 23.0} ,  \"or\", { \"type\": \"custom_attribute\", \"name\": \"i_42\", \"match\": \"exact\", \"value\": 23.0}] ] ]"

		var v interface{}
		json.Unmarshal([]byte(conditions), &v)
		typedAudience := project_config.Audience{Conditions: v}

		typedAudience.PopulateNodeConditions()

		tree := typedAudience.ConditionTree

		m := map[string]interface{}{}

		//m["s_foo"] = "foo"
		//m["s_bar"] = "not_bar"
		m["i_42"] = 42.9999999
		m["i_43"] = 41.01

		b, _ := json.Marshal(m)

		project_config.Evaluate(tree.Root, m)
		fmt.Println("conditions:", conditions)
		fmt.Println("input:", string(b))
		fmt.Println("evaluation:", project_config.Evaluate(tree.Root, m))

	}

	fmt.Println("\nTotal time:", time.Now().Sub(start))
	runtime.ReadMemStats(&m2)
	fmt.Println("Bytes allocated:", m2.TotalAlloc-m1.TotalAlloc)

}
