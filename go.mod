module github.com/optimizely/go-sdk

go 1.12

require (
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-multierror v1.1.0
	github.com/json-iterator/go v1.1.7
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.3.0
	github.com/stretchr/testify v1.4.0
	github.com/twmb/murmur3 v1.0.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
)

// Work around issue wtih git.apache.org/thrift.git
replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
