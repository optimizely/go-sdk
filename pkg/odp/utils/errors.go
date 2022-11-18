/****************************************************************************
 * Copyright 2022, Optimizely, Inc. and contributors                        *
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

// Package utils //
package utils

// InvalidSegmentIdentifier error string when fetch failed with invalid identifier
const InvalidSegmentIdentifier = "audience segments fetch failed (invalid identifier)"

// FetchSegmentsFailedError error string when fetch failed with provided reason
const FetchSegmentsFailedError = "audience segments fetch failed (%s)"

// OdpNotEnabled error string when odp is not enabled
const OdpNotEnabled = "ODP is not enabled"

// IdentityOdpDisabled error string when odp event is not dispatched as odp is disabled
const IdentityOdpDisabled = "ODP identify event is not dispatched (ODP disabled)"

// IdentityOdpNotIntegrated error string when odp event is not dispatched as odp is not integrated
const IdentityOdpNotIntegrated = "ODP identify event is not dispatched (ODP not integrated)"

// OdpNotIntegrated error string when odp is not integrated
const OdpNotIntegrated = "ODP not integrated"

// OdpEventFailed error string when odp event failed with provided reason
const OdpEventFailed = "ODP event send failed (%s)"

// OdpInvalidData error string when odp event data is invalid
const OdpInvalidData = "ODP data is not valid"
