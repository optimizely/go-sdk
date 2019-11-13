# The name of the executable (default is current directory name)
TARGET := $(shell basename "$(PWD)")
VERSION ?= $(shell git describe --tags)
.DEFAULT_GOAL := help

# Go parameters
GO111MODULE:=on
GOCMD=go
GOBIN=bin
GOPATH=$(shell $(GOCMD) env GOPATH)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint
BINARY_UNIX=$(TARGET)_unix

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

install: ## installs dev and ci dependencies
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOPATH)/bin v1.19.0

lint:
	$(GOLINT) run --out-format=tab --tests=false pkg/...

test: ## recursively test source code in pkg
	GO111MODULE=$(GO111MODULE) $(GOTEST) ./pkg/...

test-ci: ## run unit tests in CI mode
	GO111MODULE=$(GO111MODULE) $(GOTEST) -v -race ./pkg/... -coverprofile=profile.cov