package project_config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

//Attribute has atribute info
type Attribute struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

//Audience has audience info
type Audience struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"conditions"`

	ConditionTree Tree `json:"-"`
}

//Variation has variation info
type Variation struct {
	ID             string   `json:"id"`
	Variables      []string `json:"variables"`
	Key            string   `json:"key"`
	FeatureEnabled bool     `json:"featureEnabled"`
}

//Event has event info
type Event struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
}

//TrafficAllocation has trafficAllocation info
type TrafficAllocation struct {
	EntityID   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

//Experiment has experiment info
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

//FeatureFlag has feature flagging info
type FeatureFlag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIDs []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

//Group has group info
type Group struct {
	ID                string              `json:"id"`
	Policy            string              `json:"policy"`
	Experiments       []Experiment        `json:"experiments"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
}

// Rollout has rollout info
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

//Condition has condition info
type Condition struct {
	Name  string      `json:"name"`
	Match string      `json:"match"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

//Node in a condition tree
type Node struct {
	Element   interface{}
	Condition Condition
	Operator  string

	Left  *Node
	Right *Node
}

//Tree is used to create condition tree
type Tree struct {
	Root *Node
}

func nodeEvaluator(n *Node, m map[string]interface{}) bool {

	any := []bool{}

	int2float := func(value interface{}) interface{} {
		var floatType = reflect.TypeOf(float64(0))
		v := reflect.ValueOf(value)
		v = reflect.Indirect(v)

		if v.Type().String() == "float64" || v.Type().ConvertibleTo(floatType) {
			value = v.Convert(floatType).Float()
		}

		return value
	}

	for name, value := range m {

		value = int2float(value)
		n.Condition.Value = int2float(n.Condition.Value)

		valueType := reflect.TypeOf(value)
		conditionValueType := reflect.TypeOf(n.Condition.Value)
		//fmt.Println(valueType, conditionValueType)

		if n.Condition.Match == "exact" {
			//fmt.Println(reflect.ValueOf(value), reflect.ValueOf(n.Condition.Value))
			if valueType == conditionValueType {
				if valueType.String() == "float64" {
					any = append(any, name == n.Condition.Name && reflect.ValueOf(value).Float() == reflect.ValueOf(n.Condition.Value).Float())
				} else if valueType.String() == "string" {
					any = append(any, name == n.Condition.Name && reflect.DeepEqual(reflect.ValueOf(value).String(), reflect.ValueOf(n.Condition.Value).String()))
				} else {
					any = append(any, name == n.Condition.Name && reflect.DeepEqual(reflect.ValueOf(value), reflect.ValueOf(n.Condition.Value)))
				}
			}
		} else if n.Condition.Match == "gt" {
			if valueType == conditionValueType {
				any = append(any, name == n.Condition.Name && reflect.ValueOf(value).Float() > reflect.ValueOf(n.Condition.Value).Float())
			}

			//fmt.Println(reflect.ValueOf(value), reflect.ValueOf(n.Condition.Value))

		} else if n.Condition.Match == "lt" {

			if valueType == conditionValueType {
				//fmt.Println(reflect.ValueOf(value), reflect.ValueOf(n.Condition.Value), reflect.ValueOf(value).Float() < reflect.ValueOf(n.Condition.Value).Float())
				any = append(any, name == n.Condition.Name && reflect.ValueOf(value).Float() < reflect.ValueOf(n.Condition.Value).Float())
			}
		}
	}
	if len(any) == 0 {
		fmt.Println("log (just warning): types not matched - cannot evaluate one of the condition instances properly")
	}
	for _, x := range any {
		if x == true {
			return true
		}
	}
	return false
}

//Evaluate is the main function to evaluate an input condition
func Evaluate(root *Node, m map[string]interface{}) bool {

	if root.Left == nil && root.Right == nil {
		//fmt.Println(nodeEvaluator(root, m))
		return nodeEvaluator(root, m)
	}

	leftLogic := Evaluate(root.Left, m)

	var rightLogic bool
	if root.Right == nil {
		rightLogic = leftLogic
	} else {
		rightLogic = Evaluate(root.Right, m)
	}

	if root.Operator == "and" {
		//fmt.Println(rightLogic, leftLogic)
		return rightLogic && leftLogic
	}
	if root.Operator == "or" {
		//fmt.Println(rightLogic)
		return rightLogic || leftLogic
	}
	return false
}

//PopulateTypedConditions if used to build condition tree
// not sure if we can reuse it, if so  then it does not belong here
func (s *Audience) PopulateTypedConditions() error {

	value := reflect.ValueOf(&s.Conditions)
	visited := make(map[interface{}]bool)
	var retErr error

	s.ConditionTree = Tree{}
	s.ConditionTree.Root = &Node{Element: value.Interface()}

	var populateConditions func(v reflect.Value, root *Node)
	populateConditions = func(v reflect.Value, root *Node) {

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
			//fmt.Printf("%d elements\n", v.Len())
			for i := 0; i < v.Len(); i++ {
				//fmt.Printf("%s%d: ", prefix, i)
				n := &Node{Element: v.Index(i).Interface()}
				typedV := v.Index(i).Interface()
				switch typedV.(type) {
				case string:
					n.Operator = typedV.(string)
					root.Operator = n.Operator
					continue

				case map[string]interface{}:
					jsonbody, err := json.Marshal(typedV)
					if err != nil {
						// do error check
						fmt.Println(err)
						retErr = err
						return
					}
					condition := Condition{}
					if err := json.Unmarshal(jsonbody, &condition); err != nil {
						// do error check
						fmt.Println(err)
						retErr = err
						//return
					}

					n.Condition = condition
				}

				//root.Nodes = append(root.Nodes, n)
				if root.Left == nil {
					root.Left = n
				} else if root.Right == nil {
					root.Right = n
				} else {
					retErr = fmt.Errorf("Cannot populate tree nodes")
				}
				//fmt.Println("Node", n)
				populateConditions(v.Index(i), n)

			}
		}
	}

	populateConditions(value, s.ConditionTree.Root)
	return retErr
}
