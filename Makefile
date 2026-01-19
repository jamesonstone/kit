.PHONY: build test lint fmt vet clean install

BINARY_NAME=kit
VERSION?=0.1.0
LDFLAGS=-ldflags "-X github.com/jamesonstone/kit/pkg/cli.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/kit

install:
	go install $(LDFLAGS) ./cmd/kit

test:
	go test -v ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf bin/
	go clean

tidy:
	go mod tidy

all: fmt vet test build
