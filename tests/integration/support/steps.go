package support

import (
	"fmt"
	"strconv"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/facebookarchive/subset"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"gopkg.in/yaml.v3"
)

// ScenarioCtx holds both apiOptions and apiResponse for a scenario
type ScenarioCtx struct {
	apiOptions    models.APIOptions
	apiResponse   models.APIResponse
	clientWrapper ClientWrapper
}

// TheDatafileIs represents a step in the feature file
func (c *ScenarioCtx) TheDatafileIs(datafileName string) error {
	c.clientWrapper = NewClientWrapper(datafileName)
	return nil
}

// ListenerIsAdded represents a step in the feature file
func (c *ScenarioCtx) ListenerIsAdded(numberOfListeners int, ListenerName string) error {
	if c.apiOptions.Listeners == nil {
		c.apiOptions.Listeners = make(map[string]int)
	}
	c.apiOptions.Listeners[ListenerName] = numberOfListeners
	return nil
}

// IsCalledWithArguments represents a step in the feature file
func (c *ScenarioCtx) IsCalledWithArguments(apiName string, arguments *gherkin.DocString) error {
	c.apiOptions.APIName = apiName
	c.apiOptions.Arguments = arguments.Content
	err := c.clientWrapper.InvokeAPI(c.apiOptions, &c.apiResponse)
	//Reset listeners so that same listener is not added twice for a scenario
	c.apiOptions.Listeners = nil
	if err == nil {
		return nil
	}
	return fmt.Errorf("invalid api or arguments")
}

// TheResultShouldBeString represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldBeString(result string) error {
	if c.apiResponse.Type != "" && c.apiResponse.Type != entities.String {
		return fmt.Errorf("incorrect type")
	}
	if result == c.apiResponse.Result {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeInteger represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldBeInteger(result int) error {
	if c.apiResponse.Type != "" && c.apiResponse.Type != entities.Integer {
		return fmt.Errorf("incorrect type")
	}
	responseIntValue := c.apiResponse.Result
	if stringIntValue, ok := c.apiResponse.Result.(string); ok {
		responseIntValue, _ = strconv.Atoi(stringIntValue)
	}
	if result == responseIntValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeFloat represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldBeFloat(lv, rv int) error {
	floatvalue, _ := strconv.ParseFloat(fmt.Sprintf("%v.%v", lv, rv), 64)
	if c.apiResponse.Type != "" && c.apiResponse.Type != entities.Double {
		return fmt.Errorf("incorrect type")
	}
	responseFloatValue := c.apiResponse.Result
	if stringFloatValue, ok := c.apiResponse.Result.(string); ok {
		responseFloatValue, _ = strconv.ParseFloat(stringFloatValue, 64)
	}
	if floatvalue == responseFloatValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeTypedBoolean represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldBeTypedBoolean(result string) error {
	boolValue, _ := strconv.ParseBool(result)
	if c.apiResponse.Type != "" && c.apiResponse.Type != entities.Boolean {
		return fmt.Errorf("incorrect type")
	}
	responseBoolValue := c.apiResponse.Result
	if stringBoolValue, ok := c.apiResponse.Result.(string); ok {
		responseBoolValue, _ = strconv.ParseBool(stringBoolValue)
	}
	if boolValue == responseBoolValue {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeBoolean represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldBeBoolean() error {
	boolValue, _ := strconv.ParseBool(c.apiResponse.Result.(string))
	if boolValue == false {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldMatchList represents a step in the feature file
func (c *ScenarioCtx) TheResultShouldMatchList(list string) error {
	if c.apiResponse.Result == list {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// InTheResponseKeyShouldBeObject represents a step in the feature file
func (c *ScenarioCtx) InTheResponseKeyShouldBeObject(argumentType, value string) error {
	switch argumentType {
	case "listener_called":
		if value == "NULL" && c.apiResponse.ListenerCalled == nil {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("incorrect listeners called")
}

// InTheResponseShouldMatch represents a step in the feature file
func (c *ScenarioCtx) InTheResponseShouldMatch(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case "listener_called":
		var requestListenersCalled []models.DecisionListener

		if err := yaml.Unmarshal([]byte(value.Content), &requestListenersCalled); err != nil {
			break
		}
		if subset.Check(requestListenersCalled, c.apiResponse.ListenerCalled) {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// InTheResponseShouldHaveEachOneOfThese represents a step in the feature file
func (c *ScenarioCtx) InTheResponseShouldHaveEachOneOfThese(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case "listener_called":
		var requestListenersCalled []models.DecisionListener

		if err := yaml.Unmarshal([]byte(value.Content), &requestListenersCalled); err != nil {
			break
		}
		if subset.Check(requestListenersCalled, c.apiResponse.ListenerCalled) {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// ThereAreNoDispatchedEvents represents a step in the feature file
func (c *ScenarioCtx) ThereAreNoDispatchedEvents() error {
	if len(c.apiResponse.ListenerCalled) == 0 {
		return nil
	}
	return fmt.Errorf("listenersCalled should be empty")
}

// DispatchedEventsPayloadsInclude represents a step in the feature file
func (c *ScenarioCtx) DispatchedEventsPayloadsInclude(value *gherkin.DocString) error {

	config, err := c.clientWrapper.Client.GetProjectConfig()
	if err != nil {
		return fmt.Errorf("Invalid Project Config")
	}
	requestedBatchEvents, err := getDispatchedEventsFromYaml(value.Content, config)
	if err != nil {
		return fmt.Errorf("Invalid request for dispatched Events")
	}
	dispatchedEvents := c.clientWrapper.EventDispatcher.(optlyplugins.EventReceiver).GetEvents()
	if subset.Check(requestedBatchEvents, dispatchedEvents) {
		return nil
	}
	return fmt.Errorf("DispatchedEvents not equal")
}

// Reset clears all data before each scenario
func (c *ScenarioCtx) Reset() {
	c.apiOptions = models.APIOptions{}
	c.apiResponse = models.APIResponse{}
	c.clientWrapper = ClientWrapper{}
}
