module github.com/optimizely/go-sdk

go 1.12

require (
	github.com/DATA-DOG/godog v0.7.13
	github.com/google/uuid v1.1.1
	github.com/json-iterator/go v1.1.7
	github.com/optimizely/subset v0.0.0-20191016231344-a0cb13c639ec
	github.com/pkg/profile v1.3.0
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/twmb/murmur3 v1.0.0
	gopkg.in/yaml.v3 v3.0.0-20191026110619-0b21df46bc1d
)

// Work around issue wtih git.apache.org/thrift.git
replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
