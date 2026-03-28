GO ?= go
GOFLAGS ?= -mod=readonly
BUILD_TRIMPATH ?= 1
BUILD_VCS ?= 0
BUILD_LDFLAGS ?= -s -w
BUILD_STATIC ?= 1
BUILD_TAGS ?= netgo,osusergo
TEST_TAGS ?= $(BUILD_TAGS)
TEST_VERBOSE ?= 1
DEBUG_GCFLAGS ?= all=-N -l -m
GRC ?= grc
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
VSCODE_EXTENSION_DIR ?= editors/vscode-bus-language
VSIX_OUT ?=
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
GO_DEPS := Makefile Makefile.local go.mod $(wildcard go.sum)
MODULE_BIN_DEPS ?=
MODULE_BIN_DEPS_PREFIX ?= ../
MODULE_BIN_DEP_PATHS := $(foreach mod,$(MODULE_BIN_DEPS),$(MODULE_BIN_DEPS_PREFIX)$(mod)/bin/$(mod))
MODULE_BIN_EXISTING_DEP_PATHS := $(foreach p,$(MODULE_BIN_DEP_PATHS),$(wildcard $(p)))
MODULE_SRC_DEPS ?=
MODULE_SRC_DEPS_PREFIX ?= ../
MODULE_SRC_DEP_DIRS := $(foreach mod,$(MODULE_SRC_DEPS),$(MODULE_SRC_DEPS_PREFIX)$(mod))
MODULE_SRC_DEP_GO_FILES := $(shell set -eu; for d in $(MODULE_SRC_DEP_DIRS); do [ -d "$$d" ] || continue; find "$$d" -type f -name '*.go' -not -path "$$d/bin/*" -not -path "$$d/.make/*"; done | sort)
MODULE_SRC_DEP_MOD_FILES := $(foreach mod,$(MODULE_SRC_DEPS),$(wildcard $(MODULE_SRC_DEPS_PREFIX)$(mod)/go.mod) $(wildcard $(MODULE_SRC_DEPS_PREFIX)$(mod)/go.sum))
BUILD_SRC_DEPS := $(MODULE_SRC_DEP_GO_FILES) $(MODULE_SRC_DEP_MOD_FILES)
GOROOT := $(shell $(GO) env GOROOT)
WASM_EXEC_JS := $(firstword $(wildcard $(GOROOT)/lib/wasm/wasm_exec.js $(GOROOT)/misc/wasm/wasm_exec.js))
ifeq ($(ENABLE_WASM),1)
TEST_PKGS ?= $(shell CGO_ENABLED=$(CGO_ENABLED) $(GO) list ./... 2>/dev/null | grep -v '/internal/ui/wasm$$' | grep -v '/cmd/$(BINARY)-wasm$$')
FUZZ_PKGS ?= $(TEST_PKGS)
else
TEST_PKGS ?= ./...
FUZZ_PKGS ?= $(shell CGO_ENABLED=$(CGO_ENABLED) $(GO) list ./...)
endif

ifneq ($(BUILD_TRIMPATH),0)
BUILD_TRIMPATH_ARG := -trimpath
else
BUILD_TRIMPATH_ARG :=
endif
ifneq ($(BUILD_VCS),0)
BUILD_VCS_ARG := -buildvcs=true
else
BUILD_VCS_ARG := -buildvcs=false
endif
ifneq ($(BUILD_STATIC),0)
BUILD_STATIC_LDFLAGS := -extldflags "-static"
else
BUILD_STATIC_LDFLAGS :=
endif
BUILD_LDFLAGS_COMBINED := $(strip $(BUILD_LDFLAGS) $(BUILD_STATIC_LDFLAGS))
ifneq ($(strip $(BUILD_LDFLAGS_COMBINED)),)
BUILD_LDFLAGS_ARG := -ldflags '$(BUILD_LDFLAGS_COMBINED)'
else
BUILD_LDFLAGS_ARG :=
endif
ifneq ($(strip $(BUILD_TAGS)),)
BUILD_TAGS_ARG := -tags '$(BUILD_TAGS)'
else
BUILD_TAGS_ARG :=
endif
ifneq ($(strip $(TEST_TAGS)),)
TEST_TAGS_ARG := -tags '$(TEST_TAGS)'
else
TEST_TAGS_ARG :=
endif
ifneq ($(TEST_VERBOSE),0)
TEST_VERBOSE_ARG := -v
else
TEST_VERBOSE_ARG :=
endif

WASM_STAMP := $(STAMP_DIR)/wasm.stamp
FMT_STAMP := $(STAMP_DIR)/fmt.stamp
LINT_STAMP := $(STAMP_DIR)/lint.stamp
TEST_STAMP := $(STAMP_DIR)/test.stamp
FUZZ_STAMP := $(STAMP_DIR)/fuzz.stamp
BENCH_STAMP := $(STAMP_DIR)/bench.stamp
E2E_STAMP := $(STAMP_DIR)/e2e.stamp

.PHONY: all tidy build build-debug build-wasm test color-test test-fuzz test-bench color-bench bench test-docker test-e2e e2e fmt lint check benchmeta package-vscode-extension check-vscode-extension-release check-tree-sitter-bus-language check-bus-language-server install uninstall clean

all: build

tidy:
	$(GO) mod tidy

build-wasm: $(WASM_STAMP)

ifeq ($(ENABLE_WASM),1)
$(WASM_STAMP): $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS)
	mkdir -p $(STAMP_DIR)
	test -n "$(WASM_BUILD_PKG)" || (echo "ENABLE_WASM=1 requires WASM_BUILD_PKG" >&2; exit 1)
	test -n "$(WASM_OUT)" || (echo "ENABLE_WASM=1 requires WASM_OUT" >&2; exit 1)
	test -n "$(WASM_RUNTIME_DST)" || (echo "ENABLE_WASM=1 requires WASM_RUNTIME_DST" >&2; exit 1)
	test -n "$(WASM_EXEC_JS)" || (echo "wasm_exec.js not found under $(GOROOT)/lib/wasm or $(GOROOT)/misc/wasm" >&2; exit 1)
	mkdir -p "$(dir $(WASM_OUT))"
	mkdir -p "$(dir $(WASM_RUNTIME_DST))"
	cp "$(WASM_EXEC_JS)" "$(WASM_RUNTIME_DST)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=js GOARCH=wasm $(GO) build $(GOFLAGS) $(BUILD_TRIMPATH_ARG) $(BUILD_VCS_ARG) $(BUILD_LDFLAGS_ARG) $(BUILD_TAGS_ARG) -o "$(WASM_OUT)" "$(WASM_BUILD_PKG)"
	touch $(WASM_STAMP)
else
WASM_STAMP :=
endif

build: ./bin/$(BINARY) $(WASM_STAMP)

./bin/$(BINARY): $(GO_FILES) $(GO_DEPS) $(MODULE_BIN_EXISTING_DEP_PATHS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p ./bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(BUILD_TRIMPATH_ARG) $(BUILD_VCS_ARG) $(BUILD_LDFLAGS_ARG) $(BUILD_TAGS_ARG) -o ./bin/$(BINARY) $(CMD_PKG)

build-debug: $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p ./bin
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) $(BUILD_VCS_ARG) $(BUILD_TAGS_ARG) -gcflags '$(DEBUG_GCFLAGS)' -o ./bin/$(BINARY) $(CMD_PKG)

fmt: $(FMT_STAMP)

$(FMT_STAMP): $(GO_FILES)
	mkdir -p $(STAMP_DIR)
	gofmt -w .
	touch $(FMT_STAMP)

lint: $(LINT_STAMP)

$(LINT_STAMP): $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) vet $(TEST_PKGS)
	touch $(LINT_STAMP)

test: $(TEST_STAMP)

$(TEST_STAMP): $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_TAGS_ARG) $(TEST_VERBOSE_ARG) $(TEST_PKGS)
	touch $(TEST_STAMP)

color-test:
	CGO_ENABLED=$(CGO_ENABLED) $(GRC) $(GO) test $(TEST_TAGS_ARG) $(TEST_VERBOSE_ARG) $(TEST_PKGS)

test-fuzz: $(FUZZ_STAMP)

$(FUZZ_STAMP): $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	@set -eu; \
	for pkg in $(FUZZ_PKGS); do \
		fuzzes=$$(CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_TAGS_ARG) "$$pkg" -list Fuzz | awk '/^Fuzz/ {print}'); \
		if [ -n "$$fuzzes" ]; then \
			for fuzz in $$fuzzes; do \
				CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_TAGS_ARG) "$$pkg" -run="^$$fuzz$$" -fuzz="^$$fuzz$$" -fuzztime=$(FUZZTIME); \
			done; \
		fi; \
	done
	touch $(FUZZ_STAMP)

test-bench: $(BENCH_STAMP)

$(BENCH_STAMP): $(GO_FILES) $(GO_DEPS) $(BUILD_SRC_DEPS) $(WASM_STAMP)
	mkdir -p $(STAMP_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(TEST_TAGS_ARG) $(TEST_VERBOSE_ARG) -run=^$$ -bench=. -benchmem -benchtime=$(BENCHTIME) $(TEST_PKGS)
	touch $(BENCH_STAMP)

color-bench:
	CGO_ENABLED=$(CGO_ENABLED) $(GRC) $(GO) test $(TEST_TAGS_ARG) $(TEST_VERBOSE_ARG) -run=^$$ -bench=. -benchmem -benchtime=$(BENCHTIME) $(TEST_PKGS)

bench: test-bench

test-docker:
	$(DOCKER) build -t $(DOCKER_TEST_IMAGE) -f Dockerfile .
	$(DOCKER) run --rm -v "$(CURDIR)/..:/workspace" -w "/workspace/$(MODULE_DIR)" $(DOCKER_TEST_IMAGE) make test

test-e2e: $(E2E_STAMP)

$(E2E_STAMP): ./bin/$(BINARY) $(TEST_FILES)
	mkdir -p $(STAMP_DIR)
	@if [ "$${BUS_E2E_VERBOSE:-0}" = "1" ]; then \
		bash ./tests/e2e.sh; \
	else \
		log=$$(mktemp); \
		if bash ./tests/e2e.sh >"$$log" 2>&1; then \
			passed=$$(grep -Ec '^PASS([ :]|$$)' "$$log" || true); \
			skipped=$$(grep -Ec '^SKIP([ :]|$$)' "$$log" || true); \
			if [ "$$passed" -eq 0 ] && [ "$$skipped" -eq 0 ]; then passed=1; fi; \
			if [ "$$skipped" -gt 0 ]; then grep -E '^SKIP([ :]|$$)' "$$log"; fi; \
			printf "e2e OK (%s: passed %s, skipped %s)\n" "$(BINARY)" "$$passed" "$$skipped"; \
		else \
			passed=$$(grep -Ec '^PASS([ :]|$$)' "$$log" || true); \
			skipped=$$(grep -Ec '^SKIP([ :]|$$)' "$$log" || true); \
			failed=$$(grep -Ec '^FAIL([ :]|$$)' "$$log" || true); \
			if [ "$$failed" -eq 0 ]; then failed=1; fi; \
			printf "e2e FAILED (%s: passed %s, skipped %s, failed %s)\n" "$(BINARY)" "$$passed" "$$skipped" "$$failed"; \
			cat "$$log"; \
			rm -f "$$log"; \
			exit 1; \
		fi; \
		rm -f "$$log"; \
	fi
	touch $(E2E_STAMP)

e2e: test-e2e

benchmeta:
ifeq ($(wildcard $(BENCHMETA_MAIN)),)
	@echo "benchmeta: no metadata runner for $(BINARY) (expected $(BENCHMETA_MAIN)); skipping"
else
	CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(BENCHMETA_CMD_PKG) --format text
	CGO_ENABLED=$(CGO_ENABLED) $(GO) run $(BENCHMETA_CMD_PKG) --format json
endif

package-vscode-extension: $(shell find $(VSCODE_EXTENSION_DIR) -type f | sort) scripts/package_vscode_bus_language.py
	@if [ -n "$(VSIX_OUT)" ]; then \
		python3 ./scripts/package_vscode_bus_language.py --output "$(VSIX_OUT)"; \
	else \
		python3 ./scripts/package_vscode_bus_language.py; \
	fi

check-vscode-extension-release: $(shell find $(VSCODE_EXTENSION_DIR) -type f | sort) scripts/check_vscode_bus_language_release.py
	python3 ./scripts/check_vscode_bus_language_release.py

check-tree-sitter-bus-language: $(shell find editors/tree-sitter-bus -type f | sort) scripts/check_tree_sitter_bus_language.js
	node ./scripts/check_tree_sitter_bus_language.js

check-bus-language-server: $(shell find $(VSCODE_EXTENSION_DIR) -type f | sort) scripts/check_bus_language_server.py
	python3 ./scripts/check_bus_language_server.py

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
