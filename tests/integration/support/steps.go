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

	"github.com/optimizely/go-sdk/pkg/decision"

	"github.com/cucumber/gherkin-go"
	"github.com/cucumber/messages-go"
	"github.com/google/uuid"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins/userprofileservice"
	"github.com/optimizely/subset"
)

// ScenarioCtx holds both apiOptions and apiResponse for a scenario.
type ScenarioCtx struct {
	scenarioID    string
	apiOptions    models.APIOptions
	apiResponse   models.APIResponse
	clientWrapper *ClientWrapper
}

// TheDatafileIs defines a datafileName to initialize the client with.
func (c *ScenarioCtx) TheDatafileIs(datafileName string) error {
	c.apiOptions.DatafileName = datafileName
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

// TheUserProfileServiceIs defines UserProfileService type
func (c *ScenarioCtx) TheUserProfileServiceIs(upsType string) error {
	c.apiOptions.UserProfileServiceType = upsType
	return nil
}

// UserHasMappingInUserProfileService defines UserMapping in UPS
func (c *ScenarioCtx) UserHasMappingInUserProfileService(userID, experimentKey, variationKey string) error {
	if c.apiOptions.UPSMapping == nil {
		c.apiOptions.UPSMapping = make(map[string]map[string]string)
	}

	if profile, ok := c.apiOptions.UPSMapping[userID]; ok {
		profile[experimentKey] = variationKey
		c.apiOptions.UPSMapping[userID] = profile
	} else {
		c.apiOptions.UPSMapping[userID] = map[string]string{experimentKey: variationKey}
	}
	return nil
}

// IsCalledWithArguments calls an SDK API with arguments.
func (c *ScenarioCtx) IsCalledWithArguments(apiName string, arguments *gherkin.DocString) error {
	c.apiOptions.APIName = apiName
	c.apiOptions.Arguments = arguments.Content

	// Clearing old state of response, eventdispatcher and decision service
	c.apiResponse = models.APIResponse{}
	c.clientWrapper = GetInstance(c.apiOptions)
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
	if len(expectedList) == 1 && expectedList[0] == "[]" {
		expectedList = []string{}
	}
	if actualList, ok := c.apiResponse.Result.([]string); ok {
		if compareStringSlice(expectedList, actualList) {
			return nil
		}
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
	default:
		break
	}
	return fmt.Errorf("incorrect listeners called")
}

// InTheResponseShouldMatch checks that the response object contains a property with matching value.
func (c *ScenarioCtx) InTheResponseShouldMatch(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
		expectedListenersCalled := parseListeners(value.Content)
		if subset.Check(expectedListenersCalled, c.apiResponse.ListenerCalled) {
			return nil
		}
	default:
		break
	}

	return fmt.Errorf("response for %s not equal", argumentType)
}

// ResponseShouldHaveThisExactlyNTimes checks that the response object has the given property exactly N times.
func (c *ScenarioCtx) ResponseShouldHaveThisExactlyNTimes(argumentType string, count int, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
		requestListeners := parseListeners(value.Content)
		if len(requestListeners) > 0 {
			expectedListenersArray := []interface{}{}
			for i := 0; i < count; i++ {
				expectedListenersArray = append(expectedListenersArray, requestListeners[0])
			}
			if subset.Check(expectedListenersArray, c.apiResponse.ListenerCalled) {
				return nil
			}
		}
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// InTheResponseShouldHaveEachOneOfThese checks that the response object contains each of the provided properties.
func (c *ScenarioCtx) InTheResponseShouldHaveEachOneOfThese(argumentType string, value *gherkin.DocString) error {
	switch argumentType {
	case models.KeyListenerCalled:
		expectedListenersCalled := parseListeners(value.Content)
		found := false
		for _, expectedListener := range expectedListenersCalled {
			found = false
			for _, actualListener := range c.apiResponse.ListenerCalled {
				if subset.Check(expectedListener, actualListener) {
					found = true
					break
				}
			}
			if !found {
				break
			}
		}
		if found {
			return nil
		}
	default:
		break
	}
	return fmt.Errorf("response for %s not equal", argumentType)
}

// TheNumberOfDispatchedEventsIs checks the count of the dispatched events to be equal to the given value.
func (c *ScenarioCtx) TheNumberOfDispatchedEventsIs(count int) error {
	evaluationMethod := func() (bool, string) {
		dispatchedEvents := c.clientWrapper.eventDispatcher.(optlyplugins.EventReceiver).GetEvents()
		result := len(dispatchedEvents) == count
		if result {
			return result, ""
		}
		return result, "dispatchedEvents count not equal"
	}
	result, errorMessage := evaluateDispatchedEventsWithTimeout(evaluationMethod)
	if result {
		return nil
	}
	return fmt.Errorf(errorMessage)
}

// ThereAreNoDispatchedEvents checks the dispatched events count to be empty.
func (c *ScenarioCtx) ThereAreNoDispatchedEvents() error {
	evaluationMethod := func() (bool, string) {
		dispatchedEvents := c.clientWrapper.eventDispatcher.(optlyplugins.EventReceiver).GetEvents()
		result := len(dispatchedEvents) == 0
		if result {
			return result, ""
		}
		return result, fmt.Sprintf("dispatchedEvents should be empty but received %d events", len(dispatchedEvents))
	}
	result, errorMessage := evaluateDispatchedEventsWithTimeout(evaluationMethod)
	if result {
		return nil
	}
	return fmt.Errorf(errorMessage)
}

// DispatchedEventsPayloadsInclude checks dispatched events to contain the given events.
func (c *ScenarioCtx) DispatchedEventsPayloadsInclude(value *gherkin.DocString) error {

	config := c.clientWrapper.GetProjectConfig()

	expectedBatchEvents, err := parseYamlArray(value.Content, config)
	if err != nil {
		return fmt.Errorf("Invalid request for dispatched Events")
	}

	evaluationMethod := func() (bool, string) {
		eventsReceived := c.clientWrapper.eventDispatcher.(optlyplugins.EventReceiver).GetEvents()
		eventsReceivedJSON, err := json.Marshal(eventsReceived)
		if err != nil {
			return false, "Invalid response for dispatched Events"
		}
		var actualBatchEvents []map[string]interface{}

		if err := json.Unmarshal(eventsReceivedJSON, &actualBatchEvents); err != nil {
			return false, "Invalid response for dispatched Events"
		}

		// Sort's attributes under visitors which is required for subset comparison of attributes array
		sortAttributesForEvents := func(array []map[string]interface{}) []map[string]interface{} {
			sortedArray := array
			for mainIndex, event := range array {
				if visitorsArray, ok := event["visitors"].([]interface{}); ok {
					for vIndex, v := range visitorsArray {
						if visitor, ok := v.(map[string]interface{}); ok {
							// Only sort if all attributes were parsed successfuly
							parsedSuccessfully := false
							parsedAttributes := []map[string]interface{}{}
							if attributesArray, ok := visitor["attributes"].([]interface{}); ok {
								for _, tmpAttribute := range attributesArray {
									if attribute, ok := tmpAttribute.(map[string]interface{}); ok {
										parsedAttributes = append(parsedAttributes, attribute)
									}
								}
								parsedSuccessfully = len(attributesArray) == len(parsedAttributes)
							}
							if parsedSuccessfully {
								// Sort parsed attributes array and assign them to the original events array
								sortedAttributes := sortArrayofMaps(parsedAttributes, "key")
								sortedArray[mainIndex]["visitors"].([]interface{})[vIndex].(map[string]interface{})["attributes"] = sortedAttributes
							}
						}
					}
				}
			}
			return sortedArray
		}

		expectedBatchEvents = sortAttributesForEvents(expectedBatchEvents)
		actualBatchEvents = sortAttributesForEvents(actualBatchEvents)
		result := subset.Check(expectedBatchEvents, actualBatchEvents)
		if result {
			return result, ""
		}
		return result, "DispatchedEvents not equal"
	}

	result, errorMessage := evaluateDispatchedEventsWithTimeout(evaluationMethod)
	if result {
		return nil
	}
	return fmt.Errorf(errorMessage)
}

// PayloadsOfDispatchedEventsDontIncludeDecisions checks dispatched events to contain no decisions.
func (c *ScenarioCtx) PayloadsOfDispatchedEventsDontIncludeDecisions() error {
	evaluationMethod := func() (bool, string) {
		dispatchedEvents := c.clientWrapper.eventDispatcher.(optlyplugins.EventReceiver).GetEvents()
		for _, event := range dispatchedEvents {
			for _, visitor := range event.Visitors {
				for _, snapshot := range visitor.Snapshots {
					if len(snapshot.Decisions) > 0 {
						return false, "dispatched events should not include decisions"
					}
				}
			}
		}
		return true, ""
	}
	result, errorMessage := evaluateDispatchedEventsWithTimeout(evaluationMethod)
	if result {
		return nil
	}
	return fmt.Errorf(errorMessage)
}

// TheUserProfileServiceStateShouldBe checks current state of UPS
func (c *ScenarioCtx) TheUserProfileServiceStateShouldBe(value *gherkin.DocString) error {
	config := c.clientWrapper.GetProjectConfig()
	rawProfiles, err := parseYamlArray(value.Content, config)
	if err != nil {
		return fmt.Errorf("Invalid request for user profile service state")
	}

	expectedProfiles := userprofileservice.ParseUserProfiles(rawProfiles)
	actualProfiles := c.clientWrapper.userProfileService.(userprofileservice.UPSHelper).GetUserProfiles()

	success := false
	for _, expectedProfile := range expectedProfiles {
		success = false
		for _, actualProfile := range actualProfiles {
			if subset.Check(expectedProfile, actualProfile) {
				success = true
				break
			}
		}
		if !success {
			break
		}
	}
	if success {
		return nil
	}
	// @TODO: Temporary fix, need to look into it
	return getErrorWithDiff(expectedProfiles, actualProfiles, "User profile state not equal")
}

// ThereIsNoUserProfileState checks that UPS is empty
func (c *ScenarioCtx) ThereIsNoUserProfileState() error {
	actualProfiles := c.clientWrapper.userProfileService.(userprofileservice.UPSHelper).GetUserProfiles()
	if len(actualProfiles) == 0 {
		return nil
	}
	// @TODO: Temporary fix, need to look into it
	return getErrorWithDiff([]decision.UserProfile{}, actualProfiles, "User profile state not empty")
}

// Reset clears all data before each scenario, assigns new scenarioID and sets session as false
func (c *ScenarioCtx) Reset() {
	// Delete cached optly wrapper instance
	DeleteInstance()
	// Clear scenario context and generate a new scenarioID
	c.apiOptions = models.APIOptions{}
	c.apiResponse = models.APIResponse{}
	c.clientWrapper = nil
	c.scenarioID = uuid.New().String()
}
