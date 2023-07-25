PKG := github.com/iamthen0ise/faux
VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty='.m' --always --tags)

GO_LDFLAGS ?= -w -X ${PKG}/internal.Version=${VERSION}
GO_BUILDTAGS ?= e2e
DRIVE_PREFIX?=
ifeq ($(OS),Windows_NT)
    DETECTED_OS = Windows
    DRIVE_PREFIX=C:
else
    DETECTED_OS = $(shell uname -s)
endif

ifeq ($(DETECTED_OS),Windows)
	BINARY_EXT=.exe
endif

BUILD_FLAGS?=
TEST_FLAGS?=

DESTDIR ?=

all: build

.PHONY: build
build:
	GO111MODULE=on go build $(BUILD_FLAGS) -trimpath -tags "$(GO_BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" -o "$(or $(DESTDIR),./bin/build)/faux$(BINARY_EXT)" ./cmd

.PHONY: test
test:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

.PHONY: lint
lint:
	go golangci-lint

help:
	@echo Please specify a build target. The choices are:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
