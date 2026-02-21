GO ?= go
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
DESTDIR ?=
INSTALL ?= install
CGO_ENABLED ?= 0
STAMP_DIR := .make
RUN_FUZZ ?= 0
FUZZTIME ?= 1s
RUN_BENCH ?= 0
BENCHTIME ?= 1x
RUN_BENCHMETA ?= 0

-include Makefile.local

BINARY ?= $(notdir $(abspath $(CURDIR)))
MODULE_DIR := $(notdir $(abspath $(CURDIR)))
DOCKER ?= docker
DOCKER_TEST_IMAGE ?= $(MODULE_DIR)-test
ENABLE_WASM ?= $(if $(wildcard cmd/$(BINARY)-wasm/main.go),1,0)
WASM_BUILD_PKG ?= ./cmd/$(BINARY)-wasm
WASM_OUT ?= internal/ui/static/assets/app.wasm
WASM_RUNTIME_DST ?= internal/ui/static/assets/wasm_exec.js
CMD_PKG := ./cmd/$(BINARY)
BENCHMETA_CMD_PKG := ./cmd/$(BINARY)-benchmeta
BENCHMETA_MAIN := cmd/$(BINARY)-benchmeta/main.go
GO_FILES := $(shell find . -type f -name '*.go' -not -path './bin/*' -not -path './.make/*' | sort)
TEST_FILES := $(shell find tests -type f | sort)
GO_DEPS := go.mod $(wildcard go.sum)
GOROOT := $(shell $(GO) env GOROOT)
WASM_EXEC_JS := $(firstword $(wildcard $(GOROOT)/lib/wasm/wasm_exec.js $(GOROOT)/misc/wasm/wasm_exec.js))
ifeq ($(ENABLE_WASM),1)
TEST_PKGS ?= $(shell CGO_ENABLED=$(CGO_ENABLED) $(GO) list ./... 2>/dev/null | grep -v '/internal/ui/wasm$$' | grep -v '/cmd/$(BINARY)-wasm$$')
FUZZ_PKGS ?= $(TEST_PKGS)
else
TEST_PKGS ?= ./...
FUZZ_PKGS ?= $(shell CGO_ENABLED=$(CGO_ENABLED) $(GO) list ./...)
endif

WASM_STAMP := $(STAMP_DIR)/wasm.stamp
FMT_STAMP := $(STAMP_DIR)/fmt.stamp
LINT_STAMP := $(STAMP_DIR)/lint.stamp
TEST_STAMP := $(STAMP_DIR)/test.stamp
FUZZ_STAMP := $(STAMP_DIR)/fuzz.stamp
BENCH_STAMP := $(STAMP_DIR)/bench.stamp
E2E_STAMP := $(STAMP_DIR)/e2e.stamp

.PHONY: all build build-wasm test test-fuzz test-bench bench test-docker test-e2e e2e fmt lint check benchmeta install uninstall clean

all: build

build-wasm: $(WASM_STAMP)

ifeq ($(ENABLE_WASM),1)
$(WASM_STAMP): $(GO_FILES) $(GO_DEPS)
	mkdir -p $(STAMP_DIR)
	test -n "$(WASM_BUILD_PKG)" || (echo "ENABLE_WASM=1 requires WASM_BUILD_PKG" >&2; exit 1)
	test -n "$(WASM_OUT)" || (echo "ENABLE_WASM=1 requires WASM_OUT" >&2; exit 1)
	test -n "$(WASM_RUNTIME_DST)" || (echo "ENABLE_WASM=1 requires WASM_RUNTIME_DST" >&2; exit 1)
	test -n "$(WASM_EXEC_JS)" || (echo "wasm_exec.js not found under $(GOROOT)/lib/wasm or $(GOROOT)/misc/wasm" >&2; exit 1)
	mkdir -p "$(dir $(WASM_OUT))"
	mkdir -p "$(dir $(WASM_RUNTIME_DST))"
	cp "$(WASM_EXEC_JS)" "$(WASM_RUNTIME_DST)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=js GOARCH=wasm $(GO) build -o "$(WASM_OUT)" "$(WASM_BUILD_PKG)"
	touch $(WASM_STAMP)
else
WASM_STAMP :=
endif

build: ./bin/$(BINARY) $(WASM_STAMP)

./bin/$(BINARY): $(GO_FILES) $(GO_DEPS) $(WASM_STAMP)
	mkdir -p ./bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o ./bin/$(BINARY) $(CMD_PKG)

fmt: $(FMT_STAMP)

$(FMT_STAMP): $(GO_FILES)
	mkdir -p $(STAMP_DIR)
	gofmt -w .
	touch $(FMT_STAMP)

lint: $(LINT_STAMP)

$(LINT_STAMP): $(GO_FILES) $(GO_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) vet $(TEST_PKGS)
	touch $(LINT_STAMP)

test: $(TEST_STAMP)

$(TEST_STAMP): $(GO_FILES) $(GO_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_PKGS)
	touch $(TEST_STAMP)

test-fuzz: $(FUZZ_STAMP)

$(FUZZ_STAMP): $(GO_FILES) $(GO_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	@set -eu; \
	for pkg in $(FUZZ_PKGS); do \
		fuzzes=$$(CGO_ENABLED=$(CGO_ENABLED) $(GO) test "$$pkg" -list Fuzz | awk '/^Fuzz/ {print}'); \
		if [ -n "$$fuzzes" ]; then \
			for fuzz in $$fuzzes; do \
				CGO_ENABLED=$(CGO_ENABLED) $(GO) test "$$pkg" -run="^$$fuzz$$" -fuzz="^$$fuzz$$" -fuzztime=$(FUZZTIME); \
			done; \
		fi; \
	done
	touch $(FUZZ_STAMP)

test-bench: $(BENCH_STAMP)

$(BENCH_STAMP): $(GO_FILES) $(GO_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test -run=^$$ -bench=. -benchmem -benchtime=$(BENCHTIME) $(TEST_PKGS)
	touch $(BENCH_STAMP)

bench: test-bench

test-docker:
	$(DOCKER) build -t $(DOCKER_TEST_IMAGE) -f Dockerfile .
	$(DOCKER) run --rm -v "$(CURDIR)/..:/workspace" -w "/workspace/$(MODULE_DIR)" $(DOCKER_TEST_IMAGE) make test

test-e2e: $(E2E_STAMP)

$(E2E_STAMP): ./bin/$(BINARY) $(TEST_FILES)
	mkdir -p $(STAMP_DIR)
	bash ./tests/e2e.sh
	touch $(E2E_STAMP)

e2e: test-e2e

benchmeta:
ifeq ($(wildcard $(BENCHMETA_MAIN)),)
	@echo "benchmeta: no metadata runner for $(BINARY) (expected $(BENCHMETA_MAIN)); skipping"
else
	CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(BENCHMETA_CMD_PKG) --format text
	CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(BENCHMETA_CMD_PKG) --format json
endif

check: fmt lint test test-e2e
ifeq ($(RUN_FUZZ),1)
check: test-fuzz
endif
ifeq ($(RUN_BENCH),1)
check: test-bench
endif
ifeq ($(RUN_BENCHMETA),1)
check: benchmeta
endif

install: build
	mkdir -p "$(DESTDIR)$(BINDIR)"
	$(INSTALL) -m 0755 ./bin/$(BINARY) "$(DESTDIR)$(BINDIR)/$(BINARY)"
	@echo "Installed $(DESTDIR)$(BINDIR)/$(BINARY). Ensure $(BINDIR) is on PATH."

uninstall:
	rm -f "$(DESTDIR)$(BINDIR)/$(BINARY)"
	@echo "Removed $(DESTDIR)$(BINDIR)/$(BINARY) if it existed."

clean:
	rm -rf ./bin
	rm -rf $(STAMP_DIR)
