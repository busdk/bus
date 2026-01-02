# Micropayments (Module)

## What this is
Micropayments are a use-case framing for small, auditable transactions. The underlying record remains a **transaction** (append-only ledger entry).

## Scope (v1 intent preserved)
Bus is not a “payment processor”:
- in scope: capture + deterministic storage + reporting/aggregation
- out of scope: settlement/invoicing automation, blockchain network integration

## Normalized transaction shape (binding for micropayments module)
The micropayments module defines a normalized `Transaction` shape for reporting/capture sources, typically stored as unit data in a `transaction` schema:
- `id` (primary)
- `createdAt` (RFC 3339 timestamp)
- `fromUnitId`, `toUnitId`
- `serviceId` (optional)
- `amount` (string; base units)
- `asset` (string)
- `network` (optional string)
- `sourceType` (string: `"x402" | "manual" | "import" | ...`)
- `sourceRef` (optional string)
- `proofHash` or `rawProof` (optional)
- optional integrity hooks (hash/signature/nonce/validity windows)

## Uniqueness / double-booking prevention (binding)
At minimum:
- uniqueness on `(sourceType, sourceRef)` when `sourceRef` is present

For x402:
- uniqueness on `(network, transactionId)` when an x402 success header provides a transaction id

Note:
- If schema constraints cannot express this directly, a module-owned index under `.bus/` is allowed (without changing core).


