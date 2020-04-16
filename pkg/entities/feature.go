/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package entities //
package entities

// Feature represents a feature flag
type Feature struct {
	ID                 string
	Key                string
	FeatureExperiments []Experiment
	Rollout            Rollout
	VariableMap        map[string]Variable
}

// Rollout represents a feature rollout
type Rollout struct {
	ID          string
	Experiments []Experiment
}

// Variable represents a feature variable
type Variable struct {
	DefaultValue string
	ID           string
	Key          string
	Type         VariableType
}

// VariableType is the type of feature variable
type VariableType string

const (
	// String - the feature-variable type is string
	String VariableType = "string"
	// Integer - the feature-variable type is integer
	Integer VariableType = "integer"
	// Double - the feature-variable type is double
	Double VariableType = "double"
	// Boolean - the feature-variable type is boolean
	Boolean VariableType = "boolean"
	// JSON - the feature-variable type is json
	JSON VariableType = "json"
)
