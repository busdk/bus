# Bus Documentation

## Bus

Bus models an organization as **tenant-scoped Units** (teams, projects, infra, vendors, etc.) that **provide and consume Services** under explicit **relationships/contracts**. As services are used, Bus records the activity as **append-only, ledger-like transactions** (debits/credits between units), so balances can be reconciled and later settled (e.g., via generated invoices or exports to external accounting systems).

Unlike a traditional DB-first tool, Bus keeps **schemas and workspace config in a Git workspace**: changes are reviewable, mergeable, and auditable as commits. It starts **CLI-first (script/agent friendly)** and is designed to later expose the same core via a REST API, without moving feature logic into the core. Internal mutable state is **pluggable** (filesystem `.bus/` or a database backend).

## Start here

- **Roadmap start (authoritative)**: `roadmap/0.0.1.md`

The design is written as an ordered sequence of small SemVer increments:
- `roadmap/0.0.1.md`, `roadmap/0.0.2.md`, …
- Each version file is one implementable increment with acceptance criteria.

## Core principles (short form)

- **Modular-by-default**: features live in modules behind interfaces.
- **Small, stable core**: core orchestrates modules; it does not accumulate feature logic.
- **Append-only history**: ledger-like records are create-only; corrections are new records.
- **Multi-tenant from day one**: every operation is tenant-scoped.
- **Interfaces first**: adding a feature should be “implement an interface + register a module”.

## Where extension points live

The core extension points are defined in the spec docs and introduced early in the roadmap:
- `spec/module-runtime.md`
- `spec/transports-cli.md`

## Roadmap index (grouped by minor milestones)

This section is the roadmap index, included here so `docs/README.md` can act as a single entry point.

Each `docs/roadmap/{VERSION}.md` file is one implementable increment.

This index groups patch steps into **minor-version milestones** for easier planning and review.

Note: some file browsers (including GitHub’s folder view) sort filenames lexicographically, not by SemVer.
Use this index as the authoritative sequence.

### v0.0 — Foundations
- [0.0.1](roadmap/0.0.1.md) — Doc conventions, vocabulary, and non-negotiable principles.
- [0.0.2](roadmap/0.0.2.md) — **Core**: interfaces + built-in implementation selection.
- [0.0.3](roadmap/0.0.3.md) — **Core**: YAML codec only.
- [0.0.4](roadmap/0.0.4.md) — **Core**: manifest discovery (YAML only).
- [0.0.5](roadmap/0.0.5.md) — **Core**: workspace init (create `bus.yml`).

### v0.1 — CLI bootstrap + tenancy + state (filesystem)
- [0.1.1](roadmap/0.1.1.md) — **CLI**: skeleton (no commands yet).
- [0.1.2](roadmap/0.1.2.md) — **CLI command**: `bus init`.
- [0.1.3](roadmap/0.1.3.md) — **Core**: tenancy defaults.
- [0.1.4](roadmap/0.1.4.md) — **Core**: internal state interfaces + filesystem `.bus` backend.

### v0.2 — Schemas + units + CLI unit operations
- [0.2.1](roadmap/0.2.1.md) — **Core**: schema validation.
- [0.2.2](roadmap/0.2.2.md) — **Core**: schema registry (manifest `units[]`).
- [0.2.3](roadmap/0.2.3.md) — **Core**: schema init operation (create + register).
- [0.2.4](roadmap/0.2.4.md) — **CLI command**: `bus schema init`.
- [0.2.5](roadmap/0.2.5.md) — **Core**: unit create operation.
- [0.2.6](roadmap/0.2.6.md) — **CLI command**: `bus <schema> add`.
- [0.2.7](roadmap/0.2.7.md) — **Core**: unit list/show operations.
- [0.2.8](roadmap/0.2.8.md) — **CLI command**: `bus <schema> list`.
- [0.2.9](roadmap/0.2.9.md) — **CLI command**: `bus <schema> show`.

### v0.3 — Schema-driven validation + ledger pattern + more formats/backends
- [0.3.1](roadmap/0.3.1.md) — **Core**: unit validation (required + types).
- [0.3.2](roadmap/0.3.2.md) — **Core**: unit validation (refs).
- [0.3.3](roadmap/0.3.3.md) — **Core**: unit validation (uniqueness + `uniqueScope`).
- [0.3.4](roadmap/0.3.4.md) — Docs: ledger/transactions pattern (append-only).
- [0.3.5](roadmap/0.3.5.md) — **Core**: JSON codec.
- [0.3.6](roadmap/0.3.6.md) — **Core**: TOML codec.
- [0.3.7](roadmap/0.3.7.md) — **Core**: database state backend.

### v0.4 — Micropayments
- [0.4.1](roadmap/0.4.1.md) — **Core**: micropayments report operation.
- [0.4.2](roadmap/0.4.2.md) — **CLI command**: `bus micropayments report`.

### v0.5 — Hosted HTTP (facilitator server + generic endpoints)
- [0.5.1](roadmap/0.5.1.md) — Contract: hosted facilitator (HTTP + OpenAPI).
- [0.5.2](roadmap/0.5.2.md) — **HTTP**: server skeleton + OpenAPI publishing.
- [0.5.3](roadmap/0.5.3.md) — **HTTP endpoint**: `POST /v1/transactions`.
- [0.5.4](roadmap/0.5.4.md) — **HTTP endpoint**: `GET /v1/micropayments/report`.

### v0.6 — Exports + settlement
- [0.6.1](roadmap/0.6.1.md) — **Core**: export/settlement extension point interfaces.
- [0.6.2](roadmap/0.6.2.md) — **HTTP endpoint**: `GET /v1/settlement/proposal`.

### v0.7 — Operational automation + secrets
- [0.7.1](roadmap/0.7.1.md) — **Core**: operational automation (tasks/playbooks) interfaces.

### v0.8 — x402 (feature + HTTP endpoints)
- [0.8.1](roadmap/0.8.1.md) — **HTTP feature**: x402 “generate 402 body” operation.
- [0.8.2](roadmap/0.8.2.md) — **HTTP endpoint**: `POST /v1/x402/requirements`.
- [0.8.3](roadmap/0.8.3.md) — **HTTP feature**: x402 parse headers operation.
- [0.8.4](roadmap/0.8.4.md) — **HTTP feature**: x402 normalize into transaction.
- [0.8.5](roadmap/0.8.5.md) — **HTTP feature**: x402 uniqueness / double-booking prevention.
- [0.8.6](roadmap/0.8.6.md) — **HTTP endpoint**: `POST /v1/x402/ingest`.

### v0.9 — Secrets
- [0.9.1](roadmap/0.9.1.md) — **Core**: secret fields (annotation + redaction rules).

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

