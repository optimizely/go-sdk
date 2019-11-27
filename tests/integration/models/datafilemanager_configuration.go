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

package models

// DataFileManagerConfiguration represents a datafile manager configuration
type DataFileManagerConfiguration struct {
	SDKKey            string `yaml:"sdk_key"`
	Mode              string `yaml:"mode,omitempty"`
	Revision          *int   `yaml:"revision,omitempty"`
	DatafileCondition string `yaml:"datafile_condition,omitempty"`
	UpdateInterval    *int   `yaml:"update_interval,omitempty"`
	Timeout           *int   `yaml:"timeout,omitempty"`
}
