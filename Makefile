.PHONY: build run test test-coverage lint clean release-snapshot install

BINARY_NAME=kubepose
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY_NAME) ./cmd/kubepose

run: build
	./$(BINARY_NAME)

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/
	go clean

release-snapshot:
	@which goreleaser > /dev/null || (echo "goreleaser not found. Install from https://goreleaser.com" && exit 1)
	goreleaser release --snapshot --clean

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) || cp $(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME)
