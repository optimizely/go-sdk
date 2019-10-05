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
var Godogs int

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

	s.Step(`^the datafile is "([^"]*)"$`, context.TheDatafileIs)
	s.Step(`^(\d+) "([^"]*)" listener is added$`, context.ListenerIsAdded)
	s.Step(`^([^\\\"]*) is called with arguments$`, context.IsCalledWithArguments)
	s.Step(`^the result should be (?:string )?"([^"]*)"$`, context.TheResultShouldBeString)
	s.Step(`^the result should be (?:integer )?(\d+)$`, context.TheResultShouldBeInteger)
	s.Step(`^the result should be (?:double )?(\d+)\.(\d+)$`, context.TheResultShouldBeFloat)
	s.Step(`^the result should be boolean "([^"]*)"$`, context.TheResultShouldBeBoolean)
	s.Step(`^the result should be \'false\'$`, context.TheResultShouldBeFalse)
	s.Step(`^the result should match list "([^"]*)"$`, context.TheResultShouldMatchList)
	s.Step(`^in the response, "([^"]*)" should be "([^"]*)"$`, context.InTheResponseKeyShouldBeObject)
	s.Step(`^in the response, "([^"]*)" should match$`, context.InTheResponseShouldMatch)
	s.Step(`^in the response, "([^"]*)" should have each one of these$`, context.InTheResponseShouldHaveEachOneOfThese)
	s.Step(`^there are no dispatched events$`, context.ThereAreNoDispatchedEvents)
	s.Step(`^dispatched events payloads include$`, context.DispatchedEventsPayloadsInclude)
	s.BeforeScenario(func(interface{}) {
		context.Reset()
	})
}
