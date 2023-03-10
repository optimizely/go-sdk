# Optimizely Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/optimizely/go-sdk)](https://goreportcard.com/report/github.com/optimizely/go-sdk)
[![Coverage Status](https://coveralls.io/repos/github/optimizely/go-sdk/badge.svg?branch=master)](https://coveralls.io/github/optimizely/go-sdk?branch=master)


This repository houses the Go SDK for use with Optimizely Feature Experimentation and Optimizely Full Stack (legacy).

Optimizely Feature Experimentation is an A/B testing and feature management tool for product development teams that enables you to experiment at every step. Using Optimizely Feature Experimentation allows for every feature on your roadmap to be an opportunity to discover hidden insights. Learn more at [Optimizely.com](https://www.optimizely.com/products/experiment/feature-experimentation/), or see the [developer documentation](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/welcome).

Optimizely Rollouts is [free feature flags](https://www.optimizely.com/free-feature-flagging/) for development teams. You can easily roll out and roll back features in any application without code deploys, mitigating risk for every feature on your roadmap.

## Get Started

Refer to the [Go SDK's developer documentation](https://docs.developers.optimizely.com/experimentation/v4.0.0-full-stack/docs/go-sdk) for detailed instructions on getting started with using the SDK.
  
### Requirements  

Requires Golang version 1.13 or higher.

### Install the SDK

#### Install from github:

```$sh
go get github.com/optimizely/go-sdk
```

#### Install from source:
```$sh
go get github.com/optimizely/go-sdk
cd $GOPATH/src/github.com/optimizely/go-sdk
go install
```

NOTE:
We practice trunk-based development, and as such our default branch, `master` might not always be the most stable. We do tag releases on Github and you can pin your installation to those particular release versions. One way to do this is to use [*Go Modules*](https://blog.golang.org/using-go-modules) for managing external dependencies:

#### Install using go.mod:

```
module mymodule

go 1.13

require (
	github.com/optimizely/go-sdk v1.8.3
)
```

If you are already using `go.mod` in your application you can run the following:

```
go mod edit -require github.com/optimizely/go-sdk@v1.8.3
```

NOTE:
```$sh
go get github.com/optimizely/go-sdk/...
```
or
```$sh
go get github.com/optimizely/go-sdk/pkg
```
will install it as a package to pkg directory, rather than src directory. It could be useful for future development and vendoring.


## Use the Go SDK

See the example file in examples/main.go.

### Initialization

```
import optly "github.com/optimizely/go-sdk"
import "github.com/optimizely/go-sdk/client"

// Simple one-line initialization with the SDK key
client, err := optly.Client("SDK_KEY")

// You can also instantiate with a hard-coded datafile using our client factory method
optimizelyFactory := &client.OptimizelyFactory{
	Datafile: []byte("datafile_string"),
}

client, err = optimizelyFactory.Client()

```
### Make Decisions
```
import (
  optly "github.com/optimizely/go-sdk"
)

// instantiate a client
client, err := optly.Client("SDK_KEY")

// User attributes are optional and used for targeting and results segmentation
atributes := map[string]interface{}{
     "state":      "California",
     "likes_donuts": true,
}, 
user := optly.UserContext("optimizely end user", attributes)
options := []decide.OptimizelyDecideOptions{decide.IncludeReasons}
decision := userCtx.Decide("my_flag", options)

var variationKey string
if variationKey = decision.VariationKey; variationKey == "" {
  fmt.Printf("[decide] error: %v", decision.GetReasons())
  return
}
if variationKey == "control" {
  // Execute code for control variation
} else if variationKey == "treatment" {
  // Execute code for treatment variation
}

}
```

## SDK Development

### Unit Tests

Run 
``` 
make test 
```

### Contributing

Please see [CONTRIBUTING](https://github.com/optimizely/go-sdk/blob/master/CONTRIBUTING.md).

### Credits

This software is distributed with code from the following open source projects:

murmur3
Copyright 2013, Sébastien Paolacci.
License (BSD-3 Clause): https://github.com/twmb/murmur3/blob/master/LICENSE

uuid
Copyright (c) 2009, 2014 Google Inc. All rights reserved.
License (BSD-3 Clause): https://github.com/google/uuid/blob/master/LICENSE

testify
Copyright (c) 2012-2018 Mat Ryer and Tyler Bunnell.
License (MIT): https://github.com/stretchr/testify/blob/master/LICENSE

json-iterator
Copyright (c) 2016 json-iterator
License (MIT): https://github.com/json-iterator/go/blob/master/LICENSE

subset
Copyright (c) 2015, Facebook, Inc. All rights reserved.
License (BSD): https://github.com/facebookarchive/subset/blob/master/license

profile
Copyright (c) 2013 Dave Cheney. All rights reserved.
License (BSD): https://github.com/pkg/profile/blob/master/LICENSE

sync
Copyright (c) 2009 The Go Authors. All rights reserved.
https://github.com/golang/sync/blob/master/LICENSE

### Other Optimizely SDKs

- Agent - https://github.com/optimizely/agent

- Android - https://github.com/optimizely/android-sdk

- C# - https://github.com/optimizely/csharp-sdk

- Flutter - https://github.com/optimizely/optimizely-flutter-sdk

- Java - https://github.com/optimizely/java-sdk

- JavaScript - https://github.com/optimizely/javascript-sdk

- PHP - https://github.com/optimizely/php-sdk

- Python - https://github.com/optimizely/python-sdk

- React - https://github.com/optimizely/react-sdk

- Ruby - https://github.com/optimizely/ruby-sdk

- Swift - https://github.com/optimizely/swift-sdk