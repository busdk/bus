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

## Completed

- [x] Optimize `runModuleViaTempWorkspaceAndMerge` full-workspace staging path in `internal/dispatch/run.go` (copy + snapshot + full-tree merge per command). Direction: avoid whole-tree copy/diff when command touches a narrow file set; move toward change-scoped capture/merge while keeping TxFS correctness.
- [x] Optimize repeated `txfs.OpenFile` writes on already-materialized paths in `internal/txfs/txfs.go` (steady-state append/create loops). Direction: add a lower-overhead fast path that minimizes repeated path normalization/index mutations and overlay-dir checks when `changeReplace` is already established.
