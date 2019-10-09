package support

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/optimizely/go-sdk/pkg/client"
	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decision"
	"github.com/optimizely/go-sdk/pkg/entities"
	"github.com/optimizely/go-sdk/pkg/event"
	"github.com/optimizely/go-sdk/tests/integration/models"
	"github.com/optimizely/go-sdk/tests/integration/optlyplugins"
	"gopkg.in/yaml.v3"
)

// ClientWrapper - wrapper around the optimizely client that keeps track of various custom components used with the client
type ClientWrapper struct {
	Client          *client.OptimizelyClient
	DecisionService decision.Service
	EventDispatcher event.Dispatcher
}

// NewClientWrapper returns a new instance of the optly wrapper
func NewClientWrapper(datafileName string) ClientWrapper {

	datafileDir := os.Getenv("DATAFILES_DIR")
	datafile, err := ioutil.ReadFile(filepath.Clean(path.Join(datafileDir, datafileName)))
	if err != nil {
		log.Fatal(err)
	}
	configManager, err := config.NewStaticProjectConfigManagerFromPayload(datafile)
	if err != nil {
		log.Fatal(err)
	}

	eventProcessor := event.NewEventProcessor(event.BatchSize(models.EventProcessorDefaultBatchSize), event.QueueSize(models.EventProcessorDefaultQueueSize), event.FlushInterval(models.EventProcessorDefaultFlushInterval))

	optimizelyFactory := &client.OptimizelyFactory{
		Datafile: datafile,
	}

	decisionService := &optlyplugins.TestCompositeService{CompositeService: *decision.NewCompositeService("")}
	eventProcessor.EventDispatcher = &optlyplugins.ProxyEventDispatcher{}

	client, err := optimizelyFactory.Client(
		client.WithConfigManager(configManager),
		client.WithDecisionService(decisionService),
		client.WithEventProcessor(eventProcessor))
	if err != nil {
		log.Fatal(err)
	}

	return ClientWrapper{
		Client:          client,
		DecisionService: decisionService,
		EventDispatcher: eventProcessor.EventDispatcher,
	}
}

// InvokeAPI processes request with arguments
func (c *ClientWrapper) InvokeAPI(request models.APIOptions, response *models.APIResponse) error {

	c.DecisionService.(*optlyplugins.TestCompositeService).AddListeners(request.Listeners)
	var err error

	switch request.APIName {
	case "is_feature_enabled":
		err = c.isFeatureEnabled(request, response)
		break
	case "get_feature_variable":
		err = c.getFeatureVariable(request, response)
		break
	case "get_feature_variable_integer":
		err = c.getFeatureVariableInteger(request, response)
		break
	case "get_feature_variable_double":
		err = c.getFeatureVariableDouble(request, response)
		break
	case "get_feature_variable_boolean":
		err = c.getFeatureVariableBoolean(request, response)
		break
	case "get_feature_variable_string":
		err = c.getFeatureVariableString(request, response)
		break
	case "get_enabled_features":
		err = c.getEnabledFeatures(request, response)
		break
	default:
		break
	}

	response.ListenerCalled = c.DecisionService.(*optlyplugins.TestCompositeService).GetListenersCalled()
	return err
}

func (c *ClientWrapper) isFeatureEnabled(request models.APIOptions, response *models.APIResponse) error {
	var params models.IsFeatureEnabledRequestParams
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
	return err
}

func (c *ClientWrapper) getFeatureVariable(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetFeatureVariableParams
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
	return err
}

func (c *ClientWrapper) getFeatureVariableInteger(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetFeatureVariableParams
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
	return err
}

func (c *ClientWrapper) getFeatureVariableDouble(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetFeatureVariableParams
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
	return err
}

func (c *ClientWrapper) getFeatureVariableBoolean(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetFeatureVariableParams
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
	return err
}

func (c *ClientWrapper) getFeatureVariableString(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetFeatureVariableParams
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
	return err
}

func (c *ClientWrapper) getEnabledFeatures(request models.APIOptions, response *models.APIResponse) error {
	var params models.GetEnabledFeaturesParams
	err := yaml.Unmarshal([]byte(request.Arguments), &params)
	if err == nil {
		enabledFeatures := ""
		user := entities.UserContext{
			ID:         params.UserID,
			Attributes: params.Attributes,
		}
		if values, err := c.Client.GetEnabledFeatures(user); err == nil {
			enabledFeatures = strings.Join(values, ",")
		}
		response.Result = enabledFeatures
	}
	return err
}
