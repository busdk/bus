# YAML Codec

## What this is
The YAML codec behavior and deterministic encoding requirements.

## Extensions
- `.yml` (recommended default)
- `.yaml`

## Deterministic encoding (binding)
- Stable field ordering
- Stable quoting/escaping rules
- Stable boolean/null representation
- Newline at EOF

## Comment preservation
Not required. Rewriting YAML may drop comments.
