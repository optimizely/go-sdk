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

package main

import (
	"flag"
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/optimizely/go-sdk/tests/integration/support"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {

	context := new(support.ScenarioCtx)
	// Resetting context before each scenario
	s.BeforeScenario(func(interface{}) {
		context.Reset()
	})
	s.Step(`^the datafile is "([^"]*)"$`, context.TheDatafileIs)
	s.Step(`^(\d+) "([^"]*)" listener is added$`, context.ListenerIsAdded)
	s.Step(`^([^\\\"]*) is called with arguments$`, context.IsCalledWithArguments)
	s.Step(`^the result should be (?:string )?"([^"]*)"$`, context.TheResultShouldBeString)
	s.Step(`^the result should be (?:integer )?(\d+)$`, context.TheResultShouldBeInteger)
	s.Step(`^the result should be (?:double )?(\d+)\.(\d+)$`, context.TheResultShouldBeFloat)
	s.Step(`^the result should be boolean "([^"]*)"$`, context.TheResultShouldBeTypedBoolean)
	s.Step(`^the result should be \'([^"]*)\'$`, context.TheResultShouldBeBoolean)
	s.Step(`^the result should match list "([^"]*)"$`, context.TheResultShouldMatchList)
	s.Step(`^in the response, "([^"]*)" should be "([^"]*)"$`, context.InTheResponseKeyShouldBeObject)
	s.Step(`^in the response, "([^"]*)" should match$`, context.InTheResponseShouldMatch)
	s.Step(`^in the response, "([^"]*)" should have this exactly (\d+) times$`, context.ResponseShouldHaveThisExactlyNTimes)
	s.Step(`^in the response, "([^"]*)" should have each one of these$`, context.InTheResponseShouldHaveEachOneOfThese)
	s.Step(`^the number of dispatched events is (\d+)$`, context.TheNumberOfDispatchedEventsIs)
	s.Step(`^there are no dispatched events$`, context.ThereAreNoDispatchedEvents)
	s.Step(`^dispatched events payloads include$`, context.DispatchedEventsPayloadsInclude)
	s.Step(`^payloads of dispatched events don\'t include decisions$`, context.PayloadsOfDispatchedEventsDontIncludeDecisions)
}
