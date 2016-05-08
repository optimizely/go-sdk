package optimizely

type AttributeEntity struct {
	Id        string `json:"id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	SegmentId string `json:"segmentId"`
}

type AudienceEntity struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Conditions string `json:"conditions"`
}

type VariationEntity struct {
	Id         string `json:"id"`
	Key        string `json:"key"`
	EndOfRange int    `json:"endOfRange"`
}

type TrafficAllocationEntity struct {
	EntityId   string `json:"entityId"`
	EndOfRange int    `json:"endOfRange"`
}

type ExperimentEntity struct {
	Id                 string                    `json:"id"`
	Key                string                    `json:"key"`
	Status             string                    `json:"status"`
	Variations         []VariationEntity         `json:"variations"`
	PercentageIncluded int                       `json:"percentageIncluded"`
	TrafficAllocation  []TrafficAllocationEntity `json:"trafficAllocation"`
	AudienceIds        []string                  `json:"audienceIds"`
}

type EventEntity struct {
	Id            string   `json:"id"`
	Key           string   `json:"key"`
	ExperimentIds []string `json:"experimentIds"`
}

type DimensionEntity struct {
	Id        string `json:"id"`
	Key       string `json:"key"`
	SegmentId string `json:"segmentId"`
}

type ProjectConfig struct {
	AccountId   string             `json:"accountId"`
	ProjectId   string             `json:"projectId"`
	Revision    string             `json:"revision"`
	Experiments []ExperimentEntity `json:"experiments"`
	Events      []EventEntity      `json:"events"`
	Dimensions  []DimensionEntity  `json:"dimensions"`
	Attributes  []AttributeEntity  `json:"attributes"`
	Audiences   []AudienceEntity   `json:"audiences"`
}

// OptimizelyClient is the client to interface with the Optimizely server
// side APIs.
type OptimizelyClient struct {
	account_id     string
	project_config ProjectConfig
}
