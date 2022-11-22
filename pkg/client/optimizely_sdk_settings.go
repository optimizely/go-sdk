/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                   		*
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

// Package client //
package client

// OptimizelySdkSettings represents options controlling odp manager.
type OptimizelySdkSettings struct {
	// The maximum size of audience segments cache - cache is disabled if this is set to zero.
	// Default value is 10000
	SegmentsCacheSize int
	// The timeout in seconds of audience segments cache - timeout is disabled if this is set to zero.
	// Default value is 600s
	SegmentsCacheTimeoutInSecs int64
	// ODP features are disabled if this is set to true.
	// Default value is false
	DisableOdp bool
}
