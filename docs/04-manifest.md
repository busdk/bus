# Manifest Format: `bus.yml`

The manifest file (`bus.yml`) is the root configuration file for a Bus workspace.

## File Format

```yaml
kind: bus.manifest
version: 1
units: []
```

### Required Fields

#### `kind`
* **MUST** be `bus.manifest`
* Identifies the file type

#### `version`
* **MUST** be `1` in v1
* Future versions may increment this

#### `units`
* List of **unit schema references**
* Each entry defines a registered schema

## Unit Schema References

The `units` array contains schema references:

```yaml
units:
  - name: organization
    path: organization.yml
  - name: server
    path: infra/server.yml
```

### Schema Reference Fields

#### `name`
* Unique identifier for the schema
* Used in CLI commands: `bus <name> add ...`
* **MUST** be unique within `units[]`

#### `path`
* Path to the schema YAML file
* **MUST** be unique within `units[]`
* Resolved relative to the manifest location (current working directory) unless absolute

## Rules

1. `units[].name` must be unique
2. `units[].path` must be unique
3. Paths are resolved relative to `.` (current working directory) unless absolute
4. The manifest is created by `bus init`
5. The manifest is updated when schemas are added via `bus schema init`

## Example

```yaml
kind: bus.manifest
version: 1
units:
  - name: server
    path: server.yml
  - name: user
    path: user.yml
  - name: organization
    path: schemas/org.yml
```

