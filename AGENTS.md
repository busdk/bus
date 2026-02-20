# AGENTS.md — bus dispatcher

Merged guidance from `.cursor/rules/*.mdc`.

Agent-facing instructions for the `bus` core dispatcher repository. This module is the top-level CLI dispatcher and must remain minimal. Canonical design is the BusDK spec; this file consolidates local, actionable guidance.

## Project overview

**Purpose:** Provide a deterministic dispatcher that executes `bus-<command>` from PATH. The `bus` binary is the primary entry point for discovery and execution (BusDK SDD interface IF-001).

**Non-goals:** No business logic, no dataset/schema I/O, no Git operations, no network, no module-specific flags beyond shared BusDK CLI conventions. The dispatcher does not implement `--dry-run` or workspace/Git checks; subcommands handle those per the CLI spec.

**Canonical sources (BusDK spec):**

- [BusDK Design Document](https://docs.busdk.com/)
- [BusDK Software Design Document (SDD)](https://docs.busdk.com/sdd) — goals, requirements, IF-001 (dispatcher routing)
- [Command structure and discoverability](https://docs.busdk.com/cli/command-structure)
- [Independent modules](https://docs.busdk.com/architecture/independent-modules)
- [Error handling, dry-run, and diagnostics](https://docs.busdk.com/cli/error-handling-dry-run-diagnostics)
- [CLI tooling and workflow](https://docs.busdk.com/cli/index)
- [Data directory layout](https://docs.busdk.com/layout/index), [layout principles](https://docs.busdk.com/layout/layout-principles)
- [Data formats](https://docs.busdk.com/data/index): [CSV conventions](https://docs.busdk.com/data/csv-conventions), [Table Schema contract](https://docs.busdk.com/data/table-schema-contract)

**Spec vs repo naming:** The CLI spec examples use `busdk ...`; this repo implements `bus` as the dispatcher. Use binary name `bus` and subcommand prefix `bus-` everywhere in this repository.

## Invocation and behavior

- **Pattern:** `bus <command> [args...]`. This module is the dispatcher; it does not implement domain subcommands.
- **Dispatch:** Resolve the executable with `exec.LookPath("bus-" + name)` only, then exec with args unchanged, inheriting stdin, stdout, stderr, and environment.
- **Busfile dispatch selection:** in `.bus` execution mode, prefer in-process module runners when available; otherwise use `bus-<target>` shell lookup only when `bus.busfile.dispatch.shell_lookup_enabled=true`.
- **FS transactions:** `provider=fs` is valid only when all busfile targets have in-process transaction-capable runners (Tx runners); otherwise fallback/error rules apply.
- **No arguments:** Print a short usage line to stderr, exit 2, and include an available-commands list if any are discovered.
- **Missing subcommand (not found in PATH):** stderr must start with `bus:` and mention the expected `bus-<name>` in PATH, then print usage and available commands; exit 127.
- **Subcommand exit codes:** Pass through exactly. Unexpected exec failures return 1 with a short `bus:` error on stderr.

## Subcommand discoverability (listing only)

- Scan PATH left-to-right. For each directory, list entries non-recursively.
- Collect executables matching `bus-*` (Windows: `bus-*.exe`); ignore directories.
- Skip missing or inaccessible PATH entries (e.g. EACCES/EPERM) silently.
- Deduplicate by subcommand name, preferring the earliest PATH entry.
- Sort subcommand names lexicographically before printing.

## Inputs, outputs, and side effects

- **Inputs:** CLI args, PATH, environment, stdin.
- **Outputs:** Only diagnostics/usage on stderr and subcommand output pass-through on stdout/stderr.
- **Side effects:** Process execution only; no file writes or dataset mutations. Keep error messages concise and script-friendly (see [Error handling, dry-run, and diagnostics](https://docs.busdk.com/cli/error-handling-dry-run-diagnostics)).

## Build and test commands

- **Build:** `make build` or `go build -o bin/bus ./cmd/bus`
- **Tests:** `make test` or `go test ./...`
- **Format:** `make fmt` or `gofmt -w .`
- **Lint:** `make lint` or `go vet ./...`
- **All checks:** `make check` (fmt, lint, test)
- **Install:** `make install` (builds then installs to `$(BINDIR)`, default `$(PREFIX)/bin` with `PREFIX ?= $(HOME)/.local`)

## Testing instructions

- Tests must be hermetic and deterministic (no network).
- Use temporary PATH entries and compiled fake subcommands (`go build` into a temp dir) to validate dispatch behavior and listing determinism.
- Run the full suite with `go test ./...` before considering the change done.

## Code style and conventions

- Use `gofmt` for formatting; run `go vet ./...` as part of check.
- Keep the dispatcher minimal: no business logic, no dataset I/O, no Git or network usage.
- Implement behavior that matches the module contract above and the BusDK SDD (IF-001) now; do not add transitional or deferred behavior without a spec-backed requirement.

## Quality gates

- Build must succeed; `go test ./...` must pass.
- Coverage of dispatch, no-args, missing-subcommand, and discoverability behavior is required.
- Static analysis and vet must report no new findings.

## Document references used

This AGENTS.md was grounded in the following BusDK spec pages:

- [docs.busdk.com](https://docs.busdk.com/) — design spec entrypoint
- [docs.busdk.com/sdd](https://docs.busdk.com/sdd) — SDD (IF-001 bus dispatcher, goals, NFRs)
- [docs.busdk.com/cli/command-structure](https://docs.busdk.com/cli/command-structure) — command layout and discoverability
- [docs.busdk.com/architecture/independent-modules](https://docs.busdk.com/architecture/independent-modules) — module boundaries, no CLI-to-CLI as API
- [docs.busdk.com/cli/error-handling-dry-run-diagnostics](https://docs.busdk.com/cli/error-handling-dry-run-diagnostics) — exit codes, stderr diagnostics, script-friendly behavior
- [docs.busdk.com/sdd/modules](https://docs.busdk.com/sdd/modules) — module index (dispatcher described in main SDD, not a separate module SDD)

## Gitignore Rule

1. .bus MUST be tracked; never add .bus or .bus/ to .gitignore.
2. Runtime lock artifacts such as .bus-dev.lock may be ignored.
