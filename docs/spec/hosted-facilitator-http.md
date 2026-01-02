# Hosted Facilitator (HTTP + OpenAPI)

## What this is
A future thin HTTP front-end (“busd”) over the same core operations as the CLI.

## Binding requirements
- HTTP is a thin front-end: do not duplicate *core operations* (CLI and HTTP share the same core primitives).
- HTTP MAY include HTTP-only features (e.g., x402 request/response wire handling) that are not part of core.
- The server MUST publish an OpenAPI 3.x document (e.g., `/openapi.json`).
- API paths MUST be versioned (e.g., `/v1/...`) and error shapes must be stable.

## High-level endpoint surface (preserved intent)
- `POST /v1/x402/requirements` (generate 402 body)
- `POST /v1/x402/ingest` (ingest proof headers and log transaction)
- `POST /v1/transactions` (record transaction; optional if using unit creation API)
- `GET /v1/micropayments/report` (report/aggregate)
- `GET /v1/settlement/proposal` (future)


