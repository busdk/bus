# Database State Backend

## What this is
A database-backed internal state backend. When selected, internal mutable state (records, indexes, locks) is stored in a database instead of `.bus/state/`.

Schema configuration (manifest + schema docs) remains filesystem-based in the workspace.

## Binding rules
- Atomicity is provided via database transactions.

## Expected capabilities
At minimum, the DB backend must support:
- storing unit records keyed by `(schemaName, primaryId)`
- listing ids per `(schemaName)` deterministically
- enforcing uniqueness constraints needed by the unit store
- transactional multi-write operations

## Migration and compatibility
- Backend-owned schema migrations must be deterministic and versioned.
- Switching backends should not change schema config files; it only changes internal state storage.
