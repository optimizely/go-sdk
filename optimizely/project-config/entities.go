package project_config

type Attribute struct {
	Id  string `json:"id"`
	Key string `json:"key"`
}

type Audience struct {
	Id         string      `json:"id"`
	Name       string      `json:"name"`
	Conditions interface{} `json:"conditions"`
}

type Variation struct {
	Id             string   `json:"id"`
	Variables      []string `json:"variables"`
	Key            string   `json:"key"`
	FeatureEnabled bool     `json:"featureEnabled"`
}

type Event struct {
	Id            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}

type TrafficAllocation struct {
	EntityId   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

type Experiment struct {
	Id                string              `json:"id"`
	LayerId           string              `json:"layerId"`
	Key               string              `json:"key"`
	Status            string              `json:"status"`
	Variations        []Variation         `json:"variations"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
	AudienceIds       []string            `json:"audienceIds"`
	ForcedVariations  map[string]string   `json:"forcedVariations"`
	GroupId           string              `json:"-"`
	GroupPolicy       string              `json:"-"`
}

type Flag struct {
	ID            string   `json:"id"`
	RolloutID     string   `json:"rolloutId"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
	Variables     []string `json:"variables"`
}

type Group struct {
	Id                string              `json:"id"`
	Policy            string              `json:"policy"`
	Experiments       []Experiment        `json:"experiments"`
	TrafficAllocation []TrafficAllocation `json:"trafficAllocation"`
}

type Rollout struct {
	ID          string       `json:"id"`
	Experiments []Experiment `json:"experiments"`
}

//ProjectConfig is the main thing for project configuration.
type ProjectConfig struct {
	Version        string       `json:"version"`
	Rollouts       []Rollout    `json:"rollouts"`
	TypedAudiences []Audience   `json:"typedAudiences"`
	AnonymizeIP    bool         `json:"anonymizeIP"`
	ProjectId      string       `json:"projectId"`
	Variables      []string     `json:"variables"`
	FeatureFlags   []Flag       `json:"featureFlags"`
	Experiments    []Experiment `json:"experiments"`
	Audiences      []Audience   `json:"audiences"`
	Groups         []Group      `json:"groups"`
	Attributes     []Attribute  `json:"attributes"`
	BotFiltering   bool         `json:"botFiltering"`
	AccountID      string       `json:"accountId"`
	Events         []Event      `json:"events"`
	Revision       string       `json:"revision"`
}
