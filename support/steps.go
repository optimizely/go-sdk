package support

import (
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

type Listener struct {
	numberOfListeners int
	listenerName      string
}
type RequestParams struct {
	apiToCall    string
	datafileName string
	listener     *Listener
	// listerName   string // Need to replace with listener.
}

type ResponseParams struct {
	result interface{}
}

type Context struct {
	requestParams  RequestParams
	responseParams ResponseParams
}

func (c *Context) TheDatafileIs(datafileName string) error {

	c.requestParams.datafileName = datafileName

	return nil
}

func (c *Context) ListenerIsAdded(numberOfListeners int, ListenerName string) error {
	listener := new(Listener)
	listener.numberOfListeners = numberOfListeners
	listener.listenerName = ListenerName

	// Need to check assigning address
	c.requestParams.listener = listener

	return nil
}

func (c *Context) IsFeatureEnabledIsCalledWithArguments(arg1 *gherkin.DocString) error {
	return godog.ErrPending
}

func (c *Context) TheResultShouldBe(arg1 string) error {
	return godog.ErrPending
}

func (c *Context) InTheResponseShouldBe(arg1 string) error {
	return godog.ErrPending
}

func (c *Context) InTheResponseKeyShouldBeEqualsObject(arg1, arg2 string) error {
	return godog.ErrPending
}

func (c *Context) ThereAreNoDispatchedEvents() error {
	return godog.ErrPending
}
