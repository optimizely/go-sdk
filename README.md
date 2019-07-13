# Optimizely Go SDK

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