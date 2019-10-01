package support

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/optimizely/go-sdk/checking/integration/optimizely/datamodels"
	"github.com/optimizely/go-sdk/checking/integration/optimizely/eventdispatcher"
	"github.com/optimizely/go-sdk/checking/integration/optimizely/listener"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/config"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/event"
	"github.com/optimizely/go-sdk/optimizely/utils"
	"gopkg.in/yaml.v3"
)

func setupOptimizelyClient(requestParams *datamodels.RequestParams) {
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
	eventProcessor := event.NewEventProcessor(executionCtx, event.BatchSize(datamodels.EventProcessorDefaultBatchSize), event.QueueSize(datamodels.EventProcessorDefaultQueueSize), event.FlushInterval(datamodels.EventProcessorDefaultFlushInterval))

	optimizelyFactory := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	decisionService := decision.NewCompositeService("")
	eventProcessor.EventDispatcher = &eventdispatcher.ProxyEventDispatcher{}

	client, err := optimizelyFactory.Client(
		client.ConfigManager(configManager),
		client.DecisionService(decisionService),
		client.EventProcessor(eventProcessor))

	sdkDependencyModel := datamodels.SDKDependencyModel{}
	sdkDependencyModel.Client = client
	sdkDependencyModel.DecisionService = decisionService
	sdkDependencyModel.Dispatcher = eventProcessor.EventDispatcher
	sdkDependencyModel.Config = projectConfig

	requestParams.DependencyModel = &sdkDependencyModel
}

// ProcessRequest processes request with arguments
func ProcessRequest(request *datamodels.RequestParams) (*datamodels.ResponseParams, error) {

	setupOptimizelyClient(request)
	if request.DependencyModel == nil {
		return nil, fmt.Errorf("Request failure")
	}
	listenersCalled := listener.AddListener(request.DependencyModel.DecisionService, request)

	responseParams := datamodels.ResponseParams{}

	switch request.APIName {
	case "is_feature_enabled":
		var params datamodels.IsFeatureEnabledRequestParams
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
		var params datamodels.GetFeatureVariableRequestParams
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
		var params datamodels.GetFeatureVariableRequestParams
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
		var params datamodels.GetFeatureVariableRequestParams
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
		var params datamodels.GetFeatureVariableRequestParams
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
		var params datamodels.GetFeatureVariableRequestParams
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
