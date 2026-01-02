# Deterministic Formatting Requirements

To be Git-friendly and reduce conflicts, Bus must emit files in a deterministic format.

Bus v1 supports multiple document formats (YAML/TOML/JSON). Each supported codec MUST produce stable, readable output (see `16-multi-format-storage.md`).

## Comments and Round-Trip

Comment preservation is **not required** in v1.

If Bus rewrites a YAML/TOML file, comments may be lost. This MUST be documented clearly by the tool and reflected in user expectations.

## YAML Emission

YAML emission MUST be deterministic:

### Stable Field Ordering
* Fields must always appear in the same order
* Recommended order:
  1. `kind`
  2. `version`
  3. Schema-specific fields (alphabetically)
  4. `data` (for units)
  5. Nested structures (alphabetically)

### Stable Quoting Rules
* Strings that need quoting (contain special chars, start with numbers, etc.) must always be quoted
* Strings that don't need quoting should never be quoted
* Consistent boolean representation (`true`/`false`, not `True`/`False`)
* Consistent null representation (`null`, not `~` or empty)

### Avoid Timestamps in Deterministic Records
* Avoid writing “current time” into **deterministically generated** records where re-running should produce identical bytes
* For captured events (e.g., transaction capture), timestamps like `createdAt` are expected and are part of the event record

## JSON Emission

JSON emission MUST be deterministic:
- Pretty-printed (stable indentation)
- Stable key ordering (deterministic map ordering)
- Newline at EOF

## TOML Emission

TOML emission MUST be deterministic:
- Stable key ordering within tables
- Avoid emitter behavior that reorders keys unpredictably
- Prefer simple TOML constructs that round-trip cleanly to the canonical model

## `.ids` Files

`.ids` files must be:

### Sorted
* Sorted lexicographically (optional but recommended)
* Makes merges easier
* Makes diffs clearer

### Newline at EOF
* Every `.ids` file ends with a newline
* Consistent with POSIX text file standards

### No Blank Lines
* No empty lines in `.ids` files
* One ID per line
* No trailing whitespace

### Format Example
```
00000000-0000-0000-0000-000000000001
00000000-0000-0000-0000-000000000002
00000000-0000-0000-0000-000000000003
```

## Benefits

### Git-Friendly
* Deterministic output means same inputs produce same outputs
* Reduces unnecessary diffs
* Makes merges more predictable

### Merge-Friendly
* Per-record files reduce merge conflicts
* Sorted `.ids` files make three-way merges easier
* Consistent formatting makes conflicts easier to resolve

### Reproducible
* Same data always produces same files
* Important for testing
* Important for verification

## Implementation Notes

### Codec Libraries

Use libraries/emitters that support deterministic output for each codec:
- Stable key ordering
- Stable quoting/escaping rules
- Control over indentation/newlines

### Sorting
* Sort IDs before writing to `.ids` files
* Sort properties in unit data (alphabetically)
* Sort nested structures consistently

### Testing
* Test that same inputs produce identical outputs
* Test parse-equivalence across YAML/TOML/JSON for the same logical content
* Test that formatting is stable across runs
* Test that merges work correctly

## Canonical Serialization (Integrity)

If Bus computes hashes or signatures for integrity/verification, it MUST define a canonical serialization for the payload being hashed/signed.

Canonicalization MUST be independent of the source file format to avoid “same data, different bytes” problems across YAML/TOML/JSON.

