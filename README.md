# Bus

Bus models an organization as **tenant-scoped Units** (teams, projects, infra, vendors, etc.) that **provide and consume Services** under explicit **relationships/contracts**. As services are used, Bus records the activity as **append-only, ledger-like transactions** (debits/credits between units), so balances can be reconciled and later settled (e.g., via generated invoices or exports to external accounting systems).

Unlike a traditional DB-first tool, Bus keeps **schemas and workspace config in a Git workspace**: changes are reviewable, mergeable, and auditable as commits. It starts **CLI-first (script/agent friendly)** and is designed to later expose the same core via a REST API, without moving feature logic into the core. Internal mutable state is **pluggable** (filesystem `.bus/` or a database backend).

## Start here

- **Docs entry point**: `docs/README.md`
- **Roadmap start (authoritative)**: `docs/roadmap/0.0.0.md`

The design is written as an ordered sequence of small SemVer increments:
- `docs/roadmap/0.0.0.md`, `docs/roadmap/0.0.1.md`, …
- Each version file is one implementable increment with acceptance criteria.

## Core principles (short form)

- **Modular-by-default**: features live in modules behind interfaces.
- **Small, stable core**: core orchestrates modules; it does not accumulate feature logic.
- **Append-only history**: ledger-like records are create-only; corrections are new records.
- **Multi-tenant from day one**: every operation is tenant-scoped.
- **Interfaces first**: adding a feature should be “implement an interface + register a module”.

## Where extension points live

The core extension points are defined in the spec docs and introduced early in the roadmap:
- `docs/spec/module-runtime.md`
- `docs/spec/transports-cli.md`
