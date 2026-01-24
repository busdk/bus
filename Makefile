APP_NAME := bus

.PHONY: all build test fmt clean

all: build

build:
	go build -o bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin
