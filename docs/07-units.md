# Unit Instance Storage

Units are instances of schemas. Each unit is stored as a separate YAML file.

## Storage Locations

### Index File
`.bus/units/<schemaName>.ids`
* Newline-delimited list of unit primary IDs
* One ID per line
* Used for fast lookup

### Unit File
`.bus/units/<schemaName>/<primaryId>.yml`
* Individual unit record
* Filename matches the primary ID

## Unit File Format

```yaml
kind: bus.unit
version: 1
schema: user
data:
  id: 2a7b6d13-7c2a-4e17-8c9b-9b0f4a1a0d51
  serverId: 10
  contractStart: "2026-01-01"
  contractEnd: null
  monthlyFeeCents: 100
  currency: "EUR"
```

### Required Fields

#### `kind`
* **MUST** be `bus.unit`

#### `version`
* **MUST** be `1` in v1

#### `schema`
* **MUST** match a schema registered in `bus.yml`
* Identifies which schema this unit conforms to

#### `data`
* Object containing all unit properties
* **MUST** satisfy schema constraints:
  * Required properties present (after auto-generation rules)
  * Types match schema definitions
  * Uniqueness constraints satisfied

## Primary ID Behavior

### String Primary ID
* User **MUST** provide it
* No auto-generation
* Must be unique among units of that schema

### UUID Primary ID
* User may provide it, or
* Bus auto-generates it when omitted
* Auto-generated UUIDs are RFC 4122 compliant

### Integer Primary ID
* User **MUST** provide it
* No auto-generation

### File Naming
* The unit file name and `.ids` entry **MUST** match the primary ID string form
* For strings: use the string value as-is
* For UUIDs: use the full UUID string
* For integers: use the integer as a string

## Unit Operations

### Create Unit
```bash
bus <SCHEMA_NAME> add key=value key2=value2 ...
```

Example:
```bash
bus user add serverId=10 contractStart=2026-01-01 monthlyFeeCents=100 currency=EUR
```

Rules:
* After flags, every arg containing `=` is parsed as `key=value` (split on first `=`)
* Duplicate keys are an error
* Unknown keys (not in schema) are an error
* Required keys missing are an error (except primary uuid which can be auto-generated)

### List Units
```bash
bus <SCHEMA_NAME> list
```

Lists all unit IDs for the schema.

### Show Unit
```bash
bus <SCHEMA_NAME> show <primaryId>
```

Displays the full unit record.

## Data Validation

When creating or updating units, Bus validates:
1. All required properties are present (after auto-generation)
2. Property types match schema definitions
3. Unique constraints are satisfied
4. Reference properties point to valid units
5. Unique scope constraints are satisfied (if applicable)

## Storage Benefits

### Merge-Friendly
* One file per unit reduces merge conflicts
* `.ids` files are newline-delimited (easy to merge)
* Per-record files make Git diffs clearer

### Atomic Operations
* Unit creation writes the unit file first
* Then updates the `.ids` index
* If any step fails, no partial state remains

