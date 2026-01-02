# Vocabulary (Canonical)

## What this is
The canonical terms used across the roadmap and spec. If a term is used, it MUST match these meanings.

## Terms

### Workspace
The working directory that contains exactly one manifest candidate:
- `bus.yml|bus.yaml|bus.toml|bus.json`

### Manifest
The workspace root configuration document (`kind: bus.manifest`).

### Tenant
An isolation domain. All reads/writes happen in a tenant context (including local CLI mode).

### Schema
A document (`kind: bus.schema`) defining unit shape and constraints.

### Unit
A document (`kind: bus.unit`) that is an instance of a schema.

### Relation
A schema-driven typed reference (`ref:<schemaName>`) plus domain properties.

### Transaction
An append-only ledger entry. In this design, a transaction is stored as a unit in a schema you define (commonly named `transaction`).

### Module
An independently implementable feature package that implements core-owned interfaces and is wired into the app as a built-in implementation (no runtime module system).

### Capability
Something a module exposes to transports (CLI/API): commands, endpoints, exporters, providers, etc.


