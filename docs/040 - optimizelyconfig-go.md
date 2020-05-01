---
title: "OptimizelyConfig"
slug: "optimizelyconfig-go"
hidden: true
createdAt: "2020-01-17T19:10:55.970Z"
updatedAt: "2020-01-28T21:53:26.607Z"
---
[block:api-header]
{
  "title": "Overview"
}
[/block]
Full Stack SDKs open a well-defined set of public APIs, hiding all implementation details. However, some clients may need access to project configuration data within the "datafile". 

In this document, we extend our public APIs to define data models and access methods, which clients can use to access project configuration data. 

[block:api-header]
{
  "title": "OptimizelyConfig API"
}
[/block]

A public configuration data model (OptimizelyConfig) is defined below as a structured format of static Optimizely Project data.

OptimizelyConfig can be accessed from OptimizelyClient (top-level) with this public API call:
[block:code]
{
  "codes": [
    {
      "code": "client, e := optimizelyFactory.Client()\nvar config = client.GetOptimizelyConfig()",
      "language": "go"
    }
  ]
}
[/block]
`GetOptimizelyConfig` returns an `config.OptimizelyConfig` instance which include a datafile revision number, all experiments, and feature flags mapped by their key values.
[block:callout]
{
  "type": "info",
  "title": "Note",
  "body": "When the SDK datafile is updated (the client can add a notification listener for `ProjectConfigUpdateNotification` to get notified), the client is expected to call the method to get the updated OptimizelyConfig data. See examples below."
}
[/block]

[block:code]
{
  "codes": [
    {
      "code": "// OptimizelyConfig is an object describing the current project configuration data \ntype OptimizelyConfig struct {\n\tRevision       string                          \n\tExperimentsMap map[string]OptimizelyExperiment \n\tFeaturesMap    map[string]OptimizelyFeature    \n}\n\n\n// OptimizelyFeature is an object describing a feature\ntype OptimizelyFeature struct {\n\tID             string                          \n\tKey            string                          \n\tExperimentsMap map[string]OptimizelyExperiment \n\tVariablesMap   map[string]OptimizelyVariable  \n}\n\n\n// OptimizelyExperiment is an object describing a feature test or an A/B test\ntype OptimizelyExperiment struct {\n\tID            string                         \n\tKey           string                        \n\tVariationsMap map[string]OptimizelyVariation \n}\n\n\n\n// OptimizelyVariation is an object describing a variation in a feature test or A/B //test\ntype OptimizelyVariation struct {\n\tID             string                       \n\tKey            string                      \n\tFeatureEnabled bool                          \n\tVariablesMap   map[string]OptimizelyVariable \n}\n\n\n// OptimizelyVariable is an object describing a feature variable\ntype OptimizelyVariable struct {\n\tID    string \n\tKey   string \n\tType  string \n\tValue string \n}",
      "language": "go"
    }
  ]
}
[/block]

[block:api-header]
{
  "title": "Examples"
}
[/block]
OptimizelyConfig can be accessed from OptimizelyClient (top-level) like this:

[block:code]
{
  "codes": [
    {
      "code": "client, e := optimizelyFactory.Client()\n// all experiments\nvar experimentsMap = optimizelyConfig.ExperimentsMap\nvar experiments = []config.OptimizelyExperiment{}\nvar experimentKeys = []string{}\n\nfor experimentKey, experiment := range experimentsMap {\n\texperimentKeys = append(experimentKeys, experimentKey)\n\texperiments = append(experiments, experiment)\n}\n\nfor _, experimentKey := range experimentKeys {\n\tvar experiment = experimentsMap[experimentKey]\n\n\t// all variations for an experiment\n\tvar variationsMap = experiment.VariationsMap\n\tvar variations = []config.OptimizelyVariation{}\n\tvar variationKeys = []string{}\n\n\tfor variationKey, variation := range variationsMap {\n\t\tvariations = append(variations, variation)\n\t\tvariationKeys = append(variationKeys, variationKey)\n\t}\n\n\tfor _, variationKey := range variationKeys {\n\t\tvar variation = variationsMap[variationKey]\n\n\t\t// all variables for a variation\n\t\tvar variablesMap = variation.VariablesMap\n\t\tvar variables = []config.OptimizelyVariable{}\n\t\tvar variableKeys = []string{}\n\n\t\tfor variableKey, variable := range variablesMap {\n\t\t\tvariables = append(variables, variable)\n\t\t\tvariableKeys = append(variableKeys, variableKey)\n\t\t}\n\n\t\tfor _, variableKey := range variableKeys {\n\t\t\tvar variable = variablesMap[variableKey]\n\t\t\t// use variable data here...\n\t\t}\n\t}\n}\n\n// all features\nvar featuresMap = optimizelyConfig.FeaturesMap\nvar features = []config.OptimizelyFeature{}\nvar featureKeys = []string{}\n\nfor featureKey, feature := range featuresMap {\n\tfeatures = append(features, feature)\n\tfeatureKeys = append(featureKeys, featureKey)\n}\n\nfor _, featureKey := range featureKeys {\n\tvar feature = featuresMap[featureKey]\n\n\t// all experiments for a feature\n\tvar experimentsMap = feature.ExperimentsMap\n\tvar experiments = []config.OptimizelyExperiment{}\n\tvar experimentKeys = []string{}\n\n\tfor experimentKey, experiment := range experimentsMap {\n\t\texperiments = append(experiments, experiment)\n\t\texperimentKeys = append(experimentKeys, experimentKey)\n\t}\n\n\t// use experiments and other feature data here...\n}\n\n// listen to ProjectConfigUpdateNotification to get updated data\ncallback := func(notification notification.ProjectConfigUpdateNotification) {\n\tvar newConfig = client.GetOptimizelyConfig()\n\t// ...\n}\nclient.ConfigManager.OnProjectConfigUpdate(callback)",
      "language": "go"
    }
  ]
}
[/block]