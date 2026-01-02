# Locking + Atomic Writes

## What this is
Consistency rules that prevent partial writes and serialize concurrent mutating operations.

## Locking (binding)
- Any mutating command MUST acquire a lock before writing.
- Lock scope is tenant-scoped by default (to allow parallelism across tenants later).

Suggested lock file:
- `.bus/tenants/<tenantId>/lock`

## Atomic file writes (binding)
Write strategy:
1. Write to `.<name>.tmp` in the same directory.
2. `fsync` the file (recommended).
3. Atomically rename temp to final path.

## Multi-file operations (binding)
Commit-point rule:
- Write per-record files first.
- Rewrite indexes/manifest last (indexes act as the commit point).
- If failure occurs: do not update indexes/manifest; delete newly created per-record files when possible.


