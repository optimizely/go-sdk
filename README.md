# Optimizely Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/optimizely/go-sdk)](https://goreportcard.com/report/github.com/optimizely/go-sdk)
[![Coverage Status](https://coveralls.io/repos/github/optimizely/go-sdk/badge.svg?branch=master)](https://coveralls.io/github/optimizely/go-sdk?branch=master)

## Usage

### Instantiation
To start using the SDK, create an instance using our factory method:

```
import "github.com/optimizely/go-sdk/optimizely/client"

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
  "github.com/optimizely/go-sdk/optimizely/client"
  "github.com/optimizely/go-sdk/optimizely/entities"
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

## Command line interface
A CLI has been provided to illustrate the functionality of the SDK. Simply run `go-sdk` for help.
```$sh
go-sdk provides cli access to your Optimizely fullstack project

Usage:
  go-sdk [command]

Available Commands:
  help                          Help about any command
  is_feature_enabled            Is feature enabled?
  get_enabled_features          Get enabled features
  track                         Track a conversion event
  get_feature_variable_boolean  Get feature variable boolean value
  get_feature_variable_double   Get feature variable double value
  get_feature_variable_integer  Get feature variable integer value
  get_feature_variable_string   Get feature variable string value

Flags:
  -h, --help            help for go-sdk
  -s, --sdkKey string   Optimizely project SDK key

Use "go-sdk [command] --help" for more information about a command.
```

Each supported SDK API method is it's own [cobra](https://github.com/spf13/cobra) command and requires the
input of an `--sdkKey`.

### Installation
Install the CLI from github:

```$sh
go install github.com/optimizely/go-sdk
```

Install the CLI from source:
```$sh
go get github.com/optimizely/go-sdk
cd $GOPATH/src/github.com/optimizely/go-sdk
go install
```

NOTE:
We practice trunk-based development, and as such our default branch, `master` might not always be the most stable. We do tag releases on Github and you can pin your installation to those particular release versions. One way to do this is to use [*Go Modules*](https://blog.golang.org/using-go-modules) for managing external dependencies:

```
module mymodule

go 1.12

require (
	github.com/optimizely/go-sdk v0.2.0
)
```

If you are already using `go.mod` in your application you can run the following:

```
go mod edit -require github.com/optimizely/go-sdk@v0.2.0
```

NOTE:
```$sh
go get github.com/optimizely/go-sdk/...
```
or
```$sh
go get github.com/optimizely/go-sdk/optimizely
```
will install it as a package to pkg directory, rather than src directory. It could be useful for future development and vendoring.

### Commands

#### is_feature_enabled
```
Determines if a feature is enabled

Usage:
  go-sdk is_feature_enabled [flags]

Flags:
  -f, --featureKey string   feature key to enable
  -h, --help                help for is_feature_enabled
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### get_enabled_features
```
Returns enabled features for userId

Usage:
  go-sdk get_enabled_features [flags]

Flags:
  -h, --help                help for get_enabled_features
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### track
```
Tracks a conversion event

Usage:
  go-sdk track [flags]

Flags:
  -e, --eventKey string     event key to track
  -h, --help                help for track
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### get_feature_variable_boolean
```
Returns feature variable boolean value

Usage:
  go-sdk get_feature_variable_boolean [flags]

Flags:
  -f, --featureKey string   feature key for feature
  -v, --variableKey string  variable key for feature variable
  -h, --help                help for get_feature_variable_boolean
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### get_feature_variable_double
```
Returns feature variable double value

Usage:
  go-sdk get_feature_variable_double [flags]

Flags:
  -f, --featureKey string   feature key for feature
  -v, --variableKey string  variable key for feature variable
  -h, --help                help for get_feature_variable_double
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### get_feature_variable_integer
```
Returns feature variable integer value

Usage:
  go-sdk get_feature_variable_integer [flags]

Flags:
  -f, --featureKey string   feature key for feature
  -v, --variableKey string  variable key for feature variable
  -h, --help                help for get_feature_variable_integer
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
  ```

#### get_feature_variable_string
```
Returns feature variable string value

Usage:
  go-sdk get_feature_variable_string [flags]

Flags:
  -f, --featureKey string   feature key for feature
  -v, --variableKey string  variable key for feature variable
  -h, --help                help for get_feature_variable_string
  -u, --userId string       user id

Global Flags:
  -s, --sdkKey string   Optimizely project SDK key
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
