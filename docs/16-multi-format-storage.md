# Multi-Format Storage (v1)

Bus is **file-first** and **human-editable**. In v1, Bus supports **YAML**, **TOML**, and **JSON** as interchangeable encodings for the same logical document models.

This page specifies:
- Which documents may be stored in which formats
- How formats are detected and preserved
- Deterministic discovery/conflict rules
- A minimal codec/registry architecture that is extensible to new formats without rewriting unrelated functionality

## Supported Formats (v1)

Bus v1 supports:
- **YAML**: `.yml`, `.yaml`
- **TOML**: `.toml`
- **JSON**: `.json`

Bus MUST NOT introduce a new custom file format.

## Scope: Which Documents Can Be Multi-Format

In v1, the following **workspace documents** MAY be stored in any supported format:
- **Manifest**: `bus.yml` / `bus.yaml` / `bus.toml` / `bus.json`
- **Schemas**: `<schemaName>.yml|.yaml|.toml|.json`
- **Units**: per-record files under `.bus/` (e.g., `.bus/units/<schema>/<id>.<ext>`)
- **Transaction ledger records** (stored as units under `.bus/units/transaction/...`): same rules as units

`.ids` files remain newline-delimited text and are not multi-format.

## Format Detection and Preservation

### Detection

Bus MUST detect the document format by **file extension** only.

### Preservation on Write (Default)

When Bus writes an existing document path, it MUST preserve the document’s **existing format** by default:
- If a file is `*.toml`, Bus writes TOML back to that same path.
- If a file is `*.json`, Bus writes JSON back to that same path.
- If a file is `*.yml`/`*.yaml`, Bus writes YAML back to that same path.

Bus MUST NOT “convert” documents between formats unless explicitly asked via a future command (out of scope for v1).

## Discovery and Conflict Rules

Bus must be deterministic and safe when multiple candidates exist.

### Manifest Discovery

In a working directory, manifest candidates are:
- `bus.yml`
- `bus.yaml`
- `bus.toml`
- `bus.json`

Rules:
- If **exactly one** candidate exists: load it.
- If **none** exist: commands that require a manifest MUST error; `bus init` MAY create one (see below).
- If **more than one** exist: Bus MUST error and list the candidates with a clear resolution message (e.g., “delete/rename extras or pass an explicit path flag if/when supported”).

### Schema Discovery by Name

For a schema name `<schemaName>`, schema candidates are:
- `<schemaName>.yml`, `<schemaName>.yaml`, `<schemaName>.toml`, `<schemaName>.json`

Rules:
- If multiple candidates exist for the same `<schemaName>`, Bus MUST error (ambiguous schema definition).
- If a schema is registered in the manifest with a `path`, that explicit path wins and the extension determines the codec.

### Unit / Data Records

Unit record files are created and managed under `.bus/`. Mixing formats inside one workspace is allowed.

Recommended deterministic behavior:
- The **manifest format** defines the default format for newly created records (schemas and unit records) unless overridden by an explicit `--format` flag for create/init-style commands.
- Referenced files may be any supported format; the codec is chosen by that file’s extension.

## Default Format for Newly Created Files

### Default Selection

When Bus creates a new file and no explicit `--format` is provided:
- If a manifest exists: default to the manifest’s format (by extension).
- Otherwise: default to **YAML**.

### `--format` (Spec-Only)

Create/init-style commands MUST support a `--format` option (spec-only for v1) to select the format used for files that command creates.

Behavior:
- `--format yaml` produces `.yml` (or `.yaml`, but Bus MUST pick one deterministically; recommended: `.yml` in v1).
- `--format toml` produces `.toml`.
- `--format json` produces `.json`.

If an explicit `--path` is provided that already has a recognized extension, the extension wins; `--format` is ignored (or errors) to avoid surprising behavior. Bus MUST document whichever rule it chooses and apply it consistently.

## Canonical Internal Model

Bus MUST parse any supported format into a **single canonical in-memory representation** before:
- Schema validation
- Type checking
- Uniqueness checks
- Any business logic (including micropayment capture/reporting)

The canonical model is logical (maps/arrays/scalars) and independent of the input format.

## Codec Plugin Architecture (Minimal)

Adding a new format later MUST require only implementing and registering a new codec. No unrelated business logic should change.

### `Codec` Interface

Each codec MUST provide:
- **Name**: stable identifier (`"yaml"`, `"toml"`, `"json"`, …)
- **Extensions**: recognized extensions (e.g., `[".yml", ".yaml"]`)
- **Decode**: bytes → canonical document model (and/or typed structs)
- **Encode**: canonical document model (and/or typed structs) → bytes (deterministic formatting)

### `FormatRegistry`

Bus MUST have a single registry responsible for:
- Resolving codecs by file extension
- Routing loads/saves for manifest/schema/unit documents through the correct codec
- Providing a “default format” decision when creating new files

The rest of the system MUST NOT branch on format.

## Round-Trip and Stability Expectations (v1)

### Comment Preservation

Comment and exact-whitespace preservation is **not required** in v1.

Bus MUST document this clearly because re-writing YAML/TOML may drop comments.

### Deterministic Encoding

Bus MUST emit deterministic, readable encodings:
- **JSON**: pretty-printed, stable key ordering, newline at EOF
- **TOML**: stable key ordering; avoid emitter features that reorder unpredictably
- **YAML**: stable field ordering and quoting rules (see `12-formatting.md`)

### Spec-Only Tests (Implementation Guidance)

Implementation SHOULD include:
- **Parse-equivalence tests**: same logical content in YAML/TOML/JSON decodes to the same canonical representation
- **Deterministic encoding tests**: encoding the same canonical input yields byte-for-byte stable output per codec


