# AGENTS.md — bus dispatcher

Merged guidance from `.cursor/rules/*.mdc`.

Agent-facing instructions for the `bus` core dispatcher repository. This module is the top-level CLI dispatcher and must remain minimal. Canonical design is the BusDK spec; this file consolidates local, actionable guidance.

## Project overview

**Purpose:** Provide a deterministic dispatcher that executes `bus-<command>` from PATH. The `bus` binary is the primary entry point for discovery and execution (BusDK SDD interface IF-001).

**Non-goals:** No business logic, no dataset/schema I/O, no Git operations, no network, no module-specific flags beyond shared BusDK CLI conventions. The dispatcher does not implement `--dry-run` or workspace/Git checks; subcommands handle those per the CLI spec.

**Update-check policy:** The public `bus` dispatcher intentionally does **not** integrate `github.com/busdk/bus-update/pkg/updatecheck` startup enforcement. Update-check behavior belongs to `bus-*` module binaries, not the public dispatcher.

**Canonical sources (BusDK spec):**

- [BusDK Design Document](https://docs.busdk.com/)
- [BusDK Software Design Document (SDD)](sdd/docs/sdd.md) — goals, requirements, IF-001 (dispatcher routing)
- [Command structure and discoverability](https://docs.busdk.com/cli/command-structure)
- [Independent modules](https://docs.busdk.com/architecture/independent-modules)
- [Error handling, dry-run, and diagnostics](https://docs.busdk.com/cli/error-handling-dry-run-diagnostics)
- [CLI tooling and workflow](https://docs.busdk.com/cli/index)
- [Data directory layout](https://docs.busdk.com/layout/index), [layout principles](https://docs.busdk.com/layout/layout-principles)
- [Data formats](https://docs.busdk.com/data/index): [CSV conventions](https://docs.busdk.com/data/csv-conventions), [Table Schema contract](https://docs.busdk.com/data/table-schema-contract)

**Spec vs repo naming:** The CLI spec examples use `busdk ...`; this repo implements `bus` as the dispatcher. Use binary name `bus` and subcommand prefix `bus-` everywhere in this repository.

**Visibility boundary:** `bus` is public/open-source. Treat `bus-*` modules as separate private repositories unless explicitly documented otherwise; do not add in-process dependencies from `bus` into private module internals.

## Invocation and behavior

- **Pattern:** `bus <command> [args...]`. This module is the dispatcher; it does not implement domain subcommands.
- **Dispatch:** Resolve the executable with `exec.LookPath("bus-" + name)` only, then exec with args unchanged, inheriting stdin, stdout, stderr, and environment.
- **Busfile dispatch selection:** in `.bus` execution mode, only use in-process runners that are explicitly registered inside this open-source `bus` module; otherwise use `bus-<target>` shell lookup only when `bus.busfile.dispatch.shell_lookup_enabled=true`.
- **FS transactions:** `provider=fs` is valid only when all busfile targets have in-process transaction-capable runners (Tx runners); otherwise fallback/error rules apply.
- **No private module wiring:** do not add default in-process or in-process-tx runners for private/other-module commands (for example `bank` or `journal`) in this repository.
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

## Performance guardrails (Go optimization guide)

- Optimization guide reference: online at `https://docs.busdk.com/implementation/go-optimization-guide` and local mirror typically at `../docs/docs/implementation/go-optimization-guide.md`.
- Standing guidance: when review/benchmark work reveals a reusable optimization anti-pattern or workflow improvement, update the local guide at `../docs/docs/implementation/go-optimization-guide.md` in the same change (not just repo-local TODOs).
- Profile/benchmark first: for performance changes, capture a baseline benchmark (or pprof where appropriate) before changing code, then compare after.
- Avoid repeated env-slice scans/copies in per-command loops. If command env keys are constant, build once and mutate minimally instead of re-running full `upsert` passes.
- In hot loops, preallocate maps/slices with known bounds (`make(..., len(...))`) and avoid repeated reallocation churn.
- Prefer algorithmic improvements over micro-tuning. Do not ship O(commands*env), O(files*workspace), or O(lookups*deletes) patterns on critical execution paths when indexed/incremental alternatives exist.
- For `.bus` `provider=fs` paths, avoid full workspace copy + full-tree diff per command unless there is a measured, documented reason. Prefer change-scoped/incremental capture.
- Avoid linear tombstone/prefix scans for each path lookup in TxFS when delete counts can grow; use path-indexed lookup structures.
- Avoid recomputing deterministic derived values (for example sorted unique command-file lists) multiple times in the same execution unit.
- Keep file operations streaming on large files (`io.Copy`, readers/writers) and avoid read-all patterns in commit/copy paths.
- Use benchmark tests for hotspot-sensitive helpers and loops (`go test -run '^$' -bench ... -benchmem`), and include them with performance-oriented refactors.

## Quality gates

- Build must succeed; `go test ./...` must pass.
- Coverage of dispatch, no-args, missing-subcommand, and discoverability behavior is required.
- Static analysis and vet must report no new findings.

## Document references used

This AGENTS.md was grounded in the following BusDK spec pages:

- [docs.busdk.com](https://docs.busdk.com/) — design spec entrypoint
- [docs.busdk.com/sdd](sdd/docs/sdd.md) — SDD (IF-001 bus dispatcher, goals, NFRs)
- [docs.busdk.com/cli/command-structure](https://docs.busdk.com/cli/command-structure) — command layout and discoverability
- [docs.busdk.com/architecture/independent-modules](https://docs.busdk.com/architecture/independent-modules) — module boundaries, no CLI-to-CLI as API
- [docs.busdk.com/cli/error-handling-dry-run-diagnostics](https://docs.busdk.com/cli/error-handling-dry-run-diagnostics) — exit codes, stderr diagnostics, script-friendly behavior
- [sdd/docs/modules/modules.md](sdd/docs/modules/modules.md) — module index (dispatcher described in main SDD, not a separate module SDD)

## Gitignore Rule

1. .bus MUST be tracked; never add .bus or .bus/ to .gitignore.
2. In private repositories, .bus/ must be tracked; .bus/secrets may be tracked in private repositories only and must not be tracked otherwise.
3. Runtime lock artifacts such as .bus-dev.lock may be ignored.

## Session carry-forward notes

- When starting from this repo alone, editing `../docs/docs/implementation/go-optimization-guide.md` may require higher-level workspace access; from the super-project root this should be editable directly.
- Optimization-guide updates must be additive: do not remove prior guide content when adding new patterns.
- Performance findings already benchmarked in this repo and worth upstreaming to the optimization guide:
  - repeated env rewrites in per-command loops (`withBusfileEnv`/`upsertEnv`) show high allocation churn (`internal/dispatch/run_bench_test.go`)
  - repeated PATH resolution of the same target in batch preflight/dispatch is expensive (`BenchmarkPreflightDispatchTargetsRepeatedLookups` in `internal/dispatch/run_bench_test.go`)
  - tombstone lookup scaling is linear in delete count (`BenchmarkIsDeleted` in `internal/txfs/txfs_bench_test.go`)
  - delete bookkeeping currently scans full `changes` map (`BenchmarkMarkDelete` in `internal/txfs/txfs_bench_test.go`)

## Shared Superproject Conventions

- Prefer minimal, deterministic, script-friendly behavior.
- Deletion safety: tracked paths use `git rm` (or `git rm --cached`), untracked paths use `rm`.
- When a system-level CLI command fails due to incorrect parameters, record the correct invocation in the most relevant `AGENTS.md`.
- On macOS/BSD `cat`, `-A` is unsupported; use `cat -vet` or `sed -n 'l'` to visualize tabs and line endings instead.
- On macOS/BSD `awk`, avoid using `in` as a variable name (`in` is reserved in `for (x in y)`); use names like `inside` instead.
- When running shell commands that contain backticks in regex/pattern arguments (for example with `rg`), wrap the full command in single quotes or escape backticks to avoid command-substitution parse errors.
- `rg` does not support look-around by default; use `rg --pcre2` when patterns require look-ahead/look-behind.
- Use `python3` (not `python`) for Python scripting in this environment.

## Global unit documentation traceability rule

- Every top-level production-code unit (`func`, `type`, `var`, and `const` blocks when they define global API/behavior) must include an inline comment that states its purpose.
- For each top-level global unit, also include concise `Used by:` traceability in the inline comment (or immediately adjacent comment) that names the primary caller(s), owning flow, or integration point.
- Keep `Used by:` comments accurate when refactoring: update or remove stale references in the same change set.
- Do not add new undocumented top-level global units.
