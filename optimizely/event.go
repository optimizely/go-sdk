package optimizely

var OFFLINE_API_PATH = "https://%v.log.optimizely.com/event" // project_id
var END_USER_ID_TEMPLATE = "oeu-%v"                          // user_id
var REQUEST_TIMEOUT = 10

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
)

func DispatchEvent() {

}
