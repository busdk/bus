# Schemas

## What this is
The schema document model, registration in the manifest, and validation rules.

## Schema document (binding)
Logical fields:
- `kind: bus.schema`
- `version: 1`
- `name: <schemaName>`
- `properties: [...]`
- optional `operations` (`create|list|show|update|delete` booleans; update/delete reserved for future)

## Property rules (binding)
- At least one property exists.
- Exactly one property has `primary: true`.
- Primary type is `string|uuid|int`.
- Property names are unique.

## Types (v1 intent preserved)
Supported:
- `uuid`
- `int`
- `bool`
- `string`
- `date` (`YYYY-MM-DD`)
- `timestamp` (RFC 3339)
- `ref:<SchemaName>` (typed reference to another schema’s primary id)

## Registration in manifest (binding)
Manifest contains schema references under `units[]`:
- `name` (unique)
- `path` (unique; resolved relative to manifest)

## CLI schema property spec (binding)
`bus schema init` accepts repeated:
- `--property name:type[,attr[,attr...]]`

Attributes:
- `primary`, `required`, `unique`
- optional `uniqueScope=<a|b|c>` (pipe-separated)


