package project_config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Attribute struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type Audience struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"conditions"`

	LogicalOperators []string    `json:"-"`
	Criteria         []Criterion `json:"-"`
}

type Variation struct {
	ID             string   `json:"id"`
	Variables      []string `json:"variables"`
	Key            string   `json:"key"`
	FeatureEnabled bool     `json:"featureEnabled"`
}

type Event struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
}

type TrafficAllocation struct {
	EntityID   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

type Experiment struct {
	ID                string              `json:"id"`
	LayerID           string              `json:"layerId"`
	Key               string              `json:"key"`
	Status            string              `json:"status"`
	Variations        []Variation         `json:"variations"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
	AudienceIds       []string            `json:"audienceIds"`
	ForcedVariations  map[string]string   `json:"forcedVariations"`
	GroupId           string              `json:"-"`
	GroupPolicy       string              `json:"-"`
}

type FeatureFlag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

type Group struct {
	ID                string              `json:"id"`
	Policy            string              `json:"policy"`
	Experiments       []Experiment        `json:"experiments"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
}

type Rollout struct {
	ID          string       `json:"id"`
	Experiments []Experiment `json:"experiments"`
}

//ProjectConfig is the main thing for project configuration.
type ProjectConfig struct {
	Version        string        `json:"version"`
	Rollouts       []Rollout     `json:"rollouts"`
	TypedAudiences []Audience    `json:"typedAudiences"`
	AnonymizeIP    bool          `json:"anonymizeIP"`
	ProjectID      string        `json:"projectId"`
	Variables      []string      `json:"variables"`
	FeatureFlags   []FeatureFlag `json:"featureFlags"`
	Experiments    []Experiment  `json:"experiments"`
	Audiences      []Audience    `json:"audiences"`
	Groups         []Group       `json:"groups"`
	Attributes     []Attribute   `json:"attributes"`
	BotFiltering   bool          `json:"botFiltering"`
	AccountID      string        `json:"accountId"`
	Events         []Event       `json:"events"`
	Revision       string        `json:"revision"`
}

type Criterion struct {
	Name  string      `json:"name"`
	Match string      `json:"match"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
type Conditions struct {
	LogicalOperators []string    `json:"-"`
	Criteria         []Criterion `json:"-"`
}

func (s *Audience) PopulateTypedConditions() error {

	value := reflect.ValueOf(&s.Conditions)
	visited := make(map[interface{}]bool)
	var retErr error

	var populateConditions func(v reflect.Value)
	populateConditions = func(v reflect.Value) {

		// Drill down through pointers and interfaces to get a value we can print.
		for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if v.Kind() == reflect.Ptr {
				// Check for recursive data
				if visited[v.Interface()] {
					return
				}
				visited[v.Interface()] = true
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				populateConditions(v.Index(i))
			}
		case reflect.Struct:
			t := v.Type()
			for i := 0; i < t.NumField(); i++ {
				populateConditions(v.Field(i))
			}
		case reflect.Invalid:
			fmt.Printf("nil\n")
		case reflect.String:
			s.LogicalOperators = append(s.LogicalOperators, v.Interface().(string))
		case reflect.Map:
			jsonbody, err := json.Marshal(v.Interface())
			if err != nil {
				// do error check
				fmt.Println(err)
				retErr = err
				return
			}
			criterias := Criterion{}
			if err := json.Unmarshal(jsonbody, &criterias); err != nil {
				// do error check
				fmt.Println(err)
				retErr = err
				return
			}
			s.Criteria = append(s.Criteria, criterias)
		default:
			fmt.Printf("%v\n", v.Interface())
			retErr = fmt.Errorf("Cannot parse %v\n", v.Interface())
		}
	}

	populateConditions(value)
	return retErr
}
