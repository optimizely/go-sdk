package support

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/optimizely/go-sdk/checking/integration/optimizely/datamodels"
	"github.com/optimizely/go-sdk/checking/integration/optimizely/listener"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/entities"
	"github.com/optimizely/go-sdk/optimizely/notification"
	"gopkg.in/yaml.v2"
)

func setupOptimizelyClient(requestParams datamodels.RequestParams) (*client.OptimizelyClient, decision.Service, error) {
	datafileDir := os.Getenv("DATAFILES_DIR")
	datafile, err := ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, requestParams.DatafileName)))
	if err != nil {
		return nil, nil, err
	}

	optimizelyFactory := &client.OptimizelyFactory{
		Datafile: datafile,
	}
	// Creates a default, canceleable context
	notificationCenter := notification.NewNotificationCenter()
	decisionService := decision.NewCompositeService(notificationCenter)

	// clientOptions := client.Options{
	// 	DecisionService: decisionService,
	// }

	client, err := optimizelyFactory.Client(
		client.CompositeDecisionService(notificationCenter),
	)
	return client, decisionService, nil
}

// ProcessRequest processes request with arguments
func ProcessRequest(request datamodels.RequestParams) (*datamodels.ResponseParams, error) {

	client, decisionService, err := setupOptimizelyClient(request)
	if err != nil {
		return nil, err
	}
	listenersCalled := listener.AddListener(decisionService, request)

	responseParams := datamodels.ResponseParams{}

	switch request.ApiName {
	case "is_feature_enabled":
		var params datamodels.IsFeatureEnabledRequestParams
		err := yaml.Unmarshal([]byte(request.Arguments), &params)
		if err == nil {
			user := entities.UserContext{
				ID:         params.UserID,
				Attributes: params.Attributes,
			}

			isEnabled, err := client.IsFeatureEnabled(params.FeatureKey, user)
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
			value, valueType, err := client.GetFeatureVariable(params.FeatureKey, params.VariableKey, user)
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
			value, err := client.GetFeatureVariableInteger(params.FeatureKey, params.VariableKey, user)
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
			value, err := client.GetFeatureVariableDouble(params.FeatureKey, params.VariableKey, user)
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
			value, err := client.GetFeatureVariableBoolean(params.FeatureKey, params.VariableKey, user)
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
			value, err := client.GetFeatureVariableString(params.FeatureKey, params.VariableKey, user)
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
