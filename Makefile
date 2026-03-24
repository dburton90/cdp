BINARY_NAME=cd_project
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build install clean test lint

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/cd_project

install: build
	./scripts/install.sh

clean:
	rm -f $(BINARY_NAME)
	rm -f ~/.cd_project_folders
	rm -f ~/.cd_project.bash ~/.cd_project.zsh

test:
	go test -v ./...

test-cover:
	go test -cover ./...

lint:
	golangci-lint run

