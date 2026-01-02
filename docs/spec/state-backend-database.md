# Database State Backend

## What this is
A database-backed internal state backend. When selected, internal mutable state (records, indexes, locks) is stored in a database instead of `.bus/`.

Schema configuration (manifest + schema docs) remains filesystem-based in the workspace.

## Binding rules
- The database backend MUST be tenant-scoped (tenant id is part of every query/write).
- Atomicity is provided via database transactions.
- Concurrency control uses database locking semantics (e.g., advisory locks) to serialize tenant mutators.

## Expected capabilities
At minimum, the DB backend must support:
- storing unit records keyed by `(tenantId, schemaName, primaryId)`
- listing ids per `(tenantId, schemaName)` deterministically
- enforcing uniqueness constraints needed by the unit store
- transactional multi-write operations

## Migration and compatibility
- Backend-owned schema migrations must be deterministic and versioned.
- Switching backends should not change schema config files; it only changes internal state storage.
