# bus

`bus` is the BusDK top-level dispatcher. It executes `bus-<command>` binaries
from `PATH` and stays intentionally minimal.

This repository must stay independent from private `bus-*` module repositories.
It integrates with modules only by executing `bus-<command>` processes from
`PATH`; it must not import private module Go packages or add build-time sibling
module dependencies.

Canonical references:
- Design: https://docs.busdk.com/
- Bus SDD: sdd/docs/modules/bus.md
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

Before dispatch, `bus` reads `.env` from the effective working directory when
that file exists. Values are added to the child command environment, but
variables already present in the process environment keep their existing values.
Use `KEY=VALUE` or `export KEY=VALUE` lines; blank lines and `#` comments are
ignored. The dispatcher does not filter names to Bus-specific variables; any
valid environment variable name is passed through to the child command.

Nested command families still dispatch to the first command word owner. For
example:

```sh
bus operator billing catalog sync
```

runs `bus-operator billing catalog sync`. If `bus-operator` wants focused
operator families such as billing or Stripe, it must dispatch to those
implementations through Go library imports, not by relying on `bus` to execute
nested child binaries.

Bootstrap and managed-package flows use the same delegation model:

```sh
bus update --workspace /srv/busdk --dry-run
bus update package install --module bus-ledger
```

These commands are still pure dispatcher calls. `bus` resolves and runs
`bus-update` from `PATH`, and all installer/package-manager behavior remains in
that module rather than in the dispatcher.

Special audit alias:

```sh
bus audit evidence-coverage [args...]
```

If `bus-audit` is not available, this delegates to `bus-validate evidence-coverage`
for deterministic evidence-coverage reporting. Alias-local help is also supported:
`bus audit evidence-coverage -h` and `bus audit evidence-coverage --help` print the
underlying evidence-coverage help surface and exit `0`.

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

Command names may contain hyphens. Hyphenated binaries can be invoked either
directly as one command word when appropriate, or as nested words when installed
as a longer command prefix.

## Exit codes

- `0`: successful dispatch
- `2`: usage output (`bus` with no args, or `bus help` when `bus-help` missing)
- `127`: missing subcommand (`bus-<name>` not found)
- `1`: unexpected execution failure after lookup
- Any subcommand non-zero exit code is passed through unchanged

## Global flags

This module forwards subcommand global flags such as `--color`, `--format`,
`--output`, `--quiet`, and `--chdir` to the selected `bus-*` binary. It also
accepts dispatcher-level `--perf`, which enables timing output for the
dispatched command and sets `BUS_PERF=1` for instrumented modules.

Dispatcher diagnostics use the shared Bus levels `ERROR`, `WARN`, `INFO`,
`DEBUG`, and `TRACE`. Default output uses `INFO`; one `-v` or `--verbose`
enables `DEBUG`; repeated verbosity such as `-vv` or `--verbose --verbose`
enables `TRACE`; `--trace` is equivalent to TRACE; and `--quiet` keeps only
ERROR diagnostics. `--quiet` cannot be combined with verbose or trace mode.

In `.bus` files, a line that contains only dispatcher global flags becomes a
sticky directive for following commands in the same session. The same parser is
used as in normal dispatch, so later single-value flags override earlier ones.
For example:

```bus
--perf
--chdir data
accounts alpha
--chdir reports
--no-perf
ledger beta
```

Reset directives are supported for sticky state: `--no-perf`, `--no-quiet`,
`--no-verbose`, `--no-chdir`, `--no-output`, and `--no-format`. `--no-verbose`
also clears sticky trace mode. Color already
uses `--color ...` and `--no-color`.

Dispatcher-level `-C/--chdir` also applies when the invocation enters busfile
mode. Relative busfile paths are resolved after switching into that workspace,
and the executed module commands inherit the same effective working directory.
For replay authoring, dispatcher preflight accepts the same documented
`bus journal add` posting syntax as the module itself, including
`ACCOUNT=AMOUNT=ROW_DESCRIPTION` with quoted UTF-8 text and spaces in `.bus`
files and under `bus --check`.

## Editor support for `.bus`

The repository ships three editor-tooling layers for `.bus` files:

- a VS Code compatible extension under `editors/vscode-bus-language/`
- a Tree-sitter grammar under `editors/tree-sitter-bus/`
- a lightweight stdio language server at `editors/vscode-bus-language/language-server.js`

The VS Code-compatible package provides syntax highlighting, file association,
and semantic tokens for BusDK command files in editors such as VS Code and
Cursor, including sticky directive lines, assignments, and line continuations
that are common in real busfiles.

To package the installable `.vsix` artifact from source, run:

```sh
make package-vscode-extension
```

The command writes a versioned `.vsix` file into `./bin/`. That artifact is the
intended release/downloadable package for users. Maintainers can also validate
the supported release surfaces and Open VSX-compatible metadata with:

```sh
make check-vscode-extension-release
```

The release surface contract is:

- `.vsix` release asset first for VS Code, Cursor, Windsurf, and other editors
  that accept local VS Code extension packages
- Open VSX-compatible metadata under `busdk.language-bus` / `busdk/language-bus`
  for VSCodium-style distribution paths

For parser-backed highlighting in editors that use Tree-sitter, use the grammar
and highlight query in `editors/tree-sitter-bus/`.

For editors that can talk to an stdio language server, run:

```sh
node editors/vscode-bus-language/language-server.js --stdio
```

That server exposes semantic tokens for command targets, flags, assignments,
strings, dates, and numbers by standard LSP token classes.

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
## Machine-Readable Help

Live OpenCLI-compatible metadata is available on stdout and includes Bus `io.busdk.environment` metadata for `bus configure`:

```sh
bus help --format opencli
```
