# Bus CLI v1 Documentation

This directory contains documentation for Bus CLI v1, organized by topic.

## Documentation Index

1. **[Overview](01-overview.md)** - Introduction, goals, and core concepts
2. **[Constraints](02-constraints.md)** - Non-negotiable design constraints
3. **[Workspace Structure](03-workspace-structure.md)** - Files and directory organization
4. **[Manifest](04-manifest.md)** - `bus.yml` format and structure
5. **[Schemas](05-schemas.md)** - Schema file format, property types, and attributes
6. **[Billing](06-billing.md)** - Billing / ledger as schema patterns (no built-in billing engine)
7. **[Units](07-units.md)** - Unit storage format and operations
8. **[Relations](08-relations.md)** - Relations model and provider/consumer pattern
9. **[Transactions](09-transactions.md)** - Ledger and transaction format
10. **[CLI Commands](10-cli-commands.md)** - Command reference and usage
11. **[Consistency](11-consistency.md)** - Atomic operations and locking
12. **[Formatting](12-formatting.md)** - Deterministic formatting requirements
13. **[Architecture](13-architecture.md)** - Internal Go architecture and interfaces
14. **[Examples](14-examples.md)** - Example workflows and usage
15. **[Limitations](15-limitations.md)** - What v1 does not do

## Quick Start

For a quick introduction, start with:
1. [Overview](01-overview.md) - Understand the purpose and goals
2. [Examples](14-examples.md) - See it in action
3. [CLI Commands](10-cli-commands.md) - Learn the commands

## Design Principles

Bus CLI v1 is designed to be:
* **Local-first**: All data stored in files
* **Git-friendly**: Deterministic, merge-friendly structure
* **Schema-driven**: Business logic in schemas, not code
* **Atomic**: Operations either complete or make no changes
* **Modular**: Architecture supports future extensions

## Related Documents

The main design document that this documentation is based on is the Bus CLI v1 Design Document, which provides the complete specification.

