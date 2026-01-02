# Consistency Rules and Atomic Operations

Bus ensures data consistency through atomic operations and workspace locking.

## No Partial Writes

Any mutating command must either:
* Complete successfully, or
* Make no persistent changes

No command should leave the repository in an inconsistent state.

## Workspace Lock

Mutating commands must acquire a lock:
* `.bus/lock` (advisory file lock)
* Lock is acquired before any writes
* Lock is released on exit (including error exits)

### Lock Behavior
* Only one mutating operation can run at a time
* Read-only operations may proceed without locks
* Lock prevents concurrent modifications

## Atomic Write Strategy (v1)

### Single File Writes

For each file write:
1. Write to `.<name>.tmp` in the same directory
2. `fsync` file (recommended for durability)
3. `os.Rename(tmp, final)` (atomic on same filesystem)

This ensures:
* No partial file writes are visible
* Atomic replacement on Unix-like systems
* File system guarantees the rename is atomic

### Multi-File Operations

For operations that write multiple files (schema init, unit add, tx update):

1. **Prepare all new contents in memory**
   * Validate all data before writing
   * Compute all file paths and contents

2. **Write all "new files" first** (using atomic rename)
   * Unit files
   * Transaction files
   * Schema files

3. **Then rewrite affected `.ids` files and the manifest last** (also atomic rename)
   * Index files are the "commit point"
   * Manifest is updated last

4. **If any step fails:**
   * Delete any newly created per-record files from this operation
   * Do not modify `.ids`/manifest (because those are written last)

### Why This Works

* Index files are written last, so they only exist if all data files exist
* If an operation fails, indexes remain unchanged
* New data files without index entries are "orphaned" but can be cleaned up
* The manifest is updated last, so schema registration only happens if everything succeeds

## Example: Unit Creation

When creating a unit:

1. **Acquire lock** (`.bus/lock`)
2. **Validate input** (check schema, uniqueness, etc.)
3. **Generate primary ID** (if needed)
4. **Write unit file** (`.bus/units/<schema>/<id>.<ext>.tmp` → rename)
5. **Update index** (`.bus/units/<schema>.ids.tmp` → rename)
6. **Release lock**

If step 4 fails: no changes made
If step 5 fails: unit file exists but not in index (can be cleaned up)

## Example: Schema Init

When creating a schema:

1. **Acquire lock**
2. **Validate schema** (check properties, primary ID, etc.)
3. **Write schema file** (`<schemaName>.<ext>.tmp` → rename)
4. **Update manifest** (`bus.<ext>.tmp` → rename)
5. **Release lock**

If step 3 fails: no changes made
If step 4 fails: schema file exists but not registered (can be cleaned up)

## Recovery

If an operation fails partway through:
* Orphaned files (data files without index entries) can be detected
* Future versions may include a `bus repair` command
* For now, manual cleanup may be needed

## Git-Friendly Design

The atomic write strategy works well with Git:
* Git sees complete file replacements (not partial writes)
* Merge conflicts are minimized by per-record files
* Index files are simple newline-delimited lists (easy to merge)

