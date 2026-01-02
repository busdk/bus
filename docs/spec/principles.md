# Principles (Non‑Negotiable)

## What this is
The binding architectural principles that every roadmap step must preserve.

## Principles

### Modular-by-default
Every feature is a module. Core does not accrete feature logic.

### Small, stable core
Bus Core exists to orchestrate modules. Adding a feature should primarily mean:
- implement an interface
- wire/register a built-in implementation

It should not mean editing core routing logic.

### Append-only accounting history
Ledger-like records are create-only. Corrections are new records (reversals/adjustments), never edits.

### Multi-tenant from day one
All operations are tenant-scoped. Storage and queries must not allow cross-tenant reads/writes by default.

### Interfaces first
Every capability starts as a core-owned interface (or small interface set). Modules provide implementations.

## Constraints (v1 intent preserved)

### No Git operations
Bus MUST NOT perform Git operations (init/commit/merge/etc). Git is outside the tool.

### No tool-defined top-level directory hierarchy
Bus MUST NOT require directories like `schemas/` or `units/` at the top level.

### Only `.bus/` is controlled structure
When the **filesystem state backend** is selected, all Bus-owned mutable state lives under `.bus/`.

When a **database state backend** (or other non-filesystem backend) is selected, internal mutable state lives in that backend and `.bus/` is not required.


