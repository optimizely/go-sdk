/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package utils //
package utils

import "time"

// ODPEventsAPIEndpointPath defines the endpoint path for events api
const ODPEventsAPIEndpointPath = "/v3/events"

// OdpAPIKeyHeader defines key for designating the ODP API public key
const OdpAPIKeyHeader = "x-api-key"

// OdpEventType holds the value for the odp event type
const OdpEventType = "fullstack"

// OdpFSUserIDKey holds the key for the odp fullstack/feature experimentation userID
const OdpFSUserIDKey = "fs_user_id"

// OdpActionIdentified holds the value for identified action type
const OdpActionIdentified = "identified"

// DefaultBatchSize holds the default value for the batch size
const DefaultBatchSize = 10

// DefaultEventQueueSize holds the default value for the event queue size
const DefaultEventQueueSize = 10000

// DefaultEventFlushInterval holds the default value for the event flush interval
const DefaultEventFlushInterval = 1 * time.Second

// DefaultSegmentsCacheSize holds the default value for the segments cache size
const DefaultSegmentsCacheSize = 10000

// DefaultSegmentsCacheTimeout holds the default value for the segments cache timeout
const DefaultSegmentsCacheTimeout = 10 * time.Minute // 10 minutes

// DefaultOdpEventTimeout holds the default value for the odp event timeout
const DefaultOdpEventTimeout = 10 * time.Second

// DefaultSegmentFetchTimeout holds the default value for the segment fetch timeout
const DefaultSegmentFetchTimeout = 10 * time.Second
