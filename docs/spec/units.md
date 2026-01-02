# Units

## What this is
Unit document format, storage layout, and validation rules for create/list/show.

## Unit document (binding)
Logical fields:
- `kind: bus.unit`
- `version: 1`
- `schema: <schemaName>`
- `data: { ... }`

## Storage layout (binding)
Tenant-scoped under `.bus/tenants/<tenantId>/`:

- Index: `units/<schemaName>.ids`
  - one id per line
  - newline at EOF
  - no blank lines
  - deterministic ordering (recommended: lexicographic sort)

- Records: `units/<schemaName>/<primaryId>.<ext>`

## Primary id rules (binding)
- String primary id: user MUST provide.
- UUID primary id: user may provide; otherwise Bus auto-generates.
- Int primary id: user MUST provide.

File naming rule:
- record filename and `.ids` entry MUST match the primary id string form.

## Create input rules (binding)
Unit creation uses positional `key=value` tokens (see `docs/spec/transports-cli.md`).

Validation:
- unknown keys error
- duplicate keys error
- required keys present (after UUID autogen rules)
- types match schema
- `ref:<schema>` points to an existing unit id in the referenced schema (same tenant)
- uniqueness enforced, including scoped uniqueness (`uniqueScope`)


