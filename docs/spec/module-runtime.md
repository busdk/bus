# Interfaces + Implementations (Compile-Time Composition)

## What this is
Bus is modular via **internal Go interfaces** and **swappable implementations**.

Implementations are compiled into the binary, and Bus may include **multiple built-in implementations for the same interface** at the same time (e.g., filesystem vs database state backends). Selection between built-in implementations is **runtime**, driven by config, but remains **static in code** (no plugins).

There is **no runtime module system**:
- no module lifecycle (`Init/Start/Stop`)
- no dependency graph / cycle detection
- no dynamic capability registry
- no plugin discovery / loading external code

Instead, “features” are **packages** that provide implementations of core-owned interfaces and (optionally) transport bindings (CLI/HTTP) that call core operations.

## Interfaces (names are stable contracts)

### `App` (composition root; suggested)
Responsibility: hold references to the concrete implementations used by this Bus build.

`App` is constructed once (e.g., in `main`) and passed to CLI/HTTP front-ends.

### Core-owned service interfaces (examples)
Responsibility: narrow, stable seams between components. Implementations can vary without changing callers.

Examples (defined elsewhere in `docs/spec/*` and referenced here):
- workspace config IO: `ManifestStore`, `SchemaStore` (filesystem-based)
- internal state: `StateBackend` (filesystem `.bus` or database)
- domain ops: schema init/validation, unit create/list/show, reporters, exporters

## Wiring + runtime selection (binding)
All dependencies are wired via code, but selection may occur at startup:

- **Compile-time**: the binary includes one or more implementations of an interface.
- **Runtime**: Bus selects which built-in implementation to use based on config (deterministically).

### Provider registry pattern (simple, built-in)
For any interface that supports multiple built-in implementations, Bus uses a small in-process registry:
- key: stable string id (e.g., `"filesystem"`, `"database"`)
- value: factory/constructor that returns an implementation

This is **not** a module system; it is a deterministic selection mechanism for built-in implementations.

Examples:
- `StateBackend` selected by `state.backend` (see `docs/spec/state-storage.md`)
- codecs selected by file extension (registry maps extensions → codec)

### CLI binding
The CLI binds command handlers to core operations by calling methods on the composed `App`.

Rule:
- Adding a feature should still be “implement an interface + wire/register the implementation”, but wiring/registration is **code**, not plugin discovery.

## Required behaviors

### Core boundary
- Core provides orchestration and cross-cutting enforcement hooks.
- Business logic is expressed behind interfaces (in packages), with implementations swapped via wiring.


