# bus plan

## Active feature plan

- [x] Add dispatcher-facing bootstrap/package-manager UX coverage for BusDK installer flows end-to-end: define and document the supported user-facing `bus ...` command surface for bootstrap-installed package management while keeping `bus` a pure dispatcher to `bus-update`, keep missing-subcommand/help/exit-code contracts deterministic when the package-manager module is present or absent on both Windows (`.exe`) and non-Windows platforms, update README and docs/SDD wording in the same change, add unit tests for discoverability and diagnostics around the delegated install/update command family, and add e2e coverage with fake `bus-update` binaries that proves argument/exit-code pass-through without embedding package-manager logic in `bus`.

- [x] Tighten the shipped `.bus` TextMate grammar end-to-end: extend the canonical VS Code-compatible grammar so it highlights sticky directive lines, `key=value` assignments, and trailing line-continuation backslashes in addition to the existing shebang/comment/include/command/flag/string/date coverage, add source-grammar and packaged-`.vsix` validation tests for representative `.bus` fixtures, and update README plus editor-support docs/SDD wording in the same change.

- [x] Add maintainer-ready distribution coverage for the shipped VS Code-compatible `.bus` extension end-to-end: document and automate the supported release surfaces (`.vsix` release asset first, Open VSX next), keep install docs aligned for VS Code/Cursor/VSCodium-style editors, and add verification that published package metadata remains compatible with those distribution paths.

- [x] Add a parser-backed premium highlighting layer for `.bus` end-to-end with Tree-sitter as the next editor-support tier: define a Tree-sitter grammar and highlight queries from the same BusDK busfile syntax contract, document the supported Neovim/Emacs integration path, and land parser/query tests plus docs in the same change.

- [x] Add semantic-token/LSP support for `.bus` end-to-end after the parser-backed layer stabilizes: expose standard semantic token classes for commands, flags, assignments, strings, dates, and numbers, document editor integration expectations, and land deterministic tests and docs in the same change.

- [x] Add distributable `.bus` language tooling for the `bus` source format end-to-end: define a canonical syntax-highlighting grammar for `.bus` shebang/comments/commands/flags/strings, package and document a VS Code/Cursor-compatible extension plus a downloadable `.vsix` install path for users, add repository/browser fallback classification where appropriate, and land implementation docs, user install docs, and automated verification for the shipped grammar/package metadata in the same change.

- [x] FR57/Task 1: Define and document dispatcher command boundaries for identity and plan features.
  - Implementation: add dispatcher-level command discoverability/usage coverage for `auth` and `plan` command families without adding business logic to `bus`.
  - Documentation: update `README.md` and docs module pages to describe `bus auth ...` and `bus plan ...` delegation model and module split rationale.
  - Unit tests: add/adjust dispatcher tests for missing subcommand diagnostics and deterministic listing when `bus-auth` and `bus-plan` are present/absent.
  - E2E tests: add dispatcher e2e scenarios validating pass-through args/exit codes for `bus auth` and `bus plan`.

- [x] FR57/Task 2: Define stable global UX contract for auth/plan CLIs at the dispatcher boundary.
  - Implementation: specify command synopsis and shared error-envelope expectations at dispatcher-facing integration points only (no auth/plan backend logic in `bus`; CLI modules remain thin clients).
  - Documentation: add concise command contract table in docs (`modules/bus.md` + `sdd/bus.md`) for discoverability and compatibility notes.
  - Unit tests: verify usage/help output remains deterministic with the expanded command surface.
  - E2E tests: validate end-to-end dispatcher behavior with `--help`, unknown subcommands, and exit-code pass-through for new module commands.

## E2E coverage gaps

- [x] E2E Gap: command discoverability determinism with PATH shadowing and inaccessible entries (`internal/dispatch/run.go` `listSubcommands`).
  - Affected flow: `bus` no-args and `bus --help` when `PATH` contains duplicate `bus-*` commands across directories and at least one unreadable PATH directory.
  - Expected e2e behavior: output includes each command name once, resolved from the earliest PATH occurrence, and silently skips unreadable PATH entries.
  - Target test scope: add assertions in `tests/e2e/010-global-flags-help-missing.sh` (or a focused follow-up e2e file) for duplicate-command shadowing and unreadable PATH directory handling.

- [x] E2E Gap: direct path-qualified subcommand execution (`internal/dispatch/run.go` `lookPathEnv` + `Run`).
  - Affected flow: dispatch targets provided with path separators (for example `bus ./bin/bus-accounts`, `bus /tmp/bus-status`) and forwarded global flags.
  - Expected e2e behavior: PATH lookup is bypassed, the given executable is run directly, args are preserved, and exit codes pass through unchanged.
  - Target test scope: add focused e2e coverage near `tests/e2e/020-dispatch-and-path-precedence.sh` for explicit-path executables with pass-through args and failure code assertions.

- [x] E2E Gap: global flag parser error contracts (`internal/dispatch/run.go` `parseGlobalFlags`).
  - Affected flow: `bus --unknown`, `bus -C`, `bus --format`, and `bus --` without subcommand.
  - Expected e2e behavior: emit `bus: invalid usage: ...` on stderr, print usage, and exit 2 with no stdout.
  - Target test scope: `tests/e2e.sh` global-flag error-path block.

- [x] E2E Gap: pass-through shared flag forwarding to subcommand (`internal/dispatch/run.go` `parseGlobalFlags` + `Run`).
  - Affected flow: `-v`/`-vvv`, `--verbose`, `--no-color`, `--color=<mode>`, `-o/--output`, `-f/--format` forwarding before subcommand.
  - Expected e2e behavior: selected `bus-<command>` receives forwarded flags unchanged and positional args preserved.
  - Target test scope: `tests/e2e.sh` dispatch success-path assertions with argv echo fake module.

- [x] E2E Gap: busfile CLI override validation for transaction/scope (`internal/dispatch/run.go` `parseBusfileMode` + `runBusfilesWithExecutor`).
  - Affected flow: `--transaction=<value>` / `--scope=<value>` valid override plus invalid values and missing-value errors.
  - Expected e2e behavior: valid overrides execute with matching provider/scope semantics; invalid/missing values fail with `bus: invalid usage: ...` and exit 2.
  - Target test scope: `tests/e2e.sh` busfile-option matrix with deterministic stderr and exit-code checks.

- [x] E2E Gap: busfile parser edge cases for include/tokenization (`internal/dispatch/run.go` `collectBusfileCommands` + `tokenizeBusLine`).
  - Affected flow: include cycle detection, logical line continuation EOF (`\` at EOF), disallowed tokens (`|`, `;`, `<`, `>`), and escaped/quoted token forms.
  - Expected e2e behavior: syntax errors report precise `file:line` context with exit 65; valid escaped/quoted forms dispatch successfully.
  - Target test scope: `tests/e2e.sh` parser edge-case fixtures under isolated temp workspace.

- [x] E2E Gap: fs transaction recovery and warning path (`internal/dispatch/run.go` `recoverPendingFSTransactions`).
  - Affected flow: preexisting `.bus/tx` incomplete artifacts (file and directory forms) before busfile execution.
  - Expected e2e behavior: stale artifacts are removed, warning lines are printed to stderr, and subsequent transaction run succeeds.
  - Target test scope: `tests/e2e.sh` fs provider setup creating synthetic stale journal/artifact cases.

- [x] E2E Gap: config precedence from preferences over datapackage (`internal/dispatch/run.go` `applyDatapackageConfig` + `applyPreferencesConfig`).
  - Affected flow: `BUS_PREFERENCES_PATH` file sets `bus.busfile.dispatch.shell_lookup_enabled` / transaction keys conflicting with `datapackage.json`.
  - Expected e2e behavior: preference values take precedence and drive dispatch/fallback behavior deterministically.
  - Target test scope: `tests/e2e.sh` preference-precedence scenarios using per-test JSON fixtures.

- [x] E2E Gap: shebang busfile-path detection and mode switching (`internal/dispatch/run.go` `isBusfilePath` + `parseBusfileMode`).
  - Affected flow: invoking `bus <path-without-.bus-extension>` where file has `#!/usr/bin/bus` or `#!/usr/bin/env bus`, and non-bus shebang/control file variants.
  - Expected e2e behavior: bus shebang file is treated as busfile input (not subcommand dispatch), while non-bus shebang files are treated as CLI subcommands/args.
  - Target test scope: add focused scenario under `tests/e2e/` that executes both positive and negative shebang-detection cases with deterministic stdout/stderr and exit-code checks.

- [x] E2E Gap: file-scope fs transaction isolation across multi-file runs (`internal/dispatch/run.go` `partitionCommandsByFile` + `executeBusfileCommandsFS`).
  - Affected flow: `transaction.provider=fs`, `transaction.scope=file`, and multiple busfiles where one unit succeeds and a later unit fails.
  - Expected e2e behavior: successful earlier file unit is committed and preserved, failing later file unit is rolled back, and process exits with the failing command code/context.
  - Target test scope: add an fs tx e2e case using `txwrite` across two busfiles to assert per-file commit/rollback boundaries.

- [x] E2E Gap: dispatcher unexpected exec-error contract (`internal/dispatch/run.go` `Run` command execution error branch).
  - Affected flow: located `bus-<command>` exists but cannot be executed successfully by the OS (non-`ExitError` path, e.g. invalid interpreter).
  - Expected e2e behavior: stderr emits short `bus: <exec error>` diagnostic and dispatcher exits 1 (not 127 and not subcommand passthrough code).
  - Target test scope: add one deterministic fixture command that triggers an exec startup failure and assert exact exit code and `bus:`-prefixed stderr contract.

- [x] E2E Gap: `--check` preflight dispatch/error behavior (`internal/dispatch/run.go` `runBusfilesWithExecutor` + `preflightDispatchTargets`).
  - Affected flow: `bus --check <file.bus>` where targets are missing from PATH or blocked by `dispatch.shell_lookup_enabled=false`.
  - Expected e2e behavior: `--check` still performs dispatch preflight and returns 127 with `file:line` dispatch diagnostics when target resolution fails (rather than silently succeeding).
  - Target test scope: add focused `--check` error-path cases under `tests/e2e/` with deterministic stderr and exit-code assertions.

- [x] E2E Gap: busfile context env propagation to module execution (`internal/dispatch/run.go` `withBusBatchEnv` + `withBusfileEnv`).
  - Affected flow: busfile command execution should set/overwrite `BUS_BATCH=1`, `BUS_BUSFILE`, `BUS_BUSFILE_LINE`, and `BUS_TRANSACTION_PROVIDER` (for `provider=fs`) in child/module environment.
  - Expected e2e behavior: invoked module observes correct env values per command line and provider mode, including stable line attribution across multi-command files.
  - Target test scope: add a deterministic env-echo fixture command in `tests/e2e/` and assert emitted env values for `provider=none` and `provider=fs`.

- [x] E2E Gap: unsupported transaction-provider fallback vs strict failure (`internal/dispatch/run.go` `resolveTransactionProvider`).
  - Affected flow: config/preference sets `transaction.provider` to supported-but-unimplemented values (`git`, `snapshot`, `copy`) under both `fallback_to_none=true` and strict/CLI-override paths.
  - Expected e2e behavior: fallback mode emits warning and executes with `none`; strict mode (or explicit CLI override) fails usage with exit 2 and provider-specific diagnostic.
  - Target test scope: add provider-matrix e2e scenarios covering datapackage + CLI override combinations and exact stderr contract checks.

## Completed

- [x] Optimize `runModuleViaTempWorkspaceAndMerge` full-workspace staging path in `internal/dispatch/run.go` (copy + snapshot + full-tree merge per command). Direction: avoid whole-tree copy/diff when command touches a narrow file set; move toward change-scoped capture/merge while keeping TxFS correctness.
- [x] Optimize repeated `txfs.OpenFile` writes on already-materialized paths in `internal/txfs/txfs.go` (steady-state append/create loops). Direction: add a lower-overhead fast path that minimizes repeated path normalization/index mutations and overlay-dir checks when `changeReplace` is already established.

- [x] Add dispatcher-level `--perf` global flag forwarding and help/docs coverage in `bus`, so performance tracing can be enabled consistently for subcommands without breaking existing global-flag behavior, with unit tests and e2e coverage in the same change.
- [x] Add busfile session-level sticky global-flag directives in `bus` so any dispatcher global flag can be set on its own line and applied to following commands with deterministic override semantics (for example later `--chdir` replaces earlier `--chdir`), including unit tests, focused e2e coverage, and docs updates.
