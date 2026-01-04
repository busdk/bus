# Ledger / Transactions Pattern (Append‑Only)

## What this is
How Bus models “transactions” without a special transaction subsystem: transactions are units in a schema you define, treated as create-only.

## Binding rules
- A “transaction” is a unit (commonly in schema named `transaction`).
- Ledger-like records are append-only: corrections are new records, not edits.
- Bus does not ship a billing engine or billing rule interpreter in core.

## Idempotency guidance
If transactions are generated externally and re-run, prefer deterministic primary ids (often string ids) so duplicates are naturally prevented by uniqueness.


