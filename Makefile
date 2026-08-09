.PHONY: all build build-git-wt build-windows test lint fmt vet clean install install-git-hooks install-git-wt

BINARY_NAME=kit
WORKTREE_BINARY_NAME=git-wt
GIT_WT_PREFIX?=$(HOME)/.local
GIT_WT_BIN_DIR?=$(GIT_WT_PREFIX)/bin
GIT_WT_MAN_DIR?=$(GIT_WT_PREFIX)/share/man/man1
VERSION?=$(shell git describe --tags --abbrev=0 --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo dev)
LDFLAGS=-ldflags "-X github.com/jamesonstone/kit/pkg/agentcli.Version=$(VERSION)"

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
	mkdir -p $(GIT_WT_BIN_DIR) $(GIT_WT_MAN_DIR)
	install -m 0755 bin/$(WORKTREE_BINARY_NAME) $(GIT_WT_BIN_DIR)/$(WORKTREE_BINARY_NAME)
	install -m 0644 docs/man/git-wt.1 $(GIT_WT_MAN_DIR)/git-wt.1

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
