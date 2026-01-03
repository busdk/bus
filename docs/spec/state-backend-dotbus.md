# Filesystem State Backend (`.bus/`)

## What this is
The filesystem-based internal state backend. When selected, Bus stores internal mutable state under `.bus/`.

Schema configuration (manifest + schema docs) remains outside `.bus/` (workspace files).

## Binding rules
- `.bus/state/` is the only Bus-controlled directory **when the filesystem backend is selected**.
- Internal state lives under:
  - `.bus/state/...`

## Suggested layout (v1-style)
- `.bus/state/lock` (lock file)
- `.bus/state/units/<schemaName>.ids` (newline-delimited ids)
- `.bus/state/units/<schemaName>/<primaryId>.<ext>` (unit record documents)

## Atomicity + locking
- Atomic writes use temp+rename in the same directory.
- Multi-file operations follow commit-point ordering:
  - write per-record files first
  - update indexes/manifest last
