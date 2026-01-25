BIN_DIR := bin
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

BINARY ?= $(notdir $(abspath $(CURDIR)))
CMD_PKG := ./cmd/$(BINARY)

.PHONY: all build test fmt lint check install uninstall clean

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_PKG)

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	go vet ./...

check: fmt test

install: build
	mkdir -p "$(BINDIR)"
	$(INSTALL) -m 0755 $(BIN_DIR)/$(BINARY) "$(BINDIR)/$(BINARY)"
	@echo "Installed $(BINDIR)/$(BINARY). Ensure $(BINDIR) is on PATH."

uninstall:
	rm -f "$(BINDIR)/$(BINARY)"
	@echo "Removed $(BINDIR)/$(BINARY) if it existed."

clean:
	rm -rf $(BIN_DIR)
