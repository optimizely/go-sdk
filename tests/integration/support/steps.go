/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

package support

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"github.com/optimizely/subset"
	"gopkg.in/yaml.v3"
)

// ScenarioCtx holds both apiOptions and apiResponse for a scenario.
type ScenarioCtx struct {
	apiOptions    models.APIOptions
	apiResponse   models.APIResponse
	clientWrapper ClientWrapper
}

// TheDatafileIs defines a datafileName to initialize the client with.
func (c *ScenarioCtx) TheDatafileIs(datafileName string) error {
	c.clientWrapper = NewClientWrapper(datafileName)
	return nil
}

// ListenerIsAdded defines the listeners to be added to the client.
func (c *ScenarioCtx) ListenerIsAdded(numberOfListeners int, ListenerName string) error {
	if c.apiOptions.Listeners == nil {
		c.apiOptions.Listeners = make(map[string]int)
	}
	c.apiOptions.Listeners[ListenerName] = numberOfListeners
	return nil
}

// IsCalledWithArguments calls an SDK API with arguments.
func (c *ScenarioCtx) IsCalledWithArguments(apiName string, arguments *gherkin.DocString) error {
	c.apiOptions.APIName = apiName
	c.apiOptions.Arguments = arguments.Content

	// Clearing old state of response, eventdispatcher and decision service
	c.apiResponse = models.APIResponse{}
	// Only required for unsessioned tests
	c.clientWrapper.DecisionService.(*optlyplugins.TestCompositeService).ClearListenersCalled()
	c.clientWrapper.EventDispatcher.(*optlyplugins.ProxyEventDispatcher).ClearEvents()

	response, err := c.clientWrapper.InvokeAPI(c.apiOptions)
	c.apiResponse = response
	//Reset listeners so that same listener is not added twice for a scenario
	c.apiOptions.Listeners = nil
	if err == nil {
		return nil
	}
	return fmt.Errorf("invalid api or arguments")
}

// TheResultShouldBeString checks that the result is of type string with the given value.
func (c *ScenarioCtx) TheResultShouldBeString(result string) error {
	if c.apiResponse.Type != "" && c.apiResponse.Type != entities.String {
		return fmt.Errorf("incorrect type")
	}
	if result == c.apiResponse.Result {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldBeInteger checks that the result is of type integer with the given value.
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

// TheResultShouldBeFloat checks that the result is of type double with the given value.
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

// TheResultShouldBeTypedBoolean checks that the result is of type boolean with the given value.
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

// TheResultShouldBeBoolean checks that the result is equal to the given boolean value.
func (c *ScenarioCtx) TheResultShouldBeBoolean() error {
	boolValue, _ := strconv.ParseBool(c.apiResponse.Result.(string))
	if boolValue == false {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// TheResultShouldMatchList checks that the result equals to the provided list.
func (c *ScenarioCtx) TheResultShouldMatchList(list string) error {
	expectedList := strings.Split(list, ",")
	if actualList, ok := c.apiResponse.Result.([]string); ok && compareStringSlice(expectedList, actualList) {
		return nil
	}
	return fmt.Errorf("incorrect result")
}

// InTheResponseKeyShouldBeObject checks that the response object contains a property with given value.
func (c *ScenarioCtx) InTheResponseKeyShouldBeObject(argumentType, value string) error {
	switch argumentType {
	case models.KeyListenerCalled:
		if value == "NULL" && c.apiResponse.ListenerCalled == nil {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("incorrect listeners called")
}

// InTheResponseShouldMatch checks that the response object contains a property with matching value.
func (c *ScenarioCtx) InTheResponseShouldMatch(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
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

// ResponseShouldHaveThisExactlyNTimes checks that the response object has the given property exactly N times.
func (c *ScenarioCtx) ResponseShouldHaveThisExactlyNTimes(argumentType string, count int, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
		var requestListenersCalled []models.DecisionListener
		if err := yaml.Unmarshal([]byte(value.Content), &requestListenersCalled); err != nil {
			break
		}
		listener := requestListenersCalled[0]
		expectedListenersArray := []models.DecisionListener{}
		for i := 0; i < count; i++ {
			expectedListenersArray = append(expectedListenersArray, listener)
		}
		if subset.Check(expectedListenersArray, c.apiResponse.ListenerCalled) {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// InTheResponseShouldHaveEachOneOfThese checks that the response object contains each of the provided properties.
func (c *ScenarioCtx) InTheResponseShouldHaveEachOneOfThese(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
		var requestListenersCalled []models.DecisionListener

		if err := yaml.Unmarshal([]byte(value.Content), &requestListenersCalled); err != nil {
			break
		}

		found := false
		for _, expectedListener := range requestListenersCalled {
			found = false
			for _, actualListener := range c.apiResponse.ListenerCalled {
				if subset.Check(expectedListener, actualListener) {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("%v", expectedListener)
				break
			}
		}
		if found {
			return nil
		}
		break
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// TheNumberOfDispatchedEventsIs checks the count of the dispatched events to be equal to the given value.
func (c *ScenarioCtx) TheNumberOfDispatchedEventsIs(count int) error {
	dispatchedEvents := c.clientWrapper.EventDispatcher.(optlyplugins.EventReceiver).GetEvents()
	if len(dispatchedEvents) == count {
		return nil
	}
	return fmt.Errorf("dispatchedEvents count not equal")
}

// ThereAreNoDispatchedEvents checks the dispatched events count to be empty.
func (c *ScenarioCtx) ThereAreNoDispatchedEvents() error {
	dispatchedEvents := c.clientWrapper.EventDispatcher.(optlyplugins.EventReceiver).GetEvents()
	if len(dispatchedEvents) == 0 {
		return nil
	}
	return fmt.Errorf("dispatchedEvents should be empty but received %d events", len(dispatchedEvents))
}

// DispatchedEventsPayloadsInclude checks dispatched events to contain the given events.
func (c *ScenarioCtx) DispatchedEventsPayloadsInclude(value *gherkin.DocString) error {

	config, err := c.clientWrapper.Client.GetProjectConfig()
	if err != nil {
		return fmt.Errorf("Invalid Project Config")
	}
	expectedBatchEvents, err := getDispatchedEventsMapFromYaml(value.Content, config)
	if err != nil {
		return fmt.Errorf("Invalid request for dispatched Events")
	}

	eventsReceived := c.clientWrapper.EventDispatcher.(optlyplugins.EventReceiver).GetEvents()
	eventsReceivedJSON, err := json.Marshal(eventsReceived)
	if err != nil {
		return fmt.Errorf("Invalid response for dispatched Events")
	}
	var actualBatchEvents []map[string]interface{}
	if err := json.Unmarshal(eventsReceivedJSON, &actualBatchEvents); err != nil {
		return fmt.Errorf("Invalid response for dispatched Events")
	}
	if subset.Check(expectedBatchEvents, actualBatchEvents) {
		return nil
	}
	return fmt.Errorf("DispatchedEvents not equal")
}

// PayloadsOfDispatchedEventsDontIncludeDecisions checks dispatched events to contain no decisions.
func (c *ScenarioCtx) PayloadsOfDispatchedEventsDontIncludeDecisions() error {
	dispatchedEvents := c.clientWrapper.EventDispatcher.(optlyplugins.EventReceiver).GetEvents()

	for _, event := range dispatchedEvents {
		for _, visitor := range event.Visitors {
			for _, snapshot := range visitor.Snapshots {
				if len(snapshot.Decisions) > 0 {
					return fmt.Errorf("dispatched events should not include decisions")
				}
			}
		}
	}
	return nil
}

// Reset clears all data before each scenario
func (c *ScenarioCtx) Reset() {
	c.apiOptions = models.APIOptions{}
	c.apiResponse = models.APIResponse{}
	c.clientWrapper = ClientWrapper{}
}
