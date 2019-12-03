# Optimizely Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/optimizely/go-sdk)](https://goreportcard.com/report/github.com/optimizely/go-sdk)
[![Coverage Status](https://coveralls.io/repos/github/optimizely/go-sdk/badge.svg?branch=master)](https://coveralls.io/github/optimizely/go-sdk?branch=master)

## Installation

### Install from github:

```$sh
go get github.com/optimizely/go-sdk
```

### Install from source:
```$sh
go get github.com/optimizely/go-sdk
cd $GOPATH/src/github.com/optimizely/go-sdk
go install
```

NOTE:
We practice trunk-based development, and as such our default branch, `master` might not always be the most stable. We do tag releases on Github and you can pin your installation to those particular release versions. One way to do this is to use [*Go Modules*](https://blog.golang.org/using-go-modules) for managing external dependencies:

### Install using go.mod:

```
module mymodule

go 1.12

require (
	github.com/optimizely/go-sdk v1.0.0-beta7
)
```

If you are already using `go.mod` in your application you can run the following:

```
go mod edit -require github.com/optimizely/go-sdk@v1.0.0-beta7
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

## Usage

### Instantiation
To start using the SDK, create an instance using our factory method:

```
import "github.com/optimizely/go-sdk/pkg/client"

optimizelyFactory := &client.OptimizelyFactory{
  SDKKey: "[SDK_KEY_HERE]",
}

client, err := optimizelyFactory.Client()

// You can also instantiate with a hard-coded datafile
optimizelyFactory := &client.OptimizelyFactory{
	Datafile: []byte("datafile_string"),
}

client, err := optimizelyFactory.Client()

```

### Feature Rollouts
```
import (
  "github.com/optimizely/go-sdk/pkg/client"
  "github.com/optimizely/go-sdk/pkg/entities"
)

user := entities.UserContext{
  ID: "optimizely end user",
  Attributes: map[string]interface{}{
    "state":      "California",
    "likes_donuts": true,
  },
}

enabled, _ := client.IsFeatureEnabled("binary_feature", user)
```

## Credits

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
