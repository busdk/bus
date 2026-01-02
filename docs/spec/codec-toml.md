# TOML Codec

## What this is
The TOML codec behavior and deterministic encoding requirements.

## Extension
- `.toml`

## Deterministic encoding (binding)
- Stable key ordering within tables
- Avoid emitter behavior that reorders keys unpredictably
- Newline at EOF

## Comment preservation
Not required. Rewriting TOML may drop comments.
