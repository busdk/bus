# What v1 Deliberately Does Not Do

This document lists features and capabilities that are explicitly **not** included in Bus CLI v1.

## Git Operations

**v1 does not:**
* Initialize Git repositories
* Commit changes
* Merge branches
* Perform any Git operations

Git is considered "outside" the tool. Users manage Git themselves.

## Automatic Conflict Resolution

**v1 does not:**
* Automatically resolve merge conflicts
* Provide conflict resolution tools
* Detect or handle conflicting changes

The file model is designed to make Git merges easier, but Bus does not perform the merge itself.

## User Interface

**v1 does not provide:**
* Web UI
* Graphical interface
* Interactive prompts (beyond basic CLI)

v1 is CLI-only.

## REST API

**v1 does not provide:**
* HTTP API
* REST endpoints
* Remote access

v1 is local-only.

## Invoices and Settlement Flows

**v1 does not:**
* Generate invoices
* Handle settlements
* Process payments
* Create financial documents

Ledger records exist; invoices and settlements can be built on top of the ledger in future versions.

## Advanced Features

### Relation Management
* No explicit relation entities (relations are implicit in unit data)
* No bidirectional relations
* No relation metadata beyond what's in unit properties

### Billing Rules
* No built-in billing engine
* No built-in recurring generation
* Billing/ledger concepts are modeled using schemas + units (see `06-billing.md`)

### Transaction Features
* No transaction editing (append-only)
* No transaction deletion
* No transaction aggregation or reporting

### Schema Features
* No schema versioning
* No schema migration tools
* No schema validation beyond basic type checking
* No schema inheritance or composition

### Unit Features
* No unit updates (only create, list, show)
* No unit deletion
* No bulk operations
* No unit search or filtering

### Data Import/Export
* No import from other formats
* No export to other formats
* No data migration tools

## Why These Limitations?

v1 focuses on:
* **Core functionality**: Basic unit and transaction management
* **Simplicity**: Minimal feature set to prove the concept
* **Git-friendly**: File-based storage that works well with version control
* **Foundation**: Architecture that supports future features

Future versions may add:
* Unit updates and deletion
* More billing rule types
* Invoice generation
* REST API
* Web UI
* Advanced relation management
* And more...

## What v1 Does Provide

Despite these limitations, v1 provides:
* Schema-driven unit management
* The building blocks to model append-only ledgers as schemas + units
* Git-friendly file structure
* Atomic operations
* Deterministic output
* Modular architecture for future extensions

