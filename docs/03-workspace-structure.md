# Workspace Structure

## User-Managed Files (Outside `.bus/`)

These files are written by Bus but managed by the user:

### `./bus.{yml,yaml,toml,json}`
The manifest file that lists all registered schemas and their paths.

### `./<schemaName>.<ext>`
Default schema file location (unless `--path` is specified), where `<ext>` is one of:
`yml`, `yaml`, `toml`, `json`.

Bus writes these files but does not impose a directory structure for them.

## Bus-Managed Files (Inside `.bus/`)

Bus owns all files under `.bus/`:

* Indexes for fast lookup and conflict-friendly lists
* Per-record files for units and transactions

## Recommended v1 Structure

All files inside `.bus/`:

```
.bus/
  lock
  units/
    <schemaName>.ids
    <schemaName>/
      <primaryId>.<ext>
```

### File Types

#### `.ids` Files
* **Newline-delimited identifiers**
* Merge-friendly format
* One ID per line
* Used for indexes

#### Per-Record Document Files (YAML/TOML/JSON)
* One file per unit or transaction
* Reduces merge conflicts versus monolithic files
* Filename matches the primary ID

`<ext>` is chosen deterministically (see `16-multi-format-storage.md`):
- Default to the manifest’s format
- Preserve existing format on rewrite

### Directory Structure Details

#### `.bus/lock`
Advisory file lock for mutating operations.

#### `.bus/units/`
Unit storage:
* `<schemaName>.ids` - Index of all unit IDs for a schema
* `<schemaName>/<primaryId>.<ext>` - Individual unit files

Transactions are stored like any other unit under their schema (see `09-transactions.md`). Bus v1 does not maintain special derived indexes beyond `.ids`.

## Path Resolution

* Paths in the manifest are resolved relative to `.` (current working directory) unless absolute
* Schema paths stored in the manifest are relative to the manifest location

