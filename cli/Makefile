CLI_NAME := go-infra-cli
VERSION ?= dev
BIN_DIR := bin
CLI_PKG := ./cmd/go-infra-cli
LDFLAGS := -X main.version=$(VERSION)

.PHONY: cli-build cli-install cli-run cli-version

cli-build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(CLI_NAME) $(CLI_PKG)

cli-install:
	go install -ldflags "$(LDFLAGS)" $(CLI_PKG)

cli-run:
	go run -ldflags "$(LDFLAGS)" $(CLI_PKG)

cli-version:
	go run -ldflags "$(LDFLAGS)" $(CLI_PKG) version

