package support

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins/listener"
	"gopkg.in/yaml.v3"
)

func setupOptimizelyClient(requestParams *models.RequestParams) {
	if requestParams.DependencyModel != nil {
		return
	}
	datafileDir := os.Getenv("DATAFILES_DIR")
	datafile, err := ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, requestParams.DatafileName)))
	if err != nil {
		return
	}
	configManager, err := config.NewStaticProjectConfigManagerFromPayload(datafile)
	if err != nil {
		return
	}
	projectConfig, err := configManager.GetConfig()
	if err != nil {
		return
	}

	executionCtx := utils.NewCancelableExecutionCtx()
	eventProcessor := event.NewEventProcessor(executionCtx, event.BatchSize(models.EventProcessorDefaultBatchSize), event.QueueSize(models.EventProcessorDefaultQueueSize), event.FlushInterval(models.EventProcessorDefaultFlushInterval))

	optimizelyFactory := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	decisionService := &listener.TestCompositeService{CompositeService: *decision.NewCompositeService("")}
	eventProcessor.EventDispatcher = &optlyplugins.ProxyEventDispatcher{}

	client, err := optimizelyFactory.Client(
		client.ConfigManager(configManager),
		client.DecisionService(decisionService),
		client.EventProcessor(eventProcessor))

	sdkDependencyModel := models.SDKDependencyModel{}
	sdkDependencyModel.Client = client
	sdkDependencyModel.DecisionService = decisionService
	sdkDependencyModel.Dispatcher = eventProcessor.EventDispatcher
	sdkDependencyModel.Config = projectConfig

	requestParams.DependencyModel = &sdkDependencyModel
}

// ProcessRequest processes request with arguments
func ProcessRequest(request *models.RequestParams) (*models.ResponseParams, error) {

	setupOptimizelyClient(request)
	if request.DependencyModel == nil {
		return nil, fmt.Errorf("Request failure")
	}

	listenersCalled := request.DependencyModel.DecisionService.(*listener.TestCompositeService).AddListener(request)
	responseParams := models.ResponseParams{}

	switch request.APIName {
	case "is_feature_enabled":
		var params models.IsFeatureEnabledRequestParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}

			isEnabled, err := request.DependencyModel.Client.IsFeatureEnabled(params.FeatureKey, user)
			result := "false"
			if err == nil && isEnabled {
				result = "true"
			}

			responseParams.Result = result
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	case "get_feature_variable":
		var params models.GetFeatureVariableParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}
			value, valueType, err := request.DependencyModel.Client.GetFeatureVariable(params.FeatureKey, params.VariableKey, user)
			if err == nil {
				responseParams.Result = value
				responseParams.Type = valueType
			}
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	case "get_feature_variable_integer":
		var params models.GetFeatureVariableParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}
			value, err := request.DependencyModel.Client.GetFeatureVariableInteger(params.FeatureKey, params.VariableKey, user)
			if err == nil {
				responseParams.Result = value
			}
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	case "get_feature_variable_double":
		var params models.GetFeatureVariableParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}
			value, err := request.DependencyModel.Client.GetFeatureVariableDouble(params.FeatureKey, params.VariableKey, user)
			if err == nil {
				responseParams.Result = value
			}
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	case "get_feature_variable_boolean":
		var params models.GetFeatureVariableParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}
			value, err := request.DependencyModel.Client.GetFeatureVariableBoolean(params.FeatureKey, params.VariableKey, user)
			if err == nil {
				responseParams.Result = value
			}
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	case "get_feature_variable_string":
		var params models.GetFeatureVariableParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}
			value, err := request.DependencyModel.Client.GetFeatureVariableString(params.FeatureKey, params.VariableKey, user)
			if err == nil {
				responseParams.Result = value
			}
			responseParams.ListenerCalled = listenersCalled()
			return &responseParams, err
		}
		break
	default:
		break
	}
	return nil, fmt.Errorf("invalid request params")
}
