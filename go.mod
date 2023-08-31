module github.com/optimizely/go-sdk

go 1.21.0

require (
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/json-iterator/go v1.1.12
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.7.0
	github.com/stretchr/testify v1.8.2
	github.com/twmb/murmur3 v1.1.6
	golang.org/x/sync v0.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Work around issue with git.apache.org/thrift.git
replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
