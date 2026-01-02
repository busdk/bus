# Codecs + Deterministic Formatting

## What this is
Multi-format storage rules (YAML/TOML/JSON), codec extensibility, and deterministic output requirements.

## Supported formats (binding)
- YAML: `.yml`, `.yaml`
- TOML: `.toml`
- JSON: `.json`

Bus MUST NOT introduce a custom file format.

## Format detection (binding)
Bus detects format by **file extension only**.

## Preservation on rewrite (binding)
When rewriting an existing document path, Bus MUST preserve that file’s existing format by default.

## Canonical in-memory model (binding)
All formats decode into a single canonical representation before:
- validation
- type checking
- uniqueness checks
- any feature logic

## Codec plugin architecture (extension point)

### `Codec`
Each codec provides:
- Name (`"yaml"`, `"toml"`, `"json"`, …)
- Extensions
- Decode: bytes → canonical model
- Encode: canonical model → bytes (deterministic)

### `FormatRegistry`
Resolves codecs by extension and selects defaults for new files.

## Deterministic encoding (binding)

### Comments
Comment preservation is **not required**. Rewriting YAML/TOML may drop comments.

### JSON
- Pretty-printed
- Stable key ordering
- Newline at EOF

### YAML / TOML
- Stable field/key ordering
- Stable quoting/escaping rules per codec

### `.ids` files
- One ID per line
- No blank lines
- Newline at EOF
- Deterministic ordering (recommended: lexicographic sort)

## Canonical serialization (integrity hook)
If hashing/signing is used, Bus MUST define canonical bytes independent of YAML/TOML/JSON source bytes.


