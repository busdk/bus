# Ledger / Transactions (Schema Pattern)

Bus v1 does **not** have a special transaction subsystem.

A “transaction” is simply a **unit** in a schema you define (often named `transaction`) that you treat as **append-only**.

## Design Principles

### Append-Only (Create-Only)
* Transaction units are immutable records.
* Corrections are **new** transactions (reversals/adjustments), not edits.
* You can encode this intent directly in the schema via `operations` (see [Schemas](05-schemas.md)).

### Schema-Defined
* You define the structure, references, and constraints in YAML.
* Bus enforces schema validation and uniqueness, but does not interpret “billing rules”.

## Recommended Minimal Schema

The exact fields are your choice. A common minimal pattern is “posting to an account”:

```yaml
kind: bus.schema
version: 1
name: transaction
operations:
  create: true
  list: true
  show: true
  update: false
  delete: false
properties:
  - name: id
    type: uuid
    primary: true
    required: true
    unique: true
  - name: accountId
    type: ref:account
    required: true
  - name: postedAt
    type: date
    required: true
  - name: deltaCents
    type: int
    required: true
  - name: currencyId
    type: ref:currency
    required: true
  - name: note
    type: string
    required: false
```

Notes:
* A single signed integer (`deltaCents`) keeps the model small.
* Using `ref:currency` avoids hard-coded currency handling. “EUR-only” can be modeled by having only one currency unit (`id: EUR`).

## Storage

Transaction units are stored like any other unit:
* Index: `.bus/units/transaction.ids`
* Records: `.bus/units/transaction/<id>.yml`

Bus does not maintain special month indexes in v1. If you want derived indexes or reports, build them outside Bus from the unit files.

## Optional: Stable / Deterministic IDs

If you generate transactions from an external process and want idempotency, consider using a **string primary ID** for the `transaction` schema:
* Example ID format: `acct=<accountId>|month=YYYY-MM|kind=fee`
* Then re-running generation naturally overwrites nothing (duplicate creates fail due to uniqueness).

