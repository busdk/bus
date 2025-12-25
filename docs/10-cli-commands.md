# CLI Commands

## Command Routing Rules

Resolution order for `bus <t0> ...`:

1. If `t0` is a **built-in command**, run it
2. Else if `t0` matches a schema in `bus.yml.units[].name`, treat `t0` as schema command
3. Else error

### Built-in Commands (v1)

* `init` - Initialize workspace
* `schema` - Schema management
* `unit` - Compatibility prefix for unit commands

### Schema Name Collisions

If a schema name collides with a built-in, built-in wins. Schema access requires:

```bash
bus unit <schemaName> ...
```

## Init Command

### `bus init`

Creates `./bus.yml` only.

```bash
bus init
```

**Writes:**
```yaml
kind: bus.manifest
version: 1
units: []
```

**Also creates:**
* `./.bus/` directory (empty except maybe `.bus/lock` created lazily)

## Schema Commands

### `bus schema init`

Creates a schema file and registers it into `bus.yml`.

```bash
bus schema init <schemaName> [--path <file>] [--property <spec>]...
```

**Defaults:**
* If `--path` omitted: `./<schemaName>.yml`

**Property spec format (v1):**
* `--property name:type[,attr[,attr...]]`
* Attributes: `primary`, `required`, `unique`
* Optional: `uniqueScope=<a|b|c>` (pipe-separated)

**Example:**
```bash
bus schema init server \
  --property id:int,primary,required,unique \
  --property hostname:string,required,unique
```

**Rules:**
* Exactly one `primary` property
* Primary type must be `string`, `uuid`, or `int`

**On success:**
* Writes schema YAML
* Updates `bus.yml` (atomic with schema write)

## Unit Commands

### Preferred Form

```bash
bus <SCHEMA_NAME> add key=value key2=value2 ...
bus <SCHEMA_NAME> list
bus <SCHEMA_NAME> show <primaryId>
```

### Compatibility Form

```bash
bus [unit] <SCHEMA_NAME> add ...
bus [unit] <SCHEMA_NAME> list
bus [unit] <SCHEMA_NAME> show ...
```

### `add` Command

**Key=Value Parsing:**
* After flags, every arg containing `=` is parsed as `key=value` (split on first `=`)
* Duplicate keys are an error
* Unknown keys (not in schema) are an error
* Required keys missing are an error (except primary uuid which can be auto-generated)

**Example:**
```bash
bus account add name="Main" currencyId=EUR
```

**Output:**
* Prints the created primary ID (so scripts can capture it)

### `list` Command

Lists all unit IDs for a schema.

```bash
bus <SCHEMA_NAME> list
```

### `show` Command

Displays a unit record.

```bash
bus <SCHEMA_NAME> show <primaryId>
```

## Transactions and “Billing”

Bus v1 does not include a built-in transaction or billing subsystem.

If you want a ledger, define a schema (conventionally named `transaction`) and use normal unit commands:

```bash
bus transaction add ...
bus transaction list
bus transaction show <id>
```

See [Billing (Schema Pattern)](06-billing.md).

