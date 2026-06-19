# bus plan

## Active feature plan

- [ ] Add dispatcher release metadata and child-resolution audit support end to end: expose stable text plus JSON version metadata with module name, version, commit, and build time for the `bus` dispatcher, and provide a non-secret way for Services freshness proof to record which `bus-*` executable a dispatcher invocation resolves for commands such as `bus api` and `bus integration`. Preserve dispatcher-first service profiles and first-word dispatch semantics; add unit/e2e coverage for metadata output and resolution evidence without turning `bus` into a build/install tool.

## E2E coverage gaps
