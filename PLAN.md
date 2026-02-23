# bus plan

## Active feature plan

- [ ] FR57/Task 1: Define and document dispatcher command boundaries for identity and plan features.
  - Implementation: add dispatcher-level command discoverability/usage coverage for `auth` and `plan` command families without adding business logic to `bus`.
  - Documentation: update `README.md` and docs module pages to describe `bus auth ...` and `bus plan ...` delegation model and module split rationale.
  - Unit tests: add/adjust dispatcher tests for missing subcommand diagnostics and deterministic listing when `bus-auth` and `bus-plan` are present/absent.
  - E2E tests: add dispatcher e2e scenarios validating pass-through args/exit codes for `bus auth` and `bus plan`.

- [ ] FR57/Task 2: Define stable global UX contract for auth/plan CLIs at the dispatcher boundary.
  - Implementation: specify command synopsis and shared error-envelope expectations at dispatcher-facing integration points only (no auth/plan backend logic in `bus`; CLI modules remain thin clients).
  - Documentation: add concise command contract table in docs (`modules/bus.md` + `sdd/bus.md`) for discoverability and compatibility notes.
  - Unit tests: verify usage/help output remains deterministic with the expanded command surface.
  - E2E tests: validate end-to-end dispatcher behavior with `--help`, unknown subcommands, and exit-code pass-through for new module commands.

## E2E coverage gaps

- [ ] E2E Gap: global flag parser error contracts (`internal/dispatch/run.go` `parseGlobalFlags`).
  - Affected flow: `bus --unknown`, `bus -C`, `bus --format`, and `bus --` without subcommand.
  - Expected e2e behavior: emit `bus: invalid usage: ...` on stderr, print usage, and exit 2 with no stdout.
  - Target test scope: `tests/e2e.sh` global-flag error-path block.

- [ ] E2E Gap: pass-through shared flag forwarding to subcommand (`internal/dispatch/run.go` `parseGlobalFlags` + `Run`).
  - Affected flow: `-v`/`-vvv`, `--verbose`, `--no-color`, `--color=<mode>`, `-o/--output`, `-f/--format` forwarding before subcommand.
  - Expected e2e behavior: selected `bus-<command>` receives forwarded flags unchanged and positional args preserved.
  - Target test scope: `tests/e2e.sh` dispatch success-path assertions with argv echo fake module.

- [ ] E2E Gap: busfile CLI override validation for transaction/scope (`internal/dispatch/run.go` `parseBusfileMode` + `runBusfilesWithExecutor`).
  - Affected flow: `--transaction=<value>` / `--scope=<value>` valid override plus invalid values and missing-value errors.
  - Expected e2e behavior: valid overrides execute with matching provider/scope semantics; invalid/missing values fail with `bus: invalid usage: ...` and exit 2.
  - Target test scope: `tests/e2e.sh` busfile-option matrix with deterministic stderr and exit-code checks.

- [ ] E2E Gap: busfile parser edge cases for include/tokenization (`internal/dispatch/run.go` `collectBusfileCommands` + `tokenizeBusLine`).
  - Affected flow: include cycle detection, logical line continuation EOF (`\` at EOF), disallowed tokens (`|`, `;`, `<`, `>`), and escaped/quoted token forms.
  - Expected e2e behavior: syntax errors report precise `file:line` context with exit 65; valid escaped/quoted forms dispatch successfully.
  - Target test scope: `tests/e2e.sh` parser edge-case fixtures under isolated temp workspace.

- [ ] E2E Gap: fs transaction recovery and warning path (`internal/dispatch/run.go` `recoverPendingFSTransactions`).
  - Affected flow: preexisting `.bus/tx` incomplete artifacts (file and directory forms) before busfile execution.
  - Expected e2e behavior: stale artifacts are removed, warning lines are printed to stderr, and subsequent transaction run succeeds.
  - Target test scope: `tests/e2e.sh` fs provider setup creating synthetic stale journal/artifact cases.

- [ ] E2E Gap: config precedence from preferences over datapackage (`internal/dispatch/run.go` `applyDatapackageConfig` + `applyPreferencesConfig`).
  - Affected flow: `BUS_PREFERENCES_PATH` file sets `bus.busfile.dispatch.shell_lookup_enabled` / transaction keys conflicting with `datapackage.json`.
  - Expected e2e behavior: preference values take precedence and drive dispatch/fallback behavior deterministically.
  - Target test scope: `tests/e2e.sh` preference-precedence scenarios using per-test JSON fixtures.

- [ ] E2E Gap: shebang busfile-path detection and mode switching (`internal/dispatch/run.go` `isBusfilePath` + `parseBusfileMode`).
  - Affected flow: invoking `bus <path-without-.bus-extension>` where file has `#!/usr/bin/bus` or `#!/usr/bin/env bus`, and non-bus shebang/control file variants.
  - Expected e2e behavior: bus shebang file is treated as busfile input (not subcommand dispatch), while non-bus shebang files are treated as CLI subcommands/args.
  - Target test scope: add focused scenario under `tests/e2e/` that executes both positive and negative shebang-detection cases with deterministic stdout/stderr and exit-code checks.

- [ ] E2E Gap: file-scope fs transaction isolation across multi-file runs (`internal/dispatch/run.go` `partitionCommandsByFile` + `executeBusfileCommandsFS`).
  - Affected flow: `transaction.provider=fs`, `transaction.scope=file`, and multiple busfiles where one unit succeeds and a later unit fails.
  - Expected e2e behavior: successful earlier file unit is committed and preserved, failing later file unit is rolled back, and process exits with the failing command code/context.
  - Target test scope: add an fs tx e2e case using `txwrite` across two busfiles to assert per-file commit/rollback boundaries.

- [ ] E2E Gap: dispatcher unexpected exec-error contract (`internal/dispatch/run.go` `Run` command execution error branch).
  - Affected flow: located `bus-<command>` exists but cannot be executed successfully by the OS (non-`ExitError` path, e.g. invalid interpreter).
  - Expected e2e behavior: stderr emits short `bus: <exec error>` diagnostic and dispatcher exits 1 (not 127 and not subcommand passthrough code).
  - Target test scope: add one deterministic fixture command that triggers an exec startup failure and assert exact exit code and `bus:`-prefixed stderr contract.

## Completed

- [x] Optimize `runModuleViaTempWorkspaceAndMerge` full-workspace staging path in `internal/dispatch/run.go` (copy + snapshot + full-tree merge per command). Direction: avoid whole-tree copy/diff when command touches a narrow file set; move toward change-scoped capture/merge while keeping TxFS correctness.
- [x] Optimize repeated `txfs.OpenFile` writes on already-materialized paths in `internal/txfs/txfs.go` (steady-state append/create loops). Direction: add a lower-overhead fast path that minimizes repeated path normalization/index mutations and overlay-dir checks when `changeReplace` is already established.
