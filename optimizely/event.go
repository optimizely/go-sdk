package optimizely

import (
	"fmt"
	"net/url"
	"time"
)

var OFFLINE_API_PATH = "https://%v.log.optimizely.com/event" // project_id
var END_USER_ID_TEMPLATE = "oeu-%v"                          // user_id

const (
	ACCOUNT_ID  = "d"
	PROJECT_ID  = "a"
	EXPERIMENT  = "x"
	GOAL_ID     = "g"
	GOAL_NAME   = "n"
	SEGMENT     = "s"
	END_USER_ID = "u"
	REVENUE     = "v"
	SOURCE      = "src"
	TIME        = "time"
	SDK_VERSION = "0.0"
)

func DispatchEvent() {

}

type Event struct {
	params url.Values
}

// GetURLForImpressionEvent returns an url for sending impression / conversion
// events.
// project_id: ID for the project
func GetUrlForImpressionEvent(project_id string) string {
	return fmt.Sprintf(OFFLINE_API_PATH, project_id)
}

// Add experiment to variation mapping to the impression event
// experiment_key: Experiment which is being activated
// variation_id: id for variation which would be presented to the user
func (event *Event) add_experiment(project_config ProjectConfig, experiment_key string, variation_id string) {
	experiment_id := ""
	for i := 0; i < len(project_config.Experiments); i++ {
		if project_config.Experiments[i].Key == experiment_key {
			experiment_id = project_config.Experiments[i].Id
		}
	}
	event.params.Add(fmt.Sprintf("{%v}{%v}", EXPERIMENT, experiment_id), variation_id)
}

// Add imp[ression goal information to the event
// experiment_key: Experiment which is being activated
func (event *Event) add_impression_goal(project_config ProjectConfig, experiment_key string) {
	experiment_id := ""
	for i := 0; i < len(project_config.Experiments); i++ {
		if project_config.Experiments[i].Key == experiment_key {
			experiment_id = project_config.Experiments[i].Id
		}
	}
	event.params.Add(GOAL_ID, experiment_id)
	event.params.Add(GOAL_NAME, "visitor-event")
}

// Get segment Id for the provided attribute key
// project_config: the project_Config
// attributes: the attributes to search through
// attribute_key: the attribute key for which segment ID is to be determined
func GetSementId(project_config ProjectConfig, attributes []AttributeEntity, attribute_key string) string {
	for i := 0; i < len(attributes); i++ {
		if attributes[i].Key == attribute_key {
			return attributes[i].SegmentId
		}
	}
	return ""
}

// BuildAttributeParams adds attribute parameters to the URL Value Map
func build_attribute_params(event *Event, project_config ProjectConfig, attributes []AttributeEntity) {
	for i := 0; i < len(attributes); i++ {
		segment_id := GetSementId(project_config, attributes, attributes[i].Key)
		if len(segment_id) > 0 {
			event.params.Add(fmt.Sprintf("{%v}{%v}", SEGMENT, segment_id), attributes[i].Value)
		}
	}
}

// Add params which are used in both conversion and impression events
// user_id: Id for user
// attributes: array representing user attributes and values which need to be recorded
func (event *Event) add_common_params(client *OptimizelyClient, user_id string, attributes []AttributeEntity) {
	event.params.Add(PROJECT_ID, client.project_config.ProjectId)
	event.params.Add(ACCOUNT_ID, client.account_id)
	event.params.Add(END_USER_ID, user_id)
	build_attribute_params(event, client.project_config, attributes)
	event.params.Add(SOURCE, fmt.Sprintf("go-sdk-%v", SDK_VERSION))
	event.params.Add(TIME, fmt.Sprint("%v", time.Now().Unix()))
}

// Create impression Event to be sent to the logging endpoint.
// experiment_key: Experiment for which impression needs to be recorded.
// variation_id: ID for variation which would be presented to user.
// user_id: ID for user.
// attributes: Dict representing user attributes and values which need to be recorded.
// Returns: Event object encapsulating the impression event.
func CreateImpressionEvent(client *OptimizelyClient, experiment_key string, variation_id string, user_id string, attributes []AttributeEntity) *Event {
	event := &Event{}
	event.add_common_params(client, user_id, attributes)
	event.add_impression_goal(client.project_config, experiment_key)
	event.add_experiment(client.project_config, experiment_key, variation_id)
	return event
}
