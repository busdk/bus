# CLI Commands

## Command Routing Rules

Resolution order for `bus <t0> ...`:

1. If `t0` is a **built-in command**, run it
2. Else if `t0` matches a schema in the manifest `units[].name`, treat `t0` as schema command
3. Else error

### Built-in Commands (v1)

* `init` - Initialize workspace
* `schema` - Schema management
* `unit` - Compatibility prefix for unit commands
* `micropayments` - Reporting over transactions (micropayment-oriented views)
* `x402` - x402 requirement generation and ingestion

### Schema Name Collisions

If a schema name collides with a built-in, built-in wins. Schema access requires:

```bash
bus unit <schemaName> ...
```

## Format Selection (Multi-Format Workspaces)

Bus supports YAML/TOML/JSON for workspace documents (see `16-multi-format-storage.md`).

Create/init-style commands MUST support a `--format` option (spec-only) that controls the format of newly created files.

## Init Command

### `bus init`

Creates a manifest in the current directory.

Defaults:
- If `--format` is not provided: creates `./bus.yml` (YAML).
- If `--format` is provided: creates `./bus.<ext>` where `<ext>` matches the selected format.

```bash
bus init [--format yaml|toml|json]
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

Creates a schema file and registers it into the manifest.

```bash
bus schema init <schemaName> [--path <file>] [--format yaml|toml|json] [--property <spec>]...
```

**Defaults:**
* If `--path` omitted: `./<schemaName>.<ext>` where `<ext>` is the workspace default format (manifest format if present; otherwise YAML)

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
* Writes the schema document (YAML/TOML/JSON)
* Updates the manifest (atomic with schema write)

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

## Micropayments and x402 (v1)

Micropayments in Bus are modeled as **transactions** between business units/services (see `17-micropayments.md`).

x402 is one capture mechanism (see `18-x402.md`).

### `bus x402 402`

Generates the HTTP 402 “payment required” JSON body for a service from workspace config.

```bash
bus x402 402 --service <SERVICE_ID>
```

Output:
- Writes a JSON object to stdout (the 402 body).

### `bus x402 ingest`

Ingests x402 proof headers and logs a normalized transaction record.

```bash
bus x402 ingest \
  --service <SERVICE_ID> \
  --x-payment <BASE64_JSON> \
  [--x-payment-response <BASE64_JSON>]
```

Rules:
- `--x-payment` is required.
- Bus MUST enforce uniqueness constraints to prevent double booking (see `17-micropayments.md`).

### `bus micropayments report`

Reports over normalized transactions (micropayment-oriented view).

```bash
bus micropayments report [--from <ts>] [--to <ts>] [--group-by <dim>]...
```

Spec-only behavior:
- Filters and groups over `transaction` records with micropayment fields (amount/asset/from/to/source, etc.)
- Outputs deterministic, script-friendly text or JSON (future flag; out of scope for v1)

## Local vs Remote Mode (Future; Hosted Facilitator)

v1 is local-only. Future versions may support a hosted facilitator (HTTP server) where the CLI calls remote APIs instead of local files.

Spec-only requirement:
- CLI commands SHOULD accept `--remote <URL>` in the future.
- When `--remote` is set, the CLI MUST invoke the same logical operations over HTTP without changing command semantics.
- The hosted facilitator MUST publish an **OpenAPI** specification so remote client libraries can be automatically generated.

