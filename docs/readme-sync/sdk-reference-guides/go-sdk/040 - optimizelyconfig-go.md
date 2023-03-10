---
title: "OptimizelyConfig"
excerpt: ""
slug: "optimizelyconfig-go"
hidden: true
createdAt: "2020-01-17T19:10:55.970Z"
updatedAt: "2020-01-28T21:53:26.607Z"
---
### Overview

Optimizely Feature Experimentation and Optimizely Full Stack (legacy) SDKs open a well-defined set of public APIs, hiding all implementation details. However, some clients may need access to project configuration data within the "datafile". 

In this document, we extend our public APIs to define data models and access methods, which clients can use to access project configuration data. 

### OptimizelyConfig API

A public configuration data model (OptimizelyConfig) is defined below as a structured format of static Optimizely Project data.

OptimizelyConfig can be accessed from OptimizelyClient (top-level) with this public API call:
```go
client, e := optimizelyFactory.Client()
var config = client.GetOptimizelyConfig()
```
`GetOptimizelyConfig` returns an `config.OptimizelyConfig` instance which include a datafile revision number, all experiments, and feature flags mapped by their key values.

>ℹ️ Note
>
> When the SDK datafile is updated (the client can add a notification listener for `ProjectConfigUpdateNotification` to get notified), the client is expected to call the method to get the updated OptimizelyConfig data. See examples below.

```go
// OptimizelyConfig is an object describing the current project configuration data 
type OptimizelyConfig struct {
	Revision       string                          
	ExperimentsMap map[string]OptimizelyExperiment 
	FeaturesMap    map[string]OptimizelyFeature    
}


// OptimizelyFeature is an object describing a feature
type OptimizelyFeature struct {
	ID             string                          
	Key            string                          
	ExperimentsMap map[string]OptimizelyExperiment 
	VariablesMap   map[string]OptimizelyVariable  
}


// OptimizelyExperiment is an object describing a feature test or an A/B test
type OptimizelyExperiment struct {
	ID            string                         
	Key           string                        
	VariationsMap map[string]OptimizelyVariation 
}


// OptimizelyVariation is an object describing a variation in a feature test or A/B //test
type OptimizelyVariation struct {
	ID             string                       
	Key            string                      
	FeatureEnabled bool                          
	VariablesMap   map[string]OptimizelyVariable 
}


// OptimizelyVariable is an object describing a feature variable
type OptimizelyVariable struct {
	ID    string 
	Key   string 
	Type  string 
	Value string 
}
```

### Examples
OptimizelyConfig can be accessed from OptimizelyClient (top-level) like this:

```go
client, e := optimizelyFactory.Client()
// all experiments
var experimentsMap = optimizelyConfig.ExperimentsMap
var experiments = []config.OptimizelyExperiment{}
var experimentKeys = []string{}

for experimentKey, experiment := range experimentsMap {
	experimentKeys = append(experimentKeys, experimentKey)
	experiments = append(experiments, experiment)
}

for _, experimentKey := range experimentKeys {
	var experiment = experimentsMap[experimentKey]

	// all variations for an experiment
	var variationsMap = experiment.VariationsMap
	var variations = []config.OptimizelyVariation{}
	var variationKeys = []string{}

	for variationKey, variation := range variationsMap {
		variations = append(variations, variation)
		variationKeys = append(variationKeys, variationKey)
	}

	for _, variationKey := range variationKeys {
		var variation = variationsMap[variationKey]

		// all variables for a variation
		var variablesMap = variation.VariablesMap
		var variables = []config.OptimizelyVariable{}
		var variableKeys = []string{}

		for variableKey, variable := range variablesMap {
			variables = append(variables, variable)
			variableKeys = append(variableKeys, variableKey)
		}

		for _, variableKey := range variableKeys {
			var variable = variablesMap[variableKey]
			// use variable data here...
		}
	}
}

// all features
var featuresMap = optimizelyConfig.FeaturesMap
var features = []config.OptimizelyFeature{}
var featureKeys = []string{}

for featureKey, feature := range featuresMap {
	features = append(features, feature)
	featureKeys = append(featureKeys, featureKey)
}

for _, featureKey := range featureKeys {
	var feature = featuresMap[featureKey]

	// all experiments for a feature
	var experimentsMap = feature.ExperimentsMap
	var experiments = []config.OptimizelyExperiment{}
	var experimentKeys = []string{}

	for experimentKey, experiment := range experimentsMap {
		experiments = append(experiments, experiment)
		experimentKeys = append(experimentKeys, experimentKey)
	}

	// use experiments and other feature data here...
}

// listen to ProjectConfigUpdateNotification to get updated data
callback := func(notification notification.ProjectConfigUpdateNotification) {
	var newConfig = client.GetOptimizelyConfig()
	// ...
}
client.ConfigManager.OnProjectConfigUpdate(callback)
```