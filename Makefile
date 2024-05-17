BINARY=bin/monitoring-system.out

CMD_DIR=./cmd
CONFIG_DIR=./config
INTERNAL_DIR=./internal
PKG_DIR=./pkg

GO=go
GOFMT=gofmt

CONFIG_FILE=config/config.go

PKGS=$(shell $(GO) list ./... | grep -v /vendor/)

include .env

all: build

build: fmt vet
	$(GO) build -o $(BINARY) $(CMD_DIR)/main.go

run: build
	AWS_REGION=$(AWS_REGION) S3_BUCKET_NAME=$(S3_BUCKET_NAME) HOST=$(HOST) PORT=$(PORT) $(BINARY)

fmt:
	$(GOFMT) -w $(CMD_DIR) $(CONFIG_DIR) $(INTERNAL_DIR) $(PKG_DIR)

vet:
	$(GO) vet $(PKGS)

test:
	$(GO) test -v $(PKGS)

clean:
	$(GO) clean
	rm -f $(BINARY)

deps:
	$(GO) get -u ./...

.PHONY: all build run fmt vet test clean deps env