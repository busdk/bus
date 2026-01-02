# Locking + Atomic Writes

## What this is
Consistency rules that prevent partial writes and serialize concurrent mutating operations.

## Locking (binding)
- Any mutating command MUST acquire a lock before writing.
- Lock scope is tenant-scoped by default (to allow parallelism across tenants later).

Lock mechanism is backend-defined:
- filesystem backend: tenant-scoped file lock (see `docs/spec/state-backend-dotbus.md`)
- database backend: tenant-scoped database lock (e.g., advisory lock) (see `docs/spec/state-backend-database.md`)

## Atomic file writes (binding)
Atomicity mechanism is backend-defined:
- filesystem backend: temp+rename in the same directory (and fsync recommended)
- database backend: database transactions

## Multi-file operations (binding)
Commit-point rule:
- Write per-record files first.
- Rewrite indexes/manifest last (indexes act as the commit point).
- If failure occurs: do not update indexes/manifest; delete newly created per-record files when possible.

Note:
- In database backends, the equivalent is a single transaction commit (no “commit-point file”).


