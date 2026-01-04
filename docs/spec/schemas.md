# Schemas

## What this is
The schema document model, registration in the manifest, and validation rules.

## Schema document (binding)
Logical fields:
- `kind: bus.schema`
- `version: 1`
- `name: <schemaName>`
- `properties: [...]`
- optional schema specific `operations` (`create|list|show|update|delete` booleans)

## Property rules (binding)
- At least one property exists.
- Exactly one property has `primary: true`.
- Primary type is `string|uuid|int`.
- Property names are unique.
- Each property may have optional `operations` (`create|list|show|update|delete` booleans) which override the schema specific operations for the property

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
- optional `path` (unique; resolved relative to manifest)

## CLI schema property spec (binding)
`bus schema init` accepts repeated:
- `--property name[:type][,attr[,attr...]]`

Attributes:
- `primary`, `required`, `unique` or any of accepted operations (`create|list|show|update|delete`)
  - Operations may have a leading negative to remove an operation, e.g. `-update,-delete` means the property is readonly.
- optional `type`, defaults to `string`
- optional `uniqueScope=<a|b|c>` (pipe-separated)
