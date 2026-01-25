# bus

[![Go](https://img.shields.io/badge/go-1.22-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

`bus` is a tiny, git-style subcommand dispatcher. It solves one problem: route
`bus <command> ...args...` to a matching `bus-<command>` executable on your
`PATH` with zero built-in business logic. This keeps the core CLI stable while
allowing features to live in separate binaries.

## Table of contents

- [Installation](#installation)
- [Usage](#usage)
- [How dispatch works](#how-dispatch-works)
- [Command discovery](#command-discovery)
- [Exit codes](#exit-codes)
- [Design rationale](#design-rationale)
- [Features](#features)
- [Support](#support)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Tests](#tests)
- [Credits](#credits)
- [License](#license)
- [Project status](#project-status)

## Installation

Prerequisites:

- Go 1.22+

Install from a local clone:

```sh
git clone https://github.com/busdk/bus.git
cd bus
go install ./cmd/bus
```

Or install from source (replace with your desired version tag):

```sh
go install github.com/busdk/bus/cmd/bus@latest
```

Or install via Makefile into a user-local prefix:

```sh
make install
```

Override the install prefix (default is `$(HOME)/.local`):

```sh
make install PREFIX="$HOME/.local"
```

Ensure the install location is on your PATH:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Uninstall the user-local binary:

```sh
make uninstall
```

## Usage

Quick start:

```sh
bus
```

Expected output (example):

```text
usage: bus <command> [args...]

available commands:
  accounts
  assets
  attachments
  bank
  budget
  entities
  filing
  filing-prh
  filing-vero
  inventory
  invoices
  journal
  payroll
  period
  reconcile
  reports
  validate
  vat
```

This list comes from scanning your `PATH`. If you install additional `bus-*`
executables, they will appear here automatically.

Example module binaries that `bus` can dispatch to include:

- `bus-accounts`
- `bus-assets`
- `bus-attachments`
- `bus-bank`
- `bus-budget`
- `bus-entities`
- `bus-filing`
- `bus-filing-prh`
- `bus-filing-vero`
- `bus-inventory`
- `bus-invoices`
- `bus-journal`
- `bus-payroll`
- `bus-period`
- `bus-reconcile`
- `bus-reports`
- `bus-validate`
- `bus-vat`

Dispatching to a module binary:

```sh
bus accounts summary --month=2026-01
```

`bus` forwards `summary --month=2026-01` to the `bus-accounts` executable.


## How dispatch works

When you run:

```sh
bus <command> ...args...
```

`bus` looks for an executable named `bus-<command>` on your `PATH` and runs it,
passing through all remaining arguments unchanged. Standard input, output, and
error are wired through directly, and the subcommand inherits the parent
environment. When run without arguments, `bus` prints a usage line and lists
all discoverable `bus-*` executables it can find on `PATH`.

## Command discovery

`bus` scans your `PATH` to list available commands for discoverability. The
listing is deterministic for a given `PATH`:

- PATH entries are processed left to right.
- Each directory is scanned non-recursively for executables named `bus-*`
  (or `bus-*.exe` on Windows).
- Inaccessible or missing directories are skipped silently.
- If multiple PATH directories contain the same command name, the earliest
  entry in PATH wins (deduplicated by command name).
- The final list is sorted lexicographically by command name.

## Exit codes

- `2` usage error (missing subcommand)
- `127` subcommand missing (`bus-<command>` not found on `PATH`)
- pass-through for subcommand failures (exact exit code returned)
- `1` unexpected execution errors (for example, failure after lookup)

## Design rationale

This repository intentionally stays minimal. All features and business logic
live in separate repositories and are provided as standalone `bus-<name>`
executables. The core `bus` command exists only to dispatch and enumerate
available subcommands from `PATH`.

## Features

- Deterministic dispatch to `bus-<command>` executables
- Help output that lists discoverable `bus-*` commands
- Transparent stdin/stdout/stderr pass-through
- Exit code parity with the invoked subcommand

## Support

Use the issue tracker for questions, bugs, or feature requests:
`https://github.com/busdk/bus/issues`

## Roadmap

- Keep the core minimal while improving discoverability and ergonomics
- Add documentation for writing `bus-*` subcommands

## Contributing

Contributions are welcome. Please open a pull request from a fork or topic
branch and explain the change. Run the test suite locally before submitting:

```sh
go test ./...
```

See `CONTRIBUTING.md` for the full workflow and expectations.

## Tests

Run the full suite:

```sh
go test ./...
```

Tests are expected to be deterministic and to validate dispatch behavior
without requiring any external binaries to be installed.

## Credits

- Maintainer: Heusala Group Oy
- Inspired by Git's subcommand model

## License

MIT licensed. See `LICENSE`.

## Project status

Active and intentionally minimal. The core `bus` command is stable; features
ship as separate `bus-*` binaries.
