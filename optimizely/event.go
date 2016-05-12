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

// Add experiment to variation mapping to the impression event
// experiment_key: Experiment which is being activated
// variation_id: id for variation which would be presented to the user
func (event *Event) add_experiment(experiment_key string, variation_id string) {

}

// Add imp[ression goal information to the event
// experiment_key: Experiment which is being activated
func (event *Event) add_impression_goal(experiment_key string) {

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
func CreateImpressionEvent(client *OptimizelyClient, experiment_key string, variation_id string, user_id string, attributes []AttributeEntity) {
	event := &Event{}
	event.add_common_params(client, user_id, attributes)
}
