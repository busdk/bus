# Tenancy

## What this is
How Bus models multiple organizations in a single workspace while keeping data
isolated and queryable per tenant.

## Binding rules
- Every unit that represents business data MUST include an `org_id`.
- Core operations MUST require an active organization context.
- Cross-organization queries are not allowed in core operations.
- Default CLI behavior assumes a single org but allows explicit switching.

## Recommended unit schemas (v1 intent)
- `organization`: tenant profile and fiscal year metadata.
- `account`: chart of accounts scoped by `org_id`.
- `entry` and `entry_line`: ledger units scoped by `org_id`.
- `invoice`, `partner`, `bank_transaction`, `budget`: org-scoped business units.

## Storage implications
Tenancy is enforced at the query and validation layer. Storage backends may
physically share tables/files but must filter by `org_id`.
