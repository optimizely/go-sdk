package matchers

const (
	ExactMatchType     = "exact"
	ExistsMatchType    = "exists"
	LtMatchType        = "lt"
	GtMatchType        = "gt"
	SubstringMatchType = "substring"
)

var registry = make(map[string]Matcher)

func init() {
	Register(ExactMatchType, ExactMatcher{})
	Register(ExistsMatchType, ExistsMatcher{})
	Register(LtMatchType, LtMatcher{})
	Register(GtMatchType, GtMatcher{})
	Register(SubstringMatchType, SubstringMatcher{})
}

func Register(name string, matcher Matcher) {
	registry[name] = matcher
}

func Get(name string) (Matcher, bool) {
	matcher, ok := registry[name]
	return matcher, ok
}
