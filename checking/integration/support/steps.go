package support

import (
	"fmt"
	"strconv"

	"github.com/optimizely/go-sdk/optimizely/event"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/optimizely/go-sdk/checking/integration/optimizely/datamodels"
	"github.com/optimizely/go-sdk/checking/integration/optimizely/eventdispatcher"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"gopkg.in/yaml.v3"
)

// Context holds both request and response for a scenario
type Context struct {
	requestParams  datamodels.RequestParams
	responseParams datamodels.ResponseParams
}

// TheDatafileIs represents a step in the feature file
func (c *Context) TheDatafileIs(datafileName string) error {
	c.requestParams.DatafileName = datafileName
	return nil
}

// ListenerIsAdded represents a step in the feature file
func (c *Context) ListenerIsAdded(numberOfListeners int, ListenerName string) error {
	if c.requestParams.Listeners == nil {
		c.requestParams.Listeners = make(map[string]int)
	}
	c.requestParams.Listeners[ListenerName] = numberOfListeners
	return nil
}

// IsCalledWithArguments represents a step in the feature file
func (c *Context) IsCalledWithArguments(arg1 string, arg2 *gherkin.DocString) error {
	c.requestParams.APIName = arg1
	c.requestParams.Arguments = arg2.Content
	result, err := ProcessRequest(&c.requestParams)
	if err == nil {
		c.responseParams.Result = result.Result
		c.responseParams.ListenerCalled = result.ListenerCalled
		return nil
	}
	return fmt.Errorf("invalid api or arguments")
}

// TheResultShouldBeString represents a step in the feature file
func (c *Context) TheResultShouldBeString(arg1 string) error {
	if c.responseParams.Type != "" && c.responseParams.Type != entities.String {
		return fmt.Errorf("incorrect type")
	}
	if arg1 == c.responseParams.Result {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeInteger represents a step in the feature file
func (c *Context) TheResultShouldBeInteger(arg1 int) error {
	if c.responseParams.Type != "" && c.responseParams.Type != entities.Integer {
		return fmt.Errorf("incorrect type")
	}
	responseIntValue := c.responseParams.Result
	if stringIntValue, ok := c.responseParams.Result.(string); ok {
		responseIntValue, _ = strconv.Atoi(stringIntValue)
	}
	if arg1 == responseIntValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeFloat represents a step in the feature file
func (c *Context) TheResultShouldBeFloat(arg1, arg2 int) error {
	floatvalue, _ := strconv.ParseFloat(fmt.Sprintf("%v.%v", arg1, arg2), 64)
	if c.responseParams.Type != "" && c.responseParams.Type != entities.Double {
		return fmt.Errorf("incorrect type")
	}
	responseFloatValue := c.responseParams.Result
	if stringFloatValue, ok := c.responseParams.Result.(string); ok {
		responseFloatValue, _ = strconv.ParseFloat(stringFloatValue, 64)
	}
	if floatvalue == responseFloatValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeBoolean represents a step in the feature file
func (c *Context) TheResultShouldBeBoolean(arg1 string) error {
	boolValue, _ := strconv.ParseBool(arg1)
	if c.responseParams.Type != "" && c.responseParams.Type != entities.Boolean {
		return fmt.Errorf("incorrect type")
	}
	responseBoolValue := c.responseParams.Result
	if stringBoolValue, ok := c.responseParams.Result.(string); ok {
		responseBoolValue, _ = strconv.ParseBool(stringBoolValue)
	}
	if boolValue == responseBoolValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeFalse represents a step in the feature file
func (c *Context) TheResultShouldBeFalse() error {
	boolValue, _ := strconv.ParseBool(c.responseParams.Result.(string))
	if boolValue == false {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// InTheResponseKeyShouldBeObject represents a step in the feature file
func (c *Context) InTheResponseKeyShouldBeObject(arg1, arg2 string) error {
	switch arg1 {
	case "listener_called":
		if arg2 == "NULL" && c.responseParams.ListenerCalled == nil {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("incorrect listeners called")
}

// InTheResponseShouldMatch represents a step in the feature file
func (c *Context) InTheResponseShouldMatch(arg1 string, arg2 *gherkin.DocString) error {
	switch arg1 {
	case "listener_called":
		var requestListenersCalled []datamodels.DecisionListenerModel
		if err := yaml.Unmarshal([]byte(arg2.Content), &requestListenersCalled); err != nil {
			break
		}
		if Check(requestListenersCalled, c.responseParams.ListenerCalled) {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", arg1)
}

// ThereAreNoDispatchedEvents represents a step in the feature file
func (c *Context) ThereAreNoDispatchedEvents() error {
	if len(c.responseParams.ListenerCalled) == 0 {
		return nil
	}
	return fmt.Errorf("listenersCalled should be empty")
}

// DispatchedEventsPayloadsInclude represents a step in the feature file
func (c *Context) DispatchedEventsPayloadsInclude(arg1 *gherkin.DocString) error {
	requestedBatchEvents := []event.Batch{}
	if err := yaml.Unmarshal([]byte(arg1.Content), &requestedBatchEvents); err != nil {
		return fmt.Errorf("Invalid request for dispatched Events")
	}
	dispatchedEvents := c.requestParams.DependencyModel.Dispatcher.(eventdispatcher.EventReceiver).GetEvents()
	if Check(requestedBatchEvents, dispatchedEvents) {
		return nil
	}
	return fmt.Errorf("DispatchedEvents not equal")
}

// Reset clears all data before each scenario
func (c *Context) Reset() {
	c.requestParams = datamodels.RequestParams{}
	c.responseParams = datamodels.ResponseParams{}
}
