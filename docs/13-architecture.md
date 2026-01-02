# Internal Go Architecture

The tool must be modular: CLI is thin, engine is pluggable via interfaces. This aligns with the modular core principles in the Bus design.

## Core vs Front-Ends (Hosted Facilitator Architecture)

Bus MUST have:
- **Bus Core**: all business logic (load/validate docs, schema system, unit store, transaction logging, reporting, integrity primitives)
- **Front-ends**:
  - **CLI** (v1): first-class interface for shell scripting
  - **HTTP server** (future): a hosted facilitator that exposes the same core operations over HTTP

Requirement:
- CLI and HTTP MUST be thin layers over the same core. Adding HTTP later MUST NOT require rewriting core logic.
- The HTTP server MUST publish an **OpenAPI** specification so remote client libraries can be automatically generated.

## Package Structure (Recommended)

### `cmd/bus/`
* Cobra root command
* Argument parsing
* Command dispatch
* Thin layer that delegates to feature modules

### `cmd/busd/` (future)
* HTTP server front-end (“hosted facilitator”)
* Auth/transport concerns only
* Delegates to the same core operations as the CLI
* Publishes an OpenAPI spec (recommended: OpenAPI 3.x) describing all endpoints and schemas

### `internal/engine/`
* Core interfaces
* Wiring and dependency injection
* Feature module registration

### `internal/store/`
* Filesystem store implementation
* Locking mechanisms
* Atomic writer utilities

### `internal/manifest/`
* Manifest discovery and read/write (`bus.{yml,yaml,toml,json}`)
* Manifest validation
* Path resolution

### `internal/codec/`
* Format codecs (YAML/TOML/JSON)
* Format registry (extension → codec)
* Deterministic encoding rules per codec

### `internal/schema/`
* Schema parsing
* Schema validation
* Property type checking

### `internal/unit/`
* Unit CRUD operations
* Index management
* Uniqueness checks

### `internal/micropayments/` (v1 feature)
* x402 requirement generation from config
* x402 header ingestion
* Normalization into `transaction` records
* Reporting/aggregation over transactions (micropayment-oriented views)

## Core Interfaces (Minimal)

### `ManifestStore`
Manifest operations:

```go
type ManifestStore interface {
    Load() (*Manifest, error)
    Save(*Manifest) error
}
```

### `SchemaStore`
Schema loading:

```go
type SchemaStore interface {
    LoadSchema(name string) (*Schema, error) // path resolved via manifest
}
```

### `UnitStore`
Unit operations:

```go
type UnitStore interface {
    ListIDs(schema string) ([]string, error)
    Load(schema, id string) (*Unit, error)
    Create(schema string, unit *Unit) error // atomic: writes unit file + updates ids
}
```

### `Locker`
Workspace locking:

```go
type Locker interface {
    Lock() (unlock func(), err error)
}
```

## Feature Modules

Each top-level CLI command is implemented by a feature module:

### `InitFeature`
* Implements `bus init`
* Creates the manifest and `.bus/` directory

### `SchemaFeature`
* Implements `bus schema init`
* Schema file creation
* Manifest registration

### `UnitFeature`
* Implements schema-specific commands (`bus <schema> add/list/show`)
* Implements compatibility prefix (`bus unit <schema> ...`)
* Unit CRUD operations

### `MicropaymentsFeature` (v1)
* Implements micropayments reporting over `transaction` records

### `X402Feature` (v1)
* Implements x402 requirement generation and ingestion

## Design Principles

### Separation of Concerns
* CLI layer: argument parsing, user interaction
* Feature layer: business logic, orchestration
* Store layer: persistence, atomic operations
* Engine layer: pure computation

### Pluggability
* Interfaces allow swapping implementations
* Feature modules can be added/removed
* Store implementations can be changed (e.g., for testing)

### Testability
* Pure functions where possible
* Interfaces enable mocking
* Store layer can be tested independently

### Extensibility
* New features can be added as modules
* New store implementations can be added
* New settlement/export providers can be added without modifying core logic (future)

## Proposed Core Operations (Spec-Only)

Core operations should be transport-agnostic and callable from both CLI and HTTP:
- Generate x402 payment requirement (402 body) from service/policy/config
- Ingest x402 proof headers and log a normalized transaction
- Record transaction (manual/import sources)
- Report totals over transactions (by unit/service/network/asset/time window)
- Export a settlement proposal (future)

## Proposed HTTP Surface (Future; High-Level)

When a hosted facilitator exists, the HTTP server should expose the same operations:
- `POST /v1/x402/requirements` (generate 402 body)
- `POST /v1/x402/ingest` (ingest proof headers and log transaction)
- `POST /v1/transactions` (record transaction)
- `GET /v1/micropayments/report` (report/aggregate)
- `GET /v1/settlement/proposal` (future)

## OpenAPI Requirement (Future; Hosted Facilitator)

The hosted facilitator MUST provide an OpenAPI document that:
- Fully describes all HTTP endpoints, request/response bodies, and error shapes
- Includes schemas for core types (service/policy/accept and transaction records)
- Is stable and versioned (e.g., `/openapi.json` for the current server version, plus versioned API paths like `/v1/...`)

Rationale: this enables automatic generation of remote client libraries and keeps the HTTP layer a thin, spec-driven wrapper over Bus Core operations.

## Dependency Flow

```
CLI (cmd/bus)
  ↓
Feature Modules (internal/engine)
  ↓
Stores (internal/store, internal/manifest, etc.)
  ↓
Filesystem
```

## Future Considerations

This architecture supports:
* Multiple store backends (filesystem, database, etc.)
* Plugin system for custom features
* Testing with in-memory stores

