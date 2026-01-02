# Operational Automation (Tasks / Playbooks)

## What this is
How Bus can represent and run **operational work** (tasks and playbooks) attached to units and services, without moving feature logic into core.

## Concepts (binding)

### Task
A **Task** is an executable unit of work that can be attached to:
- a unit (e.g., a project, team, vendor)
- a service (i.e., something provided/consumed by units)

Minimum fields (conceptual; exact schema TBD):
- `id` (stable identifier)
- `name`
- `target` (unit/service reference)
- `runner` (how it runs; e.g., `"shell"`, `"http"`, `"agent"`)
- `inputs` (parameters)

### Playbook
A **Playbook** is an ordered set of tasks (possibly with conditions) that can be invoked as a single operation.

Minimum fields (conceptual; exact schema TBD):
- `id`
- `name`
- `steps[]` (task references + configuration)

## Design goals (binding)
- Tasks/playbooks MUST be representable as workspace-authored documents (schema-driven).
- Attaching tasks/playbooks MUST NOT require hard-coding per-feature logic into core.
- Execution SHOULD be transport-driven (CLI/HTTP) calling into core-owned interfaces.

## Security + secrets (binding)
- Task/playbook definitions MUST support referencing **secret fields** without embedding plaintext secrets in the workspace.
- See: `docs/spec/secrets.md`.
