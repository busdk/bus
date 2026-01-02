# Schema File Format

Schemas define the structure of units. They are document files (YAML/TOML/JSON) stored adjacent to the manifest (`bus.{yml,yaml,toml,json}`) by default.

Format is detected by file extension and preserved on rewrite (see `16-multi-format-storage.md`).

## Schema Kind

The schema kind is **always** `bus.schema`.

## Core Schema Fields

The examples below use YAML for readability, but the same logical schema can be expressed in TOML or JSON.

```yaml
kind: bus.schema
version: 1
name: user
properties:
  - name: id
    type: uuid
    primary: true
    required: true
    unique: true
  - name: serverId
    type: ref:server
    required: true
  - name: contractStart
    type: date
    required: true
  - name: contractEnd
    type: date
    required: false
  - name: monthlyFeeCents
    type: int
    required: true
  - name: currency
    type: string
    required: true
```

### Required Schema Fields

#### `kind`
* **MUST** be `bus.schema`

#### `version`
* **MUST** be `1` in v1

#### `name`
* Schema name (must match the name in the manifest)

#### `properties`
* Array of property definitions
* At least one property must be defined
* Exactly one property must have `primary: true`

### Optional Schema Fields

#### `operations` (optional)
Controls which operations are allowed for units of this schema.

This is how you model **create-only** schemas (append-only ledgers) while allowing other schemas to be editable in future versions.

```yaml
operations:
  create: true
  list: true
  show: true
  update: false
  delete: false
```

Notes:
* v1 implements `create`/`list`/`show`.
* `update`/`delete` are reserved for future versions; schemas may still declare intent now.

## Property Types (v1)

Supported property types:

### `uuid`
* RFC 4122 string form
* Example: `2a7b6d13-7c2a-4e17-8c9b-9b0f4a1a0d51`

### `int`
* Base-10 signed integer
* Example: `100`, `-50`

### `bool`
* `true` or `false` (case-insensitive)
* Example: `true`, `False`

### `string`
* Text value
* Example: `"EUR"`, `"hostname"`

### `date`
* `YYYY-MM-DD` format
* Example: `"2026-01-01"`

### `timestamp`
* RFC 3339 / ISO 8601 date-time string (UTC recommended)
* Example: `"2026-01-02T12:00:00Z"`

### `ref:<SchemaName>`
* Reference to another unit's **primary ID**
* Typed by that schema
* Example: `ref:server` references a unit from the `server` schema

## Property Attributes

Each property entry supports these attributes:

### `primary`
* `true` or `false`
* **Exactly one** property **MUST** have `primary: true`
* Primary type **MUST** be `string`, `uuid`, or `int` in v1

### `required`
* `true` or `false`
* If `true`, `add` must supply it (unless it's primary uuid and auto-generated)

### `unique`
* `true` or `false`
* If `true`, value must be unique among units of that schema

### `uniqueScope` (optional)
* Array of property names: `[propA, propB, ...]`
* If present, uniqueness is enforced per distinct tuple of scope values
* Supports "natural multi-tenant" patterns where uniqueness is scoped by an organization/tenant ref

Example:
```yaml
- name: email
  type: string
  unique: true
  uniqueScope: [organizationId]
```

This means `email` must be unique within each `organizationId`, but the same email can exist in different organizations.

## Primary ID Behavior

### String Primary ID
* User **MUST** provide it
* No auto-generation
* Must be unique among units of that schema

### UUID Primary ID
* User may provide it, or
* Bus auto-generates it when omitted

### Integer Primary ID
* User **MUST** provide it
* No auto-generation

## Schema Creation

Schemas are created using:

```bash
bus schema init <schemaName> [--path <file>] [--property <spec>]...
```

See [CLI Commands Documentation](10-cli-commands.md) for details.

