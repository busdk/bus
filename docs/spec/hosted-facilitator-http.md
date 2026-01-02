# Hosted Facilitator (HTTP + OpenAPI)

## What this is
A future thin HTTP front-end (“busd”) over the same core/module operations as the CLI.

## Binding requirements
- HTTP is a thin front-end: do not implement capability logic twice (CLI and HTTP share core/module operations).
- The server MUST publish an OpenAPI 3.x document (e.g., `/openapi.json`).
- API paths MUST be versioned (e.g., `/v1/...`) and error shapes must be stable.

## High-level endpoint surface (preserved intent)
- `POST /v1/x402/requirements` (generate 402 body)
- `POST /v1/x402/ingest` (ingest proof headers and log transaction)
- `POST /v1/transactions` (record transaction; optional if using unit creation API)
- `GET /v1/micropayments/report` (report/aggregate)
- `GET /v1/settlement/proposal` (future)


