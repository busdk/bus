# Secret Fields

## What this is
Rules for representing **sensitive values** in schema/unit documents while keeping workspaces Git-friendly.

## Goals (binding)
- Workspaces MUST be safe to commit/review without leaking secrets.
- Schemas MUST be able to declare fields as “secret”.
- Transports (CLI/HTTP) MUST be able to accept secret values without requiring plaintext in repo files.

## Representation (binding)

### Schema annotation
A schema MAY mark a field as secret (exact schema shape TBD), meaning:
- the value MUST NOT be stored in plaintext in workspace-authored unit documents
- the value MUST NOT be printed by default in list/show outputs

### Unit document
When a unit includes a secret field:
- the unit SHOULD store a placeholder/reference (not plaintext)
- the actual secret value is resolved at runtime via a secret provider

## Secret provider (extension point)
Secret resolution MUST be implemented behind a core-owned interface (e.g., `SecretProvider`) with built-in implementations selected at runtime.

## Non-goals
- Defining a specific secret manager integration in v1 (AWS/GCP/Vault/etc).
- Perfect redaction guarantees for arbitrary transports; default behavior should be safe.
