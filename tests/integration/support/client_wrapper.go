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
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins/userprofileservice"
	"gopkg.in/yaml.v3"
)

// Map to hold clientwrapper instances against scenarioID
var clientInstance *ClientWrapper

// ClientWrapper - wrapper around the optimizely client that keeps track of various custom components used with the client
type ClientWrapper struct {
	Client             *client.OptimizelyClient
	DecisionService    decision.Service
	EventDispatcher    event.Dispatcher
	UserProfileService decision.UserProfileService
	OverrideStore      decision.ExperimentOverrideStore
}

// DeleteInstance deletes cached instance of optly wrapper
func DeleteInstance() {
	clientInstance = nil
}

// GetInstance returns a cached or new instance of the optly wrapper
func GetInstance(apiOptions models.APIOptions) *ClientWrapper {

	if clientInstance != nil {
		return clientInstance
	}

	datafileDir := os.Getenv("DATAFILES_DIR")
	datafile, err := ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, apiOptions.DatafileName)))
	if err != nil {
		log.Fatal(err)
	}
	configManager, err := config.NewStaticProjectConfigManagerFromPayload(datafile)
	if err != nil {
		log.Fatal(err)
	}

	eventProcessor := event.NewBatchEventProcessor(
		event.WithBatchSize(models.EventProcessorDefaultBatchSize),
		event.WithQueueSize(models.EventProcessorDefaultQueueSize),
		event.WithFlushInterval(models.EventProcessorDefaultFlushInterval),
	)

	optimizelyFactory := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	eventProcessor.EventDispatcher = &optlyplugins.ProxyEventDispatcher{}

	config, err := configManager.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	overrideService := decision.NewMapExperimentOverridesStore(decision.WithConfig(config))
	userProfileService := userprofileservice.CreateUserProfileService(config, apiOptions)
	compositeExperimentService := decision.NewCompositeExperimentService(
		decision.WithUserProfileService(userProfileService),
		decision.WithOverrideStore(overrideService),
	)

	// @TODO: Add sdkKey dynamically once event-batching support is implemented
	compositeService := *decision.NewCompositeService("", decision.WithCompositeExperimentService(compositeExperimentService))
	decisionService := &optlyplugins.TestCompositeService{CompositeService: compositeService}

	client, err := optimizelyFactory.Client(
		client.WithConfigManager(configManager),
		client.WithDecisionService(decisionService),
		client.WithEventProcessor(eventProcessor),
	)
	if err != nil {
		log.Fatal(err)
	}

	clientInstance = &ClientWrapper{
		Client:             client,
		DecisionService:    decisionService,
		EventDispatcher:    eventProcessor.EventDispatcher,
		UserProfileService: userProfileService,
		OverrideStore:      overrideService,
	}
	return clientInstance
}

// InvokeAPI processes request with arguments
func (c *ClientWrapper) InvokeAPI(request models.APIOptions) (models.APIResponse, error) {

	c.DecisionService.(*optlyplugins.TestCompositeService).AddListeners(request.Listeners)
	var response models.APIResponse
	var err error

	switch models.SDKAPI(request.APIName) {
	case models.IsFeatureEnabled:
		response, err = c.isFeatureEnabled(request)
		break
	case models.GetFeatureVariable:
		response, err = c.getFeatureVariable(request)
		break
	case models.GetFeatureVariableInteger:
		response, err = c.getFeatureVariableInteger(request)
		break
	case models.GetFeatureVariableDouble:
		response, err = c.getFeatureVariableDouble(request)
		break
	case models.GetFeatureVariableBoolean:
		response, err = c.getFeatureVariableBoolean(request)
		break
	case models.GetFeatureVariableString:
		response, err = c.getFeatureVariableString(request)
		break
	case models.GetEnabledFeatures:
		response, err = c.getEnabledFeatures(request)
		break
	case models.GetVariation:
		response, err = c.getVariation(request)
		break
	case models.Activate:
		response, err = c.activate(request)
		break
	case models.Track:
		response, err = c.track(request)
		break
	case models.SetForcedVariation:
		response, err = c.setForcedVariation(request)
		break
	case models.GetForcedVariation:
		response, err = c.getForcedVariation(request)
		break
	default:
		break
	}
	// TODO: For event batching, it should be conditional.
	c.Client.Close()
	response.ListenerCalled = c.DecisionService.(*optlyplugins.TestCompositeService).GetListenersCalled()
	return response, err
}

func (c *ClientWrapper) isFeatureEnabled(request models.APIOptions) (models.APIResponse, error) {
	var params models.IsFeatureEnabledRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}

		isEnabled, err := c.Client.IsFeatureEnabled(params.FeatureKey, user)
		result := "false"
		if err == nil && isEnabled {
			result = "true"
		}
		response.Result = result
	}
	return response, err
}

func (c *ClientWrapper) getFeatureVariable(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetFeatureVariableRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		value, valueType, err := c.Client.GetFeatureVariable(params.FeatureKey, params.VariableKey, user)
		if err == nil {
			response.Result = value
			response.Type = valueType
		}
	}
	return response, err
}

func (c *ClientWrapper) getFeatureVariableInteger(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetFeatureVariableRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		value, err := c.Client.GetFeatureVariableInteger(params.FeatureKey, params.VariableKey, user)
		if err == nil {
			response.Result = value
		}
	}
	return response, err
}

func (c *ClientWrapper) getFeatureVariableDouble(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetFeatureVariableRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		value, err := c.Client.GetFeatureVariableDouble(params.FeatureKey, params.VariableKey, user)
		if err == nil {
			response.Result = value
		}
	}
	return response, err
}

func (c *ClientWrapper) getFeatureVariableBoolean(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetFeatureVariableRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		value, err := c.Client.GetFeatureVariableBoolean(params.FeatureKey, params.VariableKey, user)
		if err == nil {
			response.Result = value
		}
	}
	return response, err
}

func (c *ClientWrapper) getFeatureVariableString(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetFeatureVariableRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		value, err := c.Client.GetFeatureVariableString(params.FeatureKey, params.VariableKey, user)
		if err == nil {
			response.Result = value
		}
	}
	return response, err
}

func (c *ClientWrapper) getEnabledFeatures(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetEnabledFeaturesRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		var enabledFeatures []string
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		if values, err := c.Client.GetEnabledFeatures(user); err == nil {
			enabledFeatures = values
		}
		response.Result = enabledFeatures
	}
	return response, err
}

func (c *ClientWrapper) getVariation(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetVariationRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		response.Result, _ = c.Client.GetVariation(params.ExperimentKey, user)
		if response.Result == "" {
			response.Result = "NULL"
		}
	}
	return response, err
}

func (c *ClientWrapper) activate(request models.APIOptions) (models.APIResponse, error) {
	var params models.GetVariationRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		response.Result, _ = c.Client.Activate(params.ExperimentKey, user)
		if response.Result == "" {
			response.Result = "NULL"
		}
	}
	return response, err
}

func (c *ClientWrapper) track(request models.APIOptions) (models.APIResponse, error) {
	var params models.TrackRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		err = c.Client.Track(params.EventKey, user, params.EventTags)
	}
	response.Result = "NULL"
	return response, err
}

func (c *ClientWrapper) setForcedVariation(request models.APIOptions) (models.APIResponse, error) {
	var params models.ForcedVariationRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	response.Result = "NULL"
	if err == nil {
		// For removeForcedVariation cases
		if params.VariationKey == "" {
			c.OverrideStore.(*decision.MapExperimentOverridesStore).RemoveVariation(decision.ExperimentOverrideKey{ExperimentKey: params.ExperimentKey, UserID: params.UserID})
		} else {
			result := c.OverrideStore.(*decision.MapExperimentOverridesStore).SetVariation(decision.ExperimentOverrideKey{ExperimentKey: params.ExperimentKey, UserID: params.UserID}, params.VariationKey)
			if result {
				response.Result = params.VariationKey
			}
		}
	}
	return response, err
}

func (c *ClientWrapper) getForcedVariation(request models.APIOptions) (models.APIResponse, error) {
	var params models.ForcedVariationRequestParams
	var response models.APIResponse
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	response.Result = "NULL"
	if err == nil {
		variation, success := c.OverrideStore.(*decision.MapExperimentOverridesStore).GetVariation(decision.ExperimentOverrideKey{ExperimentKey: params.ExperimentKey, UserID: params.UserID})
		if success {
			response.Result = variation
		}
	}
	return response, err
}
