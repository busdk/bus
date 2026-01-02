# Bus CLI v1 Overview

## Purpose

The Bus CLI is a **local, Git-friendly** data management tool that implements the Bus design principles: relationships/contracts and an append-only ledger.

## Goals

The first usable version (v1) of Bus CLI aims to:

* Write **deterministic**, merge-friendly files
* Never leave the repository in an inconsistent state (atomic operations)
* **Not** perform Git operations itself (Git is "outside" the tool)
* Store user-facing schema definitions next to the manifest (`bus.{yml,yaml,toml,json}`) by default
* Store all mutable state under `.bus/`
* Support a **transaction ledger** for micropayment capture and reporting (settlement is future)

## Core Concepts

### Units

Units are instances of schemas. Each unit has:
* Properties defined by its schema, including a primary ID property
* References to other units (typed relations)

The primary ID is one of the unit's properties, specified by the schema. The schema defines which property serves as the ID and its type:
* **String IDs**: Unique identifiers that must be provided when the unit is instantiated
* **UUID IDs**: Auto-populate with a fresh UUID (v4) when the unit is created, or may be provided by the user
* **Integer IDs**: Must be provided when the unit is created (no auto-generation)

### Schemas

Schemas define:
* Property types and constraints
* Primary ID configuration
* Optional allowed operations (schema-level CRUD policy)
* Relationships to other schemas (typed references)

### Transactions

Transactions are append-only records that:
* Represent economic flows (provider → consumer)
* Support internal chargeback flows
* Are immutable (corrections are new transactions)

In v1, a “transaction” is just a **unit** in a schema you define (often named `transaction`).

### Relations

Relations are represented as typed references (`ref:<schema>`) plus any domain properties you choose (e.g., contract start/end, pricing terms). This keeps the data model minimal and schema-driven.

## Design Philosophy

Bus v1 focuses on:
* **Local-first**: All data stored in files, Git-friendly structure
* **Schema-driven**: Business logic defined in schemas, not code
* **Deterministic**: Same inputs produce same outputs (merge-friendly)
* **Atomic**: Operations either complete fully or make no changes
* **Modular**: Internal architecture supports pluggable features

