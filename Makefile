BINARY=bin/monitoring-system.out

CMD_DIR=./src/cmd
CONFIG_DIR=./src/config
INTERNAL_DIR=./src/internal
PKG_DIR=./src/pkg
WEB_DIR=./src/web

BIN_INSTALL_DIR=/usr/bin/monitoring-system

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

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

deploy: build
	sudo mkdir -p $(BIN_INSTALL_DIR) $(BIN_INSTALL_DIR)/web /etc/monitoring-system /usr/share/monitoring-system
	sudo cp $(BINARY) $(BIN_INSTALL_DIR)
	sudo chmod +x $(BIN_INSTALL_DIR)/monitoring-system.out
	sudo cp -r $(WEB_DIR)/* $(BIN_INSTALL_DIR)/web

	sudo cp monitoring-system.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable monitoring-system.service
	sudo systemctl restart monitoring-system.service

.PHONY: all build run test clean deps deploy
