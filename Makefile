APP_NAME := bus

.PHONY: all build test fmt check clean

all: build

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

check: fmt test

clean:
	rm -rf bin
