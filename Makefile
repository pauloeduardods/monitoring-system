BINARY=bin/monitoring-system.out

CMD_DIR=./src/cmd
CONFIG_DIR=./src/config
INTERNAL_DIR=./src/internal
PKG_DIR=./src/pkg

GO=go
GOFMT=gofmt

PKGS=$(shell $(GO) list ./... | grep -v /vendor/)

all: build

build:
	$(GO) build -o $(BINARY) $(CMD_DIR)/main.go

run: build
	$(BINARY)

test:
	$(GO) test -v $(PKGS)

clean:
	$(GO) clean
	rm -f $(BINARY)

deps:
	$(GO) get -u ./...

.PHONY: all build run test clean deps env