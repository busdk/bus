# Roadmap Index (Grouped by Minor Version Milestones)

Each `docs/roadmap/{VERSION}.md` file is one implementable increment.

This index groups patch steps into **minor-version milestones** for easier planning and review.

## v0.0 — Foundations
- [0.0.0](0.0.0.md) — Doc conventions, vocabulary, and non-negotiable principles.
- [0.0.1](0.0.1.md) — **Core**: interfaces + built-in implementation selection.
- [0.0.2](0.0.2.md) — **Core**: YAML codec only.
- [0.0.3](0.0.3.md) — **Core**: manifest discovery (YAML only).
- [0.0.4](0.0.4.md) — **Core**: workspace init (create `bus.yml`).

## v0.1 — CLI bootstrap + tenancy + state (filesystem)
- [0.0.5](0.0.5.md) — **CLI**: skeleton (no commands yet).
- [0.0.6](0.0.6.md) — **CLI command**: `bus init`.
- [0.0.7](0.0.7.md) — **Core**: tenancy defaults.
- [0.0.8](0.0.8.md) — **Core**: internal state interfaces + filesystem `.bus` backend.

## v0.2 — Schemas + units + CLI unit operations
- [0.0.9](0.0.9.md) — **Core**: schema validation.
- [0.0.10](0.0.10.md) — **Core**: schema registry (manifest `units[]`).
- [0.0.11](0.0.11.md) — **Core**: schema init operation (create + register).
- [0.0.12](0.0.12.md) — **CLI command**: `bus schema init`.
- [0.0.13](0.0.13.md) — **Core**: unit create operation.
- [0.0.14](0.0.14.md) — **CLI command**: `bus <schema> add`.
- [0.0.15](0.0.15.md) — **Core**: unit list/show operations.
- [0.0.16](0.0.16.md) — **CLI command**: `bus <schema> list`.
- [0.0.17](0.0.17.md) — **CLI command**: `bus <schema> show`.

## v0.3 — Schema-driven validation + ledger pattern + more formats/backends
- [0.0.18](0.0.18.md) — **Core**: unit validation (required + types).
- [0.0.19](0.0.19.md) — **Core**: unit validation (refs).
- [0.0.20](0.0.20.md) — **Core**: unit validation (uniqueness + `uniqueScope`).
- [0.0.21](0.0.21.md) — Docs: ledger/transactions pattern (append-only).
- [0.0.22](0.0.22.md) — **Core**: JSON codec.
- [0.0.23](0.0.23.md) — **Core**: TOML codec.
- [0.0.24](0.0.24.md) — **Core**: database state backend.

## v0.4 — Micropayments
- [0.0.32](0.0.32.md) — **Core**: micropayments report operation.
- [0.0.33](0.0.33.md) — **CLI command**: `bus micropayments report`.

## v0.5 — Hosted HTTP + x402
- [0.0.41](0.0.41.md) — Contract: hosted facilitator (HTTP + OpenAPI).
- [0.0.44](0.0.44.md) — **HTTP**: server skeleton + OpenAPI publishing.
- [0.0.34](0.0.34.md) — **HTTP feature**: x402 “generate 402 body” operation.
- [0.0.35](0.0.35.md) — **HTTP endpoint**: `POST /v1/x402/requirements`.
- [0.0.36](0.0.36.md) — **HTTP feature**: x402 parse headers operation.
- [0.0.37](0.0.37.md) — **HTTP feature**: x402 normalize into transaction.
- [0.0.38](0.0.38.md) — **HTTP feature**: x402 uniqueness / double-booking prevention.
- [0.0.39](0.0.39.md) — **HTTP endpoint**: `POST /v1/x402/ingest`.
- [0.0.45](0.0.45.md) — **HTTP endpoint**: `POST /v1/transactions`.
- [0.0.46](0.0.46.md) — **HTTP endpoint**: `GET /v1/micropayments/report`.

## v0.6 — Exports + settlement
- [0.0.40](0.0.40.md) — **Core**: export/settlement extension point interfaces.
- [0.0.47](0.0.47.md) — **HTTP endpoint**: `GET /v1/settlement/proposal`.

## v0.7 — Operational automation + secrets
- [0.0.42](0.0.42.md) — **Core**: operational automation (tasks/playbooks) interfaces.
- [0.0.43](0.0.43.md) — **Core**: secret fields (annotation + redaction rules).


