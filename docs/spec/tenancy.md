# Tenancy (Isolation From Day One)

## What this is
Tenant selection and isolation rules that apply to all core interfaces and module capabilities.

## Tenant selection (binding)
Resolution order for `TenantID`:
1. Explicit request/CLI flag (`--tenant`)
2. Manifest default tenant
3. Fallback to `"default"`

## Isolation (binding)
Rules:
- No storage interface may read/write without `TenantID`.
- Reads and writes MUST NOT cross tenant boundaries by default.

## Storage roots (binding)
All Bus-owned mutable state is under `.bus/`, tenant-scoped as:
- `.bus/tenants/<tenantId>/...`


