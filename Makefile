APP_NAME := bus
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

.PHONY: all build test fmt check install uninstall clean

all: build

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

check: fmt test

install: build
	mkdir -p "$(BINDIR)"
	$(INSTALL) -m 0755 "bin/$(APP_NAME)" "$(BINDIR)/$(APP_NAME)"
	@printf "installed %s\nensure %s is on your PATH\n" "$(BINDIR)/$(APP_NAME)" "$(BINDIR)"

uninstall:
	rm -f "$(BINDIR)/$(APP_NAME)"
	@printf "removed %s (if it existed)\n" "$(BINDIR)/$(APP_NAME)"

clean:
	rm -rf bin
