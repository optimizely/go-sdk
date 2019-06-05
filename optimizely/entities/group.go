package entities

// Group represents a grouping of entities and their traffic allocation ranges
type Group struct {
	ID                string
	TrafficAllocation []Range
}
