# Bus Documentation

Bus models an organization as **tenant-scoped Units** (teams, projects, infra, vendors, etc.) that **provide and consume Services** under explicit **relationships/contracts**. As services are used, Bus records the activity as **append-only, ledger-like transactions** (debits/credits between units), so balances can be reconciled and later settled (e.g., via generated invoices or exports to external accounting systems).

Unlike a traditional DB-first tool, Bus keeps **schemas and workspace config in a Git workspace**: changes are reviewable, mergeable, and auditable as commits. It starts **CLI-first (script/agent friendly)** and is designed to later expose the same core via a REST API, without moving feature logic into the core. Internal mutable state is managed behind an extendable **`StateBackend` interface**, with multiple built-in implementations (e.g. filesystem `.bus/` or a database backend) that are **selected at runtime via config**. Over time, Bus can also serve as a foundation for **operational automation** by attaching tasks/playbooks to units and services (while handling sensitive values as protected “secret” fields).

## Roadmap

Each file in `docs/roadmap/{VERSION}.md` is one implementable increment for Bus `{VERSION}`.

Minor version “milestones” group related patches; patch steps are still authored as `0.0.x` docs:
- Grouped view: `docs/roadmap/README.md`

### v0.0 — Foundations
- [0.0.0](roadmap/0.0.0.md) — Doc conventions, vocabulary, and non-negotiable principles.
- [0.0.1](roadmap/0.0.1.md) — **Core**: interfaces + built-in implementation selection.
- [0.0.2](roadmap/0.0.2.md) — **Core**: YAML codec only.
- [0.0.3](roadmap/0.0.3.md) — **Core**: manifest discovery (YAML only).
- [0.0.4](roadmap/0.0.4.md) — **Core**: workspace init (create `bus.yml`).

### v0.1 — CLI bootstrap + tenancy + state (filesystem)
- [0.0.5](roadmap/0.0.5.md) — **CLI**: skeleton (no commands yet).
- [0.0.6](roadmap/0.0.6.md) — **CLI command**: `bus init`.
- [0.0.7](roadmap/0.0.7.md) — **Core**: tenancy defaults.
- [0.0.8](roadmap/0.0.8.md) — **Core**: internal state interfaces + filesystem `.bus` backend.

### v0.2 — Schemas + units + CLI unit operations
- [0.0.9](roadmap/0.0.9.md) — **Core**: schema validation.
- [0.0.10](roadmap/0.0.10.md) — **Core**: schema registry (manifest `units[]`).
- [0.0.11](roadmap/0.0.11.md) — **Core**: schema init operation (create + register).
- [0.0.12](roadmap/0.0.12.md) — **CLI command**: `bus schema init`.
- [0.0.13](roadmap/0.0.13.md) — **Core**: unit create operation.
- [0.0.14](roadmap/0.0.14.md) — **CLI command**: `bus <schema> add`.
- [0.0.15](roadmap/0.0.15.md) — **Core**: unit list/show operations.
- [0.0.16](roadmap/0.0.16.md) — **CLI command**: `bus <schema> list`.
- [0.0.17](roadmap/0.0.17.md) — **CLI command**: `bus <schema> show`.

### v0.3 — Schema-driven validation + ledger pattern + more formats/backends
- [0.0.18](roadmap/0.0.18.md) — **Core**: unit validation (required + types).
- [0.0.19](roadmap/0.0.19.md) — **Core**: unit validation (refs).
- [0.0.20](roadmap/0.0.20.md) — **Core**: unit validation (uniqueness + `uniqueScope`).
- [0.0.21](roadmap/0.0.21.md) — Docs: ledger/transactions pattern (append-only).
- [0.0.22](roadmap/0.0.22.md) — **Core**: JSON codec.
- [0.0.23](roadmap/0.0.23.md) — **Core**: TOML codec.
- [0.0.24](roadmap/0.0.24.md) — **Core**: database state backend.

### v0.4 — Micropayments
- [0.0.32](roadmap/0.0.32.md) — **Core**: micropayments report operation.
- [0.0.33](roadmap/0.0.33.md) — **CLI command**: `bus micropayments report`.

### v0.5 — Hosted HTTP + x402
- [0.0.41](roadmap/0.0.41.md) — Contract: hosted facilitator (HTTP + OpenAPI).
- [0.0.44](roadmap/0.0.44.md) — **HTTP**: server skeleton + OpenAPI publishing.
- [0.0.34](roadmap/0.0.34.md) — **HTTP feature**: x402 “generate 402 body” operation.
- [0.0.35](roadmap/0.0.35.md) — **HTTP endpoint**: `POST /v1/x402/requirements`.
- [0.0.36](roadmap/0.0.36.md) — **HTTP feature**: x402 parse headers operation.
- [0.0.37](roadmap/0.0.37.md) — **HTTP feature**: x402 normalize into transaction.
- [0.0.38](roadmap/0.0.38.md) — **HTTP feature**: x402 uniqueness / double-booking prevention.
- [0.0.39](roadmap/0.0.39.md) — **HTTP endpoint**: `POST /v1/x402/ingest`.
- [0.0.45](roadmap/0.0.45.md) — **HTTP endpoint**: `POST /v1/transactions`.
- [0.0.46](roadmap/0.0.46.md) — **HTTP endpoint**: `GET /v1/micropayments/report`.

### v0.6 — Exports + settlement
- [0.0.40](roadmap/0.0.40.md) — **Core**: export/settlement extension point interfaces.
- [0.0.47](roadmap/0.0.47.md) — **HTTP endpoint**: `GET /v1/settlement/proposal`.

### v0.7 — Operational automation + secrets
- [0.0.42](roadmap/0.0.42.md) — **Core**: operational automation (tasks/playbooks) interfaces.
- [0.0.43](roadmap/0.0.43.md) — **Core**: secret fields (annotation + redaction rules).

## Spec (reusable concepts)

- [Principles (non-negotiable)](spec/principles.md)
- [Vocabulary](spec/vocabulary.md)
- [Module runtime](spec/module-runtime.md)
- [Transports (CLI)](spec/transports-cli.md)
- [Workspace + manifest](spec/workspace-manifest.md)
- [Tenancy](spec/tenancy.md)
- [Codecs + deterministic formatting](spec/codecs-and-formatting.md)
- [YAML codec](spec/codec-yaml.md)
- [JSON codec](spec/codec-json.md)
- [TOML codec](spec/codec-toml.md)
- [Locking + atomic writes](spec/locking-and-atomic-writes.md)
- [Schemas](spec/schemas.md)
- [Units](spec/units.md)
- [Internal state storage (runtime-selectable backend)](spec/state-storage.md)
- [Filesystem state backend (`.bus/`)](spec/state-backend-dotbus.md)
- [Database state backend](spec/state-backend-database.md)
- [Relations](spec/relations.md)
- [Ledger / transactions](spec/ledger-transactions.md)
- [Micropayments](spec/micropayments.md)
- [x402](spec/x402.md)
- [Exports + settlement extension points](spec/exports-and-settlement.md)
- [Hosted facilitator (HTTP + OpenAPI)](spec/hosted-facilitator-http.md)
- [Operational automation (tasks/playbooks)](spec/operational-automation.md)
- [Secret fields](spec/secrets.md)

