# Optimizely Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/optimizely/go-sdk)](https://goreportcard.com/report/github.com/optimizely/go-sdk)
[![Coverage Status](https://coveralls.io/repos/github/optimizely/go-sdk/badge.svg?branch=go-alpha)](https://coveralls.io/github/optimizely/go-sdk?branch=go-alpha)

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
  help               Help about any command
  is_feature_enabled Is feature enabled?

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
