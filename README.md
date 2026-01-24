# bus

`bus` is a tiny, git-style subcommand dispatcher. It contains no accounting
logic and knows nothing about specific modules beyond the dispatch rule.

## How dispatch works

When you run:

```
bus <command> ...args...
```

`bus` looks for an executable named `bus-<command>` on your `PATH` and runs it,
passing through all remaining arguments unchanged. Standard input/output/error
are wired through directly, and the subcommand runs with the parent environment.

## Exit codes

- `2` usage error (no subcommand provided)
- `127` subcommand missing (`bus-<command>` not found on `PATH`)
- passthrough for subcommand failures (exact exit code returned)
- `1` unexpected execution errors (for example, failure after lookup)

## Example install

If you build or install a module binary named `bus-accounts` on your `PATH`,
then `bus accounts` will dispatch to it:

```
go install github.com/example/bus-accounts@latest
bus accounts summary
```

## Design rationale

This repository intentionally stays minimal. All features and business logic
live in separate repositories and are provided as standalone `bus-<name>`
executables. The core `bus` command exists only to dispatch.

## Contributing

- Run `go test ./...`; tests must pass.
- This repo intentionally contains no business logic.
