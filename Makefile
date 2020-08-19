# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
.DEFAULT_GOAL := help

# Go parameters
GO111MODULE:=on
GOCMD=go
GOPATH=$(shell $(GOCMD) env GOPATH)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=$(GOPATH)/bin/golangci-lint

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

clean: ## runs `go clean` and removes the bin/ dir
	GO111MODULE=$(GO111MODULE) $(GOCLEAN) --modcache
	rm -rf $(GOBIN)

cover: ## run unit tests with coverage
	GO111MODULE=$(GO111MODULE) $(GOTEST) -race ./pkg/... -coverprofile=profile.cov

install: ## installs dev and ci dependencies
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.19.0

lint: ## runs `golangci-lint` linters defined in `.golangci.yml` file
	$(GOLINT) run --out-format=tab --tests=false pkg/...

test: ## recursively test source code in pkg without coverage
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./pkg/...

benchmark: ## recursively test source code in pkg without coverage
	GO111MODULE=$(GO111MODULE) $(GOTEST) -bench=. -run=^a ./pkg/...

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
