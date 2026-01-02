# Non-Negotiable Constraints

These constraints are fundamental to Bus CLI v1 and cannot be changed.

## 1. No Git Operations

**Bus MUST NOT** perform Git operations:
* Must not initialize a Git repository
* Must not commit changes
* Must not merge branches
* Must not perform any Git operations

Git is considered "outside" the tool. Users manage Git themselves.

## 2. No Tool-Defined Top-Level Directory Hierarchy

**Bus MUST NOT** create or rely on directories like `schemas/` or `units/` at the top level.

* By default, schema files are adjacent to the **manifest** in the current working directory
* Users may choose another structure via `--path` flags
* Bus stores those paths in the manifest (`bus.{yml,yaml,toml,json}`)

## 3. Only `.bus/` is Controlled Structure

Bus may create `.bus/` and subdirectories under it for its own state:
* Indexes
* Object files

All Bus-managed mutable state lives under `.bus/`.

## 4. Manifest is in Current Working Directory

For v1, commands operate on the current directory only:
* `bus init` creates a manifest in `./` (default: `./bus.yml`)
* Other commands require exactly one manifest candidate to exist:
  * `bus.yml`, `bus.yaml`, `bus.toml`, or `bus.json`

## 5. Schema-Driven Units with Primary ID

Each unit schema defines:
* All properties, including the primary ID property
* Primary ID property declares:
  * Which property is primary (`primary: true`)
  * Its type (string, UUID, or integer)
  * Uniqueness, requiredness, etc.

## 6. Unit Creation Uses `key=value`

Unit properties must be passed as positional `key=value` tokens:
* No `--set` flags
* No `--field` flags
* Simple `key=value` format

## 7. Relations and Ledgers are Schema Patterns

v1 includes:
* **Basic relations** between entities (typed references)
* A **schema-driven unit store** (units are the only primitive data record)

Bus v1 does **not** include built-in billing automation rules (scheduled charging, auto-invoicing) or a special “billing engine”.

If you want a general-purpose append-only ledger, define your own `transaction` schema and treat it as **create-only** (corrections are new transactions, not edits).

Micropayments are supported as a first-class **transaction ledger** for capture and reporting (settlement is future). See `17-micropayments.md` and `18-x402.md`.

