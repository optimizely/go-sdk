/****************************************************************************
 * Copyright 2022,2024 Optimizely, Inc. and contributors                    *
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

package client

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OptimizelyUserContextODPDecideTestSuite struct {
	suite.Suite
	userID           string
	flagKey          string
	doOnce           sync.Once // required since we only need to read datafile once
	datafile         []byte
	factory          OptimizelyFactory
	optimizelyClient *OptimizelyClient
}

func (o *OptimizelyUserContextODPDecideTestSuite) SetupTest() {
	o.doOnce.Do(func() {
		absPath, _ := filepath.Abs("../../test-data/odp-test-datafile.json")
		o.datafile, _ = os.ReadFile(absPath)
	})
	o.userID = "tester"
	o.flagKey = "flag-segment"
	o.factory = OptimizelyFactory{Datafile: o.datafile}
	o.optimizelyClient, _ = o.factory.Client()
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsSegmentHitInABTest() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.SetQualifiedSegments([]string{"odp-segment-1", "odp-segment-none"})
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("variation-a", decision.VariationKey)
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsOtherAudienceHitInABTest() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, map[string]interface{}{"age": 30})
	userContext.SetQualifiedSegments([]string{"odp-segment-none"})
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("variation-a", decision.VariationKey)
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsSegmentHitInRollout() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.SetQualifiedSegments([]string{"odp-segment-2"})
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("rollout-variation-on", decision.VariationKey)
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsSegmentMissInRollout() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.SetQualifiedSegments([]string{"odp-segment-none"})
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("rollout-variation-off", decision.VariationKey)
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsEmptySegments() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, nil)
	userContext.SetQualifiedSegments([]string{})
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("rollout-variation-off", decision.VariationKey)
}

func (o *OptimizelyUserContextODPDecideTestSuite) TestDecideWithQualifiedSegmentsDefault() {
	userContext := o.optimizelyClient.CreateUserContext(o.userID, nil)
	decision := userContext.Decide(context.Background(), o.flagKey, nil)
	o.Equal("rollout-variation-off", decision.VariationKey)
}

func TestOptimizelyUserContextODPDecideTestSuite(t *testing.T) {
	suite.Run(t, new(OptimizelyUserContextODPDecideTestSuite))
}
