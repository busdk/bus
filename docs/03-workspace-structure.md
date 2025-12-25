# Workspace Structure

## User-Managed Files (Outside `.bus/`)

These files are written by Bus but managed by the user:

### `./bus.yml`
The manifest file that lists all registered schemas and their paths.

### `./<schemaName>.yml`
Default schema file location (unless `--path` is specified).

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
      <primaryId>.yml
      tx/
        <YYYY>/
          <MM>/
            index.ids
            <txId>.yml
```

### File Types

#### `.ids` Files
* **Newline-delimited identifiers**
* Merge-friendly format
* One ID per line
* Used for indexes

#### Per-Record YAML Files
* One file per unit or transaction
* Reduces merge conflicts versus monolithic files
* Filename matches the primary ID

### Directory Structure Details

#### `.bus/lock`
Advisory file lock for mutating operations.

#### `.bus/units/`
Unit storage:
* `<schemaName>.ids` - Index of all unit IDs for a schema
* `<schemaName>/<primaryId>.yml` - Individual unit files
* `<schemaName>/tx/` - Transaction storage for this schema:
  * `<YYYY>/<MM>/index.ids` - Index of transaction IDs for a month (sequential IDs: 0, 1, 2, ...)
  * `<YYYY>/<MM>/<txId>.yml` - Individual transaction files
  * Transaction IDs are sequential and continue across months (e.g., December ends with ID 42, January starts with ID 43)

## Path Resolution

* Paths in `bus.yml` are resolved relative to `.` (current working directory) unless absolute
* Schema paths stored in `bus.yml` are relative to the manifest location

