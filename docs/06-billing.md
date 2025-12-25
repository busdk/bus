# Billing (Schema Pattern)

Bus v1 does **not** ship a built-in billing engine, contract interpreter, or special transaction subsystem.

If you want “billing”, you model it using the same primitives as everything else:
* **Schemas** define structure and constraints
* **Units** store state and events
* **Typed references** (`ref:<schema>`) encode relationships

This keeps v1 simpler: Bus is a schema-driven unit store. Billing is a **domain model you define**.

## Minimal Model: `Account` + `Transaction`

The simplest practical billing/ledger model is:
* `account` units: “what we track balances for”
* `transaction` units: “append-only postings affecting an account”

### Currency as a Unit (e.g., “EUR-only”)

Instead of hard-coding currency lists or special handling, define a `currency` schema and reference it.

If you want an **EUR-only** system:
* Create exactly one `currency` unit with primary ID `EUR`
* Require all accounts/transactions to reference `currencyId: EUR`

Schema:

```yaml
kind: bus.schema
version: 1
name: currency
properties:
  - name: id
    type: string
    primary: true
    required: true
    unique: true
  - name: decimals
    type: int
    required: true
```

Unit:

```yaml
kind: bus.unit
version: 1
schema: currency
data:
  id: EUR
  decimals: 2
```

### Account Schema

```yaml
kind: bus.schema
version: 1
name: account
properties:
  - name: id
    type: uuid
    primary: true
    required: true
    unique: true
  - name: name
    type: string
    required: true
  - name: currencyId
    type: ref:currency
    required: true
```

### Transaction Schema (Append-Only Posting)

Transactions are just units. To implement an append-only ledger, treat `transaction` units as **create-only** records.

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
* `deltaCents` can be positive or negative (no extra “debit/credit” field required).
* `currencyId` is a typed reference. If you only have `EUR`, every transaction points to `EUR`.

## Writing Data (No Special Billing Commands)

Because `transaction` is an ordinary schema, you create entries with the normal unit command:

```bash
bus transaction add accountId=<ACCOUNT_UUID> postedAt=2026-01-01 deltaCents=1000 currencyId=EUR note="January fee"
```

## Where “Recurring Billing” Lives

Recurring billing is an **application concern**, not a Bus v1 feature:
* You can generate `transaction` units from any rules you want (cron job, script, CI, etc.)
* Bus only needs to validate that the resulting unit matches the schema and constraints

If you want idempotency, you can also choose a deterministic primary ID strategy (e.g., use a string primary ID like `acct=<id>|month=YYYY-MM|type=fee`) so “re-running generation” doesn’t create duplicates.
