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

## Reporting / aggregation (binding)
The “micropayments report” is a **pure, deterministic aggregation** over the normalized transactions that already exist as units (append-only ledger entries). It is intentionally designed to be implementable *on top of Bus* without introducing a “payments subsystem”.

### Inputs
- **time window**: optional `from` / `to` timestamps (RFC 3339)
- **filters**: optional constraints like `asset`, `fromUnitId`, `toUnitId`, `serviceId`, `sourceType`, `network`
- **group-by dimensions** (one or more):
  - `fromUnitId`
  - `toUnitId`
  - `serviceId` (optional field)
  - `asset`
  - `sourceType`
  - `network` (optional field)
  - **time bucket** derived from `createdAt`: `hour | day | month` (exact bucket names/format are an output contract)

### Aggregates (recommended minimum)
Each output row SHOULD include:
- `count` (number of transactions in the group)
- `amountSum` (sum of `amount` in base units, as a string)
- `firstAt` / `lastAt` (min/max of `createdAt` in the group)

### Determinism requirements
Given the same set of input transactions:
- grouping MUST be stable and independent of input iteration order
- output rows MUST have a stable ordering (e.g., lexicographic order by group key tuple)
- numeric addition MUST be exact for the `amount` representation chosen (do not use float)

## Examples: what the report should produce
The examples below show the *shape* and *intent* of the aggregation. The exact transport format (JSON/TOML/table) is a CLI concern; the core operation should return a deterministic structured result.

### Example A — daily gross received per payee (toUnitId) and asset
Given these normalized transactions (base units):

| id | createdAt | fromUnitId | toUnitId | amount | asset |
| --- | --- | --- | --- | --- | --- |
| tx_001 | 2026-01-01T10:05:00Z | u_alice | u_api | 10 | USDc |
| tx_002 | 2026-01-01T10:06:00Z | u_alice | u_api | 15 | USDc |
| tx_003 | 2026-01-01T18:02:00Z | u_bob | u_api |  5 | USDc |
| tx_004 | 2026-01-02T09:12:00Z | u_alice | u_api |  2 | USDc |
| tx_005 | 2026-01-02T09:13:00Z | u_bob | u_api |  1 | USDc |
| tx_006 | 2026-01-02T09:14:00Z | u_bob | u_cdn |  3 | USDc |

Request:
- `groupBy = [day(createdAt), toUnitId, asset]`
- `from = 2026-01-01T00:00:00Z`, `to = 2026-01-03T00:00:00Z`

Result rows (illustrative):

| day | toUnitId | asset | count | amountSum | firstAt | lastAt |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-01-01 | u_api | USDc | 3 | 30 | 2026-01-01T10:05:00Z | 2026-01-01T18:02:00Z |
| 2026-01-02 | u_api | USDc | 2 |  3 | 2026-01-02T09:12:00Z | 2026-01-02T09:13:00Z |
| 2026-01-02 | u_cdn | USDc | 1 |  3 | 2026-01-02T09:14:00Z | 2026-01-02T09:14:00Z |

This supports “how much did each provider receive per day?” without any invoicing engine.

### Example B — per-customer cost by service (usage-based pricing)
If `serviceId` is recorded, group by payer + service:
- `groupBy = [month(createdAt), fromUnitId, serviceId, asset]`

This produces rows suitable for a customer-facing “usage summary”:
- “In Jan 2026, u_alice spent 27 USDc on service `svc.api.calls`.”
- “In Jan 2026, u_bob spent 6 USDc on service `svc.cdn.egress`.”

### Example C — source-based audit (imported vs x402 vs manual)
Group by payment source to answer “where did these entries come from?”:
- `groupBy = [day(createdAt), sourceType]`

This produces an operator-audit view:
- count and amountSum per day per source type (e.g., `x402` vs `import`), helping detect ingestion bugs or double-booking.


