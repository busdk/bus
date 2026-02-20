# Contributing to `bus`

## Scope

`bus` is the BusDK dispatcher. Keep contributions focused on:

- command dispatch (`bus <module> ...`)
- `.bus` command-file parsing and orchestration
- deterministic CLI/error behavior

Do not add module business logic to this repository.

## Local workflow

1. Build:
```bash
make build
```
2. Unit tests:
```bash
make test
```
3. End-to-end tests:
```bash
make e2e
```
4. Full checks:
```bash
make check
```

## Pull request expectations

- Include tests for every behavior change.
- Keep output and exit-code changes explicit in tests.
- Update docs in `../docs/docs/modules/bus.md` and `../docs/docs/sdd/bus.md` when behavior changes.
- Keep implementation minimal and deterministic.
