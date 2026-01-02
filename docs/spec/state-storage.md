# Internal State Storage (Pluggable)

## What this is
Bus separates **workspace configuration** (manifest + schemas) from **internal mutable state** (indexes, records, locks, derived indexes).

- **Workspace configuration** is always **filesystem-based** in the workspace directory:
  - `bus.yml|bus.yaml|bus.toml|bus.json`
  - schema documents referenced by the manifest
- **Internal mutable state** is **pluggable** via a state backend:
  - filesystem state backend (stored under `.bus/`) is one option
  - database state backend is another option

## Design goals (binding)
- Core and modules MUST NOT assume `.bus/` exists.
- All internal state access goes through core-owned interfaces.
- A new state backend should be addable by implementing interfaces + registering a built-in provider, not by rewriting feature logic.

## Extension point interfaces (names are stable contracts)

### `StateBackend`
Responsibility: provide tenant-scoped storage primitives for internal mutable state.

Minimum responsibilities:
- create/read/write unit records (as typed documents)
- maintain queryable indexes needed by the unit store (e.g., ids list)
- provide locking/transaction semantics for mutating operations

### `StateTransaction` (optional but recommended)
Responsibility: provide atomic multi-write semantics appropriate for the backend.

Rule:
- Filesystem backends implement atomicity via temp+rename and commit-point ordering.
- Database backends implement atomicity via database transactions.

### `StateLock`
Responsibility: serialize concurrent mutating operations (tenant-scoped).

Rule:
- Lock semantics are tenant-scoped by default.
- The underlying mechanism is backend-defined (file lock, DB advisory lock, etc.).

## Configuration (manifest)
The manifest MAY include a state backend selection:

- `state.backend`: `"filesystem"` | `"database"` | `<implementation-id>`
- `state.config`: opaque object (validated by the selected backend provider)

Schema configuration remains filesystem-based regardless of `state.backend`.
