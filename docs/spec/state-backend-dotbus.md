# Filesystem State Backend (`.bus/`)

## What this is
The filesystem-based internal state backend. When selected, Bus stores internal mutable state under `.bus/`.

Schema configuration (manifest + schema docs) remains outside `.bus/` (workspace files).

## Binding rules
- `.bus/` is the only Bus-controlled directory **when the filesystem backend is selected**.
- Tenant-scoped internal state lives under:
  - `.bus/tenants/<tenantId>/...`

## Suggested layout (v1-style)
- `.bus/tenants/<tenantId>/lock` (lock file)
- `.bus/tenants/<tenantId>/units/<schemaName>.ids` (newline-delimited ids)
- `.bus/tenants/<tenantId>/units/<schemaName>/<primaryId>.<ext>` (unit record documents)

## Atomicity + locking
- Locking uses a tenant-scoped file lock.
- Atomic writes use temp+rename in the same directory.
- Multi-file operations follow commit-point ordering:
  - write per-record files first
  - update indexes/manifest last
