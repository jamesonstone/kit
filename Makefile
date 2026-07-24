.PHONY: all build build-git-wt build-windows test lint fmt vet clean install install-git-hooks install-git-wt

BINARY_NAME=kit
WORKTREE_BINARY_NAME=git-wt
VERSION?=$(shell git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-X github.com/jamesonstone/kit/pkg/cli.Version=$(VERSION)"

build: install-git-wt
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/kit

build-git-wt:
	go build -o bin/$(WORKTREE_BINARY_NAME) ./cmd/git-wt

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME).exe ./cmd/kit
	GOOS=windows GOARCH=amd64 go build -o bin/$(WORKTREE_BINARY_NAME).exe ./cmd/git-wt

install: install-git-wt
	go install $(LDFLAGS) ./cmd/kit

install-git-wt: build-git-wt
	mkdir -p $(HOME)/.local/bin
	install -m 0755 bin/$(WORKTREE_BINARY_NAME) $(HOME)/.local/bin/$(WORKTREE_BINARY_NAME)

install-git-hooks:
	chmod +x .githooks/pre-commit
	git config core.hooksPath .githooks

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
