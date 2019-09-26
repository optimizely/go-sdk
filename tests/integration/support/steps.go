package support

import (
	"fmt"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/optimizely/go-sdk/Tests/integration/optimizely/datamodels"
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
	c.requestParams.ApiName = arg1
	c.requestParams.Arguments = arg2.Content
	result, err := ProcessRequest(c.requestParams)
	if err == nil {
		c.responseParams.Result = result.Result
		return nil
	}
	return fmt.Errorf("invalid api or arguments")
}

// TheResultShouldBe represents a step in the feature file
func (c *Context) TheResultShouldBe(arg1 string) error {
	if arg1 == c.responseParams.Result {
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

// ThereAreNoDispatchedEvents represents a step in the feature file
func (c *Context) ThereAreNoDispatchedEvents() error {
	if len(c.responseParams.ListenerCalled) == 0 {
		return nil
	}
	return fmt.Errorf("listenersCalled should be empty")
}
