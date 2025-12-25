# Internal Go Architecture

The tool must be modular: CLI is thin, engine is pluggable via interfaces. This aligns with the modular core principles in the Bus design.

## Package Structure (Recommended)

### `cmd/bus/`
* Cobra root command
* Argument parsing
* Command dispatch
* Thin layer that delegates to feature modules

### `internal/engine/`
* Core interfaces
* Wiring and dependency injection
* Feature module registration

### `internal/store/`
* Filesystem store implementation
* Locking mechanisms
* Atomic writer utilities

### `internal/manifest/`
* Read/write `bus.yml`
* Manifest validation
* Path resolution

### `internal/schema/`
* Schema parsing
* Schema validation
* Property type checking

### `internal/unit/`
* Unit CRUD operations
* Index management
* Uniqueness checks

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
* Creates `bus.yml` and `.bus/` directory

### `SchemaFeature`
* Implements `bus schema init`
* Schema file creation
* Manifest registration

### `UnitFeature`
* Implements schema-specific commands (`bus <schema> add/list/show`)
* Implements compatibility prefix (`bus unit <schema> ...`)
* Unit CRUD operations

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
* New billing rule types can be added

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

