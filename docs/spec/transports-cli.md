# Transports (CLI First)

## What this is
How external interfaces (CLI now, HTTP later) invoke module capabilities without embedding business logic in the transport or in core.

## Request context (cross-cutting)
Every capability invocation MUST carry a `RequestContext` containing:
- `TenantID`
- `Actor` (opaque; for authz hooks)
- `RequestID` (for audit hooks)

## CLI command capability

### `Command`
Responsibility: a script-friendly operation exposed via CLI.

Minimum shape:
- `Name() string` — stable name (e.g., `"schema.init"`, `"units.add"`).
- `Synopsis() string`
- `Run(ctx RequestContext, args []string) (Result, error)`

## Command routing (binding)
When a user runs `bus <t0> ...`, the CLI resolves in this order:
1. If `t0` matches a built-in command capability, run it.
2. Else if `t0` matches a registered schema name, route to schema-unit commands.
3. Else error.

Collision rule:
- Built-in wins. Schema commands remain reachable via `bus unit <schemaName> ...`.

## Script-friendly parsing (binding)
Unit creation input uses positional `key=value` tokens:
- After flags, every arg containing `=` is parsed as `key=value` (split on first `=`).
- Duplicate keys are an error.
- Unknown keys (not in schema) are an error.


