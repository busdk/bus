# bus

`bus` is the BusDK top-level dispatcher. It executes `bus-<command>` binaries
from `PATH` and stays intentionally minimal.

Canonical references:
- Design: https://docs.busdk.com/
- Bus SDD: https://docs.busdk.com/sdd/bus
- Bus module docs: https://docs.busdk.com/modules/bus
- Command structure: https://docs.busdk.com/cli/command-structure

## Purpose

The dispatcher provides deterministic routing for:

```sh
bus <command> [args...]
```

It does not implement business logic, data I/O, Git operations, network calls,
or module-specific global flags.

## Usage

Run without arguments to print usage and discoverable commands:

```sh
bus
```

Dispatch to a subcommand executable:

```sh
bus accounts summary --month=2026-01
```

This resolves and runs `bus-accounts` from `PATH`, passing through arguments,
stdin, stdout, stderr, and environment unchanged.

### Special case: `help`

If `bus-help` exists on `PATH`, `bus help ...` dispatches to it.
If `bus-help` is missing, `bus help` prints usage and available commands and
exits with code `2`.

## Discoverability rules

When listing available commands, `bus`:
- Scans `PATH` left-to-right.
- Reads each directory non-recursively for `bus-*` executables
  (`bus-*.exe` on Windows).
- Silently skips missing/inaccessible entries.
- Deduplicates by command name, preferring the earliest `PATH` entry.
- Sorts command names lexicographically before printing.

## Exit codes

- `0`: successful dispatch
- `2`: usage output (`bus` with no args, or `bus help` when `bus-help` missing)
- `127`: missing subcommand (`bus-<name>` not found)
- `1`: unexpected execution failure after lookup
- Any subcommand non-zero exit code is passed through unchanged

## Global flags

This module intentionally does not implement dispatcher-level global flags such
as `--version`, `--color`, `--format`, `--output`, `--quiet`, `--dry-run`, or
`--chdir`. These are handled by subcommands per BusDK CLI conventions.

## Development

Build:

```sh
make build
```

Run unit tests:

```sh
make test
```

Run end-to-end tests:

```sh
make e2e
```

Run all checks:

```sh
make check
```
