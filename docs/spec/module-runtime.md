# Module Runtime (Core Extension Point)

## What this is
The core module system contract. Bus Core owns these interfaces and orchestrates the lifecycle. Modules implement them.

## Interfaces (names are stable contracts)

### `Module`
Responsibility: declare identity/dependencies and register capabilities.

Required:
- `ID() string` — globally unique module id (e.g., `"core.units"`).
- `Version() string` — module version (informational).
- `Requires() []string` — module ids required to be present.
- `Register(r Registrar) error` — register capabilities and implementations.

Optional lifecycle hooks:
- `Init(ctx RuntimeContext) error`
- `Start(ctx RuntimeContext) error`
- `Stop(ctx RuntimeContext) error`

### `Registrar`
Responsibility: the only way modules interact with core at registration time.

Minimum responsibilities:
- Register a **capability** (commands/endpoints/exporters/etc).
- Provide an implementation for a **core-owned interface** (dependency injection by interface).

### `ModuleRegistry`
Responsibility: discovery + storage of available modules.

- `RegisterModule(m Module) error`
- `ListModules() []Module`

### `ModuleRuntime`
Responsibility: validate, order, initialize, and run modules.

- `Load(reg ModuleRegistry, cfg RuntimeConfig) error`
- `Start() error`
- `Stop() error`
- `Capabilities() CapabilitySet`

### `CapabilitySet` (opaque)
Responsibility: read-only view of registered capabilities for transports.

Rule:
- Core treats capability registration as **data**, not branching logic.

## Required behaviors

### Dependency ordering
- Runtime MUST produce a deterministic initialization order.
- Cycles in `Requires()` MUST be detected and reported as an error.

### Core boundary
- Core provides orchestration and cross-cutting enforcement hooks.
- Business logic lives in modules.


