all: build

.PHONY: build
build:
	$(MAKE) build-amd64
	$(MAKE) build-arm64

.PHONY: build-amd64
build-amd64:
	GO111MODULE=on GOARCH=amd64 go build -trimpath -tags "$(GO_BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" -o "./bin/faux-amd64$(BINARY_EXT)" ./cmd

.PHONY: build-arm64
build-arm64:
	GO111MODULE=on GOARCH=arm64 go build -trimpath -tags "$(GO_BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" -o "./bin/faux-arm64$(BINARY_EXT)" ./cmd

.PHONY: test
test:
	go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

.PHONY: lint
lint:
	golangci-lint run

.PHONY: run
run:
	go run ./cmd/main.go