# AGENTS.md — bus/tests

Scope: applies to `bus/tests/**`.

## E2E layout

1. Keep `tests/e2e.sh` as a simple runner only.
2. Put actual scenarios under `tests/e2e/` as numbered files: `NNN-name.sh`.
3. Keep names compact but specific to one behavior (for example `070-fs-transaction-commit-rollback-largefile.sh`).
4. One file should cover one feature/behavior only.
5. Run scripts in deterministic numeric order from the runner.
6. Runner output style must be shell-trace driven: use `set -x` in `tests/e2e.sh` so each executed test appears as one command line (for example `+ bash .../tests/e2e/050-trace-output.sh`).
7. Runner must hash-check the test subject binary (`bin/bus`) before and after each test case and fail immediately if it changes during the run.

## E2E script style

1. Keep scripts as simple, linear command/assert flows.
2. Avoid complex Bash programming patterns in test files.
3. Reuse shared helpers from `tests/e2e/lib.sh` when needed.
4. Prefer explicit checks (`test`, `grep`, `diff`, exit-code asserts) over abstractions.
5. Keep tests hermetic and local-only.
6. Keep success output minimal in per-test scripts: print only `e2e OK` when the test passes.

## Change policy

1. New user-visible CLI behavior requires a new focused `tests/e2e/NNN-*.sh` script or a focused update to one existing script.
2. Do not re-grow monolithic e2e scripts; split by behavior when touching e2e coverage.
