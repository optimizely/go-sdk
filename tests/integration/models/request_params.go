package models

import (
	"github.com/optimizely/go-sdk/optimizely"
	"github.com/optimizely/go-sdk/optimizely/client"
	"github.com/optimizely/go-sdk/optimizely/decision"
	"github.com/optimizely/go-sdk/optimizely/event"
)

// RequestParams represents parameters for a scenario
type RequestParams struct {
	APIName         string
	Arguments       string
	DatafileName    string
	Listeners       map[string]int
	DependencyModel *SDKDependencyModel
}

// SDKDependencyModel represents sdk dependencies required for request processing
type SDKDependencyModel struct {
	Client          *client.OptimizelyClient
	DecisionService decision.Service
	Config          optimizely.ProjectConfig
	Dispatcher      event.Dispatcher
}
