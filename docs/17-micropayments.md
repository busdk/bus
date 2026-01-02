# Micropayments (v1): Transactions Between Business Units

In Bus, a **micropayment** is a small, auditable **transaction** between business units/services that can later be **settled** (e.g., netted and converted into balancing invoices or accounting exports).

Terminology:
- **Transaction**: the core ledger record between units.
- **Micropayment**: a use-case framing for small transactions. In Bus documentation, we use “micropayment” mainly when describing capture/settlement context; the underlying record remains a **transaction**.

Bus v1 is **not** a crypto “payment processor”. It is a **micropayment collection + accounting-oriented ledger**:
- capture transactions (from multiple sources)
- validate and store them deterministically
- support reporting/aggregation over the ledger

Settlement/invoicing is **designed** but **not implemented** in v1.

## Lifecycle

Bus models micropayments as a lifecycle:
- **Capture**: record a transaction (e.g., from x402, manual entry, or import)
- **Aggregation**: group/summarize for reporting (by unit/service/network/asset/time window)
- **Settlement proposal** (future): compute net positions and produce a settlement plan
- **Invoicing/export** (future): generate invoice lines or accounting exports (not automated in v1)

v1 scope: **capture + logging + reporting** only.

## Core Model: `Transaction`

Bus Core MUST treat `Transaction` as the normalized ledger entry used for reporting and future settlement.

### Minimal Fields (Logical Schema)

Representable identically in YAML/TOML/JSON (see `16-multi-format-storage.md`).

`transaction` schema (micropayment-oriented shape):
- **id** (primary)
- **createdAt** (timestamp; RFC 3339)
- **fromUnitId** (payer business unit)
- **toUnitId** (payee business unit)
- **serviceId** (optional relation)
- **amount** (string; base units)
- **asset** (string; e.g., `"USD"`, `"USDC"`, `"EUR"`)
- **network** (optional string; only when a payment option requires it)
- **sourceType** (string; e.g., `"x402"`, `"manual"`, `"import"`)
- **sourceRef** (optional string; transaction id / external reference)
- **integrity** (optional fields; see “Integrity and Verification”)
  - contentHash (optional)
  - signature (optional)
  - nonce (optional)
  - validAfter (optional)
  - validBefore (optional)
- **rawProof** (optional string) or **proofHash** (optional string)

### Uniqueness Constraints (Spec)

To avoid double-booking, the store MUST support uniqueness constraints sufficient for:
- **At least**: uniqueness on \((sourceType, sourceRef)\) when `sourceRef` is present
- For x402 specifically: uniqueness on \((network, transaction)\) when the success header provides a transaction id
- Optional additional constraint: \((network, nonce)\) for replay resistance when nonces are used

## Logical Schemas for Services and x402 (Format-Agnostic)

These are logical schema specs; projects MAY name the schemas differently, but the fields and relations should remain consistent.

### A) `service` schema

Minimal extension-friendly service definition:
- **id** (primary)
- **name** (string)
- **resource** (string; endpoint/resource identifier)
- **description** (optional)

### B) `x402_policy` schema

One policy per service (or per service version) describing whether x402 capture is enabled:
- **id** (primary)
- **serviceId** (relation to `service`)
- **enabled** (boolean)
- **x402Version** (number; default `1`)
- **message** (optional string; human-readable hint shown to clients)

### C) `x402_accept` schema

One row per “accept option” in the 402 response:
- **id** (primary)
- **policyId** (relation to `x402_policy`)
- **scheme** (string)
- **network** (string)
- **maxAmountRequired** (string; base units)
- **payTo** (string)
- **asset** (string)
- **resource** (string) OR inherit from `service.resource` (see deterministic rule below)
- **description** (optional)
- **maxTimeoutSeconds** (optional number)

#### Deterministic `resource` Rule

To avoid ambiguity, Bus MUST define one deterministic rule. Recommended in v1:
- `x402_accept.resource` is **optional**.
- If present, it overrides.
- If absent, Bus uses `service.resource` via `x402_policy.serviceId`.

## Capture Mechanisms (v1)

Bus v1 supports multiple capture sources. x402 is one of them.

### x402 Capture (v1)

See `18-x402.md` for the wire formats and how Bus generates/ingests them.

### Manual/Import Capture (v1)

Bus v1 MAY also support recording transactions directly (e.g., manual entry or import) as `transaction` units. These sources MUST still produce normalized `Transaction` records to keep reporting and future settlement consistent.

## Integrity and Verification (Non-Blockchain)

Verification in Bus is about **authenticity**, **integrity**, and **replay prevention** of transactions and their proofs. It is **not** about blockchain network settlement or chain verification.

Bus Core SHOULD support simple, precise mechanisms (spec-only in v1):
- **Canonical serialization**: define a canonical payload for signing/hashing (stable key ordering, stable types)
- **Signed receipts**: payer signs the canonical transaction payload (or a receipt referencing it)
- **Nonces + validity windows**: `nonce`, `validAfter`, `validBefore` to reduce replay
- **Content hashes**: `contentHash` for immutable audit trails
- **Idempotency keys / uniqueness constraints**: prevent double booking at ingestion time

Design goal: keep verification primitives small and composable; do not embed complex network logic in the core.

## Settlement and Invoicing Extensions (Future; Modular)

Bus is primarily a **transaction collector and ledger** between business units. Over time, Bus should support converting transactions into settlement artifacts without modifying the core.

### Minimal Modular Architecture

Core responsibilities:
- Define the `Transaction` model
- Provide storage/query/report primitives over transactions
- Provide optional verification primitives (hash/signature/nonces, uniqueness checks)

Extension responsibilities (plugin-style):
- Implement **settlement providers** that turn transactions into settlement artifacts, e.g.:
  - netting reports (by counterparty and period)
  - invoice line generation (exportable, not automated charging)
  - accounting exports (CSV/JSON) for external systems

Requirement:
- New payment options and settlement/export methods MUST be addable as extensions without modifying unrelated core logic.

### Producer vs Collector Modes

Bus should be usable from both sides:
- **Producer mode**: a client or service captures transactions locally for audit/reporting
- **Collector mode** (future hosted facilitator): a central Bus service receives transactions from multiple producers and maintains a shared ledger

Even when a hosted facilitator exists, the **CLI remains first-class** for scripting; the CLI can operate locally or call a remote facilitator (spec-only).


