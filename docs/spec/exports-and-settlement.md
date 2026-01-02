# Exports + Settlement (Extension Points)

## What this is
Interfaces and module boundaries for producing external artifacts (reports/exports/settlement proposals) from tenant-scoped data without adding business logic to core.

## Extension point interfaces (names are stable contracts)

### `Exporter`
Produces an artifact from tenant-scoped data.

- `ID() string`
- `Export(ctx RequestContext, req ExportRequest) (ExportResult, error)`

### `SettlementProposer`
Computes a settlement proposal from normalized transactions.

- `ID() string`
- `Propose(ctx RequestContext, req ProposalRequest) (ProposalResult, error)`

### `ReportingView`
Defines named report views without hard-coding them into core.

- `Name() string`
- `Run(ctx RequestContext, req ReportRequest) (ReportResult, error)`

## Rule
Alternative implementations are supported:
- multiple exporters/proposers/views can coexist and be selected by id/name.


