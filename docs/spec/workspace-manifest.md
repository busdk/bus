# Workspace + Manifest

## What this is
Workspace discovery, manifest candidates, and path resolution rules.

## Manifest candidates (binding)
In the current working directory, manifest candidates are:
- `bus.yml`
- `bus.yaml`
- `bus.toml`
- `bus.json`

Rules:
- If **exactly one** candidate exists: load it.
- If **none** exist: commands that require a manifest MUST error; `bus init` MAY create one.
- If **more than one** exists: Bus MUST error and list the candidates.

## Manifest minimal fields (binding)
The manifest is a document with:
- `kind: bus.manifest`
- `version: 1`

## Path resolution (binding)
- Paths stored in the manifest are resolved relative to the manifest location unless absolute.

## Directory constraints (binding)
- Bus MUST NOT impose tool-defined top-level directories (e.g., `schemas/`, `units/`) outside `.bus/`.
- Bus MAY create and own only `.bus/` and its subdirectories.


