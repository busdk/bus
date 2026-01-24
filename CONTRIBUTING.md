# Contributing to Bus

Thanks for helping build Bus. This repository is in a **planning-first** stage,
so documentation changes are the primary contributions right now.

## How to contribute
1. Open or find a GitHub issue describing the change:
   `https://github.com/hyperifyio/bus/issues`
2. Create a branch from `main`.
3. Make focused changes that match the roadmap and spec docs.
4. Open a pull request and link the issue using the full URL.

## Traceability requirements (binding)
- Every change must link to a canonical issue using the **full URL**.
- Tests and implementation comments must reference the issue URL.
- If a change cannot include tests, it must be approved and tracked by a
  follow-up issue (full URL) with explicit rationale.

## Tests
When implementation starts, tests are required for every behavior. Follow the
definition of done documented in the repository rules: tests must be
deterministic, cover new lines and branches, and include unit + integration
coverage where applicable.

## Documentation changes
- Keep `README.md` (repo front page) aligned with `docs/README.md`.
- Add new spec documents under `docs/spec/` and link them from `docs/README.md`.
- Add new roadmap steps under `docs/roadmap/` and link them from `docs/README.md`.

## Code of conduct
By participating, you agree to follow `CODE_OF_CONDUCT.md`.
