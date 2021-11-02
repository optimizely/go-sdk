/****************************************************************************
 * Copyright 2021, Optimizely, Inc. and contributors                        *
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

// Package decision //
package decision

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/optimizely/go-sdk/pkg/config"
	"github.com/optimizely/go-sdk/pkg/decide"
	"github.com/stretchr/testify/suite"
)

var doOnce sync.Once // required since we only need to read datafile once
var datafile []byte

type ForcedDecisionServiceTestSuite struct {
	suite.Suite
	forcedDecisionService *ForcedDecisionService
	projectConfig         config.ProjectConfig
}

func (s *ForcedDecisionServiceTestSuite) SetupTest() {
	s.forcedDecisionService = NewForcedDecisionService("abc")
	doOnce.Do(func() {
		absPath, _ := filepath.Abs("../../test-data/decide-test-datafile.json")
		datafile, _ = ioutil.ReadFile(absPath)
	})

	configManager := config.NewStaticProjectConfigManagerWithOptions("", config.WithInitialDatafile(datafile))
	s.projectConfig, _ = configManager.GetConfig()
}

func (s *ForcedDecisionServiceTestSuite) TestSetForcedDecision() {
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: ""}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: ""}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: ""}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: ""}))
}

func (s *ForcedDecisionServiceTestSuite) TestGetForcedDecision() {
	forcedDecision, err := s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("3", forcedDecision.VariationKey)
	s.NoError(err)

	// Check if changing forcedDecision changes the decision inside the forced decision map
	forcedDecision.VariationKey = "4"
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("3", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"})
	s.Equal("3", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""})
	s.Equal("3", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""})
	s.Equal("3", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)
}

func (s *ForcedDecisionServiceTestSuite) TestRemoveForcedDecision() {
	forcedDecision, err := s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.Equal("decision not found", err.Error())

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.RemoveForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.RemoveForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "2"})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.RemoveForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: ""})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.RemoveForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}))
	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""})
	s.Equal("", forcedDecision.VariationKey)
	s.NoError(err)
}

func (s *ForcedDecisionServiceTestSuite) TestRemoveAllForcedDecision() {
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "a"}, OptimizelyForcedDecision{VariationKey: "b"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "2", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "b"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: "a"}, OptimizelyForcedDecision{VariationKey: "b"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "b"}))

	s.Len(s.forcedDecisionService.forcedDecisions, 4)
	s.True(s.forcedDecisionService.RemoveAllForcedDecisions())
	s.Len(s.forcedDecisionService.forcedDecisions, 0)

	forcedDecision, err := s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "a"})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)

	forcedDecision, err = s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: "2", RuleKey: ""})
	s.Equal("", forcedDecision.VariationKey)
	s.Error(err)
}

func (s *ForcedDecisionServiceTestSuite) TestFindValidatedForcedDecision() {
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "a"}))
	variation, reasons, err := s.forcedDecisionService.FindValidatedForcedDecision(s.projectConfig, OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: ""}, &decide.Options{IncludeReasons: true})
	s.NoError(err)
	s.Len(reasons.ToReport(), 1)
	s.Equal("Variation (a) is mapped to flag (feature_1) and user (abc) in the forced decision map.", reasons.ToReport()[0])
	s.Equal("a", variation.Key)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: "exp_with_audience"}, OptimizelyForcedDecision{VariationKey: "a"}))
	variation, reasons, err = s.forcedDecisionService.FindValidatedForcedDecision(s.projectConfig, OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: "exp_with_audience"}, &decide.Options{IncludeReasons: true})
	s.NoError(err)
	s.Len(reasons.ToReport(), 1)
	s.Equal("Variation (a) is mapped to flag (feature_1), rule (exp_with_audience) and user (abc) in the forced decision map.", reasons.ToReport()[0])
	s.Equal("a", variation.Key)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "feature_2", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "variation_with_traffic"}))
	variation, reasons, err = s.forcedDecisionService.FindValidatedForcedDecision(s.projectConfig, OptimizelyDecisionContext{FlagKey: "feature_2", RuleKey: ""}, &decide.Options{IncludeReasons: true})
	s.NoError(err)
	s.Len(reasons.ToReport(), 1)
	s.Equal("Variation (variation_with_traffic) is mapped to flag (feature_2) and user (abc) in the forced decision map.", reasons.ToReport()[0])
	s.Equal("variation_with_traffic", variation.Key)

	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: ""}, OptimizelyForcedDecision{VariationKey: "fake"}))
	variation, reasons, err = s.forcedDecisionService.FindValidatedForcedDecision(s.projectConfig, OptimizelyDecisionContext{FlagKey: "feature_1", RuleKey: ""}, &decide.Options{IncludeReasons: true})
	s.Error(err)
	s.Len(reasons.ToReport(), 1)
	s.Equal("Invalid variation is mapped to flag (feature_1) and user (abc) in the forced decision map.", reasons.ToReport()[0])
	s.Nil(variation)

	variation, reasons, err = s.forcedDecisionService.FindValidatedForcedDecision(s.projectConfig, OptimizelyDecisionContext{FlagKey: "feature_3", RuleKey: ""}, &decide.Options{IncludeReasons: true})
	s.Error(err)
	s.Len(reasons.ToReport(), 0)
	s.Nil(variation)
}

func (s *ForcedDecisionServiceTestSuite) TestCreateCopy() {
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: "3"}))
	s.True(s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: "1", RuleKey: "2"}, OptimizelyForcedDecision{VariationKey: ""}))

	ucCopy := s.forcedDecisionService.CreateCopy()
	s.Equal(len(s.forcedDecisionService.forcedDecisions), len(ucCopy.forcedDecisions))

	ucCopy.RemoveAllForcedDecisions()
	s.NotEqual(len(s.forcedDecisionService.forcedDecisions), len(ucCopy.forcedDecisions))
}

func (s *ForcedDecisionServiceTestSuite) TestAsyncBehaviour() {
	var wg sync.WaitGroup
	wg.Add(2)

	setForcedDecisions := func() {
		i := 0
		for i < 100 {
			s.forcedDecisionService.SetForcedDecision(OptimizelyDecisionContext{FlagKey: fmt.Sprint(i), RuleKey: "b"}, OptimizelyForcedDecision{VariationKey: "c"})
			i++
		}
		wg.Done()
	}

	getForcedDecisions := func() {
		i := 0
		for i < 100 {
			s.forcedDecisionService.GetForcedDecision(OptimizelyDecisionContext{FlagKey: fmt.Sprint(i), RuleKey: "b"})
			i++
		}
		wg.Done()
	}

	removeAllForcedDecisions := func() {
		s.forcedDecisionService.RemoveAllForcedDecisions()
		wg.Done()
	}

	go setForcedDecisions()
	go getForcedDecisions()
	wg.Wait()
	s.Len(s.forcedDecisionService.forcedDecisions, 100)

	wg.Add(2)
	go getForcedDecisions()
	go removeAllForcedDecisions()
	wg.Wait()
	s.Len(s.forcedDecisionService.forcedDecisions, 0)
}

func TestForcedDecisionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ForcedDecisionServiceTestSuite))
}
