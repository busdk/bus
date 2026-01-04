# Units

## What this is
Unit document format and validation rules for create/list/show.

Internal state storage (records + indexes) is backend-defined and pluggable:
- filesystem backend: `docs/spec/state-backend-dotbus.md`
- database backend: `docs/spec/state-backend-database.md`

## Unit document (binding)
Logical fields:
- `kind: bus.unit`
- `version: 1`
- `schema: <schemaName>`
- `data: { ... }`

## Storage layout
Backend-defined (see links above). Regardless of backend:
- listing ids is deterministic
- records are keyed by `(schemaName, primaryId)`

## Primary id rules (binding)
- String primary id: user MUST provide.
- UUID primary id: user may provide; otherwise Bus auto-generates (random next free UUID).
- Int primary id: user MUST provide; otherwise Bus auto-generates (auto incremental next free ID).

File naming rule:
- record filename and `.ids` entry MUST match the primary id string form.

## Create input rules (binding)
Unit creation uses positional `key=value` tokens (see `docs/spec/transports-cli.md`).

Validation:
- unknown keys error
- duplicate keys error
- required keys present (after UUID autogen rules)
- types match schema
- `ref:<schema>` points to an existing unit id in the referenced schema
- uniqueness enforced, including scoped uniqueness (`uniqueScope`)
