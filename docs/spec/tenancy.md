# Tenancy (Isolation From Day One)

## What this is
Tenant selection and isolation rules that apply to all core interfaces and transport-exposed capabilities.

## Tenant selection (binding)
Resolution order for `TenantID`:
1. Explicit request/CLI flag (`--tenant`)
2. Manifest default tenant
3. Fallback to `"default"`

## Isolation (binding)
Rules:
- No storage interface may read/write without `TenantID`.
- Reads and writes MUST NOT cross tenant boundaries by default.

## Internal state scoping (binding)
Internal mutable state MUST be tenant-scoped regardless of storage backend.

Backend-specific roots and layouts are defined in:
- `docs/spec/state-storage.md`
- `docs/spec/state-backend-dotbus.md`
- `docs/spec/state-backend-database.md`


