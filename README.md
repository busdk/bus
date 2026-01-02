# Bus

Bus is a **modular, schema-driven** system for managing business-unit data and append-only accounting history in a **local, Git-friendly** workspace.

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
