# Finnish Small Business Accounting System (Design Plan)

## What this is
This document defines the domain design for a modular, CLI-first accounting
system for Finnish small businesses, built on Bus principles. It describes the
core modules, data model, compliance constraints, and the initial build scope.

## Goals (binding)
- Double-entry, append-only ledger with a verifiable audit trail.
- Finnish compliance for kirjanpitolaki and KILA guidance (immutability,
  period closing, traceability).
- VAT (ALV) support with rate-aware reporting and auditability.
- CLI-first with parity REST API (OpenAPI).
- Offline-first and VCS-friendly storage (Bus never runs Git operations).
- Multi-tenant support from the start (multiple organizations per workspace).
- Modular boundaries so future payroll/budgeting/bank integrations can plug in.

## Non-goals (v1 intent preserved)
- Full payroll processing, tax withholding, or HR workflows.
- A complete CRM, billing engine, or collections automation.
- Real-time bank API integration (initially import-based).
- Multi-user auth and permissions (future).
- Any mutation of posted entries (corrections are new entries).

## Architecture overview
### Interfaces
- **CLI**: primary interface for accountants and developers.
- **HTTP API**: same operations as CLI, exposed as REST with OpenAPI.

### Core modules (domain)
- **General Ledger**: validates double-entry, enforces append-only rules, and
  supports period close locks.
- **Chart of Accounts**: per-organization account structures and metadata.
- **Invoicing**: sales and purchase invoices, posting to ledger.
- **VAT reporting**: period VAT summaries and audit trail into entries.
- **Bank imports**: ingest bank statements and match to invoices/entries.
- **Products (minimal)**: optional product/service catalog for invoices.
- **Budgeting (structure only)**: store budget lines for later reporting.
- **Organization management**: tenant creation, selection, and isolation.

### Modularity and internal messaging
Modules interact through core-owned interfaces. For example:
- Invoicing posts to the ledger via a ledger interface.
- Bank imports emit a "bank transaction imported" event for matching.
The implementation can be in-process calls for v1, while preserving
boundaries for later services.

## Compliance constraints (binding)
- **Append-only ledger**: posted entries are never edited or deleted.
- **Corrections are entries**: errors are fixed with reversing/adjusting
  entries referencing the original voucher.
- **Period close**: closed periods block new entries dated within them.
- **Audit trail**: entries link back to source documents and forward to reports.
- **Retention and exportability**: data remains locally accessible and
  reportable in human-readable form.

## Data model (core entities)
All entities are stored as schema-defined units (per Bus vocabulary), scoped
by `org_id`.
- **Organization**: `{org_id, name, business_id, fiscal_year_start, ...}`
- **Account**: `{account_id, org_id, code, name, type, parent_account, ...}`
- **Journal Entry**: `{entry_id, org_id, date, description, source, source_id}`
- **Entry Line**: `{entry_id, account_id, debit, credit, memo}`
- **Invoice**: `{invoice_id, org_id, type, date, due_date, status, total}`
- **Invoice Line**: `{invoice_id, description, qty, unit_price, vat_code, ...}`
- **VAT Rate**: `{vat_code, percentage, description, valid_from, valid_to}`
- **Partner**: `{partner_id, org_id, name, type, ...}`
- **Bank Transaction**: `{txn_id, org_id, date, counterparty, amount, ...}`
- **Budget**: `{budget_id, org_id, year, account_id, period, amount}`

## Storage model
- **Backend**: Bus state backends apply (filesystem `.bus/` or database).
- **Offline-first**: no online dependency for basic operations.
- **VCS-friendly**: data stored in deterministic formats for external Git use.
  Bus never runs Git commands (per principles).

## CLI and API shape (v1 intent)
- CLI commands cover organization management, chart of accounts, posting
  entries, invoicing, and reporting.
- REST endpoints mirror CLI operations and expose OpenAPI for integrations.

## Initial implementation scope (weekend MVP)
1. Organization creation and selection.
2. Chart of accounts CRUD (metadata only).
3. Journal entry posting with double-entry validation and append-only storage.
4. Trial balance reporting and simple income statement/balance sheet totals.
5. Sales and purchase invoices that post to the ledger.
6. VAT report summary from ledger postings (by rate).
7. Period close enforcement (block backdated entries).

## Future extensions (non-binding)
- Payroll, fixed assets, and advanced budgeting modules.
- Bank API integrations and automated categorization rules.
- Multi-user auth, roles, and permissions.
