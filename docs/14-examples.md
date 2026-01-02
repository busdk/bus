# Example Workflow

This document shows a complete example workflow using Bus CLI v1.

The example models a tiny **ledger** using only schemas + units:
* `currency` (so “EUR-only” is just “only create EUR”)
* `account`
* `transaction` (append-only)

## 1. Initialize Workspace

```bash
bus init
```

**Creates:**
* `./bus.yml` (empty manifest; default YAML)
* `./.bus/` directory

You may also initialize in TOML or JSON (see `16-multi-format-storage.md`):

```bash
bus init --format toml
```

## 2. Define Schemas

### Create Currency Schema

```bash
bus schema init currency \
  --property id:string,primary,required,unique \
  --property decimals:int,required
```

### Create Account Schema

```bash
bus schema init account \
  --property id:uuid,primary,required,unique \
  --property name:string,required \
  --property currencyId:ref:currency,required
```

### Create Transaction Schema

```bash
bus schema init transaction \
  --property id:uuid,primary,required,unique \
  --property accountId:ref:account,required \
  --property postedAt:date,required \
  --property deltaCents:int,required \
  --property currencyId:ref:currency,required \
  --property note:string
```

### Make Transactions Create-Only (Append-Only Ledger)

Edit the transaction schema file (`transaction.<ext>`) and add:

```yaml
operations:
  create: true
  list: true
  show: true
  update: false
  delete: false
```

## 3. Add Data

### Add EUR Currency (EUR-only)

```bash
bus currency add id=EUR decimals=2
```

### Add an Account

```bash
bus account add name="Main" currencyId=EUR
```

Bus prints the created account ID; you’ll use it for transactions.

### Post a Transaction (Append-Only)

```bash
bus transaction add accountId=<ACCOUNT_ID> postedAt=2026-01-01 deltaCents=1000 currencyId=EUR note="January fee"
```

To correct a mistake, create a reversing transaction instead of editing:

```bash
bus transaction add accountId=<ACCOUNT_ID> postedAt=2026-01-02 deltaCents=-1000 currencyId=EUR note="Reversal"
```

## 4. Query Units

### List Transactions

```bash
bus transaction list
```

### Show a Transaction

```bash
bus transaction show <TX_ID>
```

## Result

You now have a small ledger built entirely from normal schemas + units, with “append-only” and “EUR-only” enforced by convention and schema constraints.

## 5. Micropayments + x402 Capture (End-to-End; Spec-Only)

This example shows:
- x402 configuration as normal units
- generating a 402 response body
- ingesting proof headers
- the resulting normalized `transaction` record

See `17-micropayments.md` and `18-x402.md` for the schema and wire specs.

### Example: YAML Workspace

Assume you have schemas registered for:
- `business_unit` (payer/payee units)
- `service`
- `x402_policy`
- `x402_accept`
- `transaction`

Create the config and identities (illustrative):

```bash
bus business_unit add id=unit_client name="Client BU"
bus business_unit add id=unit_service name="Service BU"

bus service add id=svc_widgets name="Widgets API" resource="GET /v1/widgets"

bus x402_policy add id=pol_widgets serviceId=svc_widgets enabled=true x402Version=1 message="Pay per request"

bus x402_accept add \
  id=acc_widgets \
  policyId=pol_widgets \
  scheme=exact \
  network=testnet \
  maxAmountRequired=100 \
  payTo="merchant:example" \
  asset=USD \
  description="Pay-per-request" \
  maxTimeoutSeconds=30
```

Generate the 402 JSON body:

```bash
bus x402 402 --service svc_widgets
```

Ingest a successful client retry (headers are base64 JSON; values below are placeholders):

```bash
bus x402 ingest \
  --service svc_widgets \
  --x-payment <BASE64_JSON> \
  --x-payment-response <BASE64_JSON>
```

Result: Bus logs a normalized `transaction` unit (illustrative YAML form):

```yaml
kind: bus.unit
version: 1
schema: transaction
data:
  id: "mtx-2026-01-02-0001"
  createdAt: "2026-01-02T12:00:00Z"
  fromUnitId: unit_client
  toUnitId: unit_service
  serviceId: svc_widgets
  amount: "100"
  asset: "USD"
  network: testnet
  sourceType: x402
  sourceRef: "tx_abc123"
  proofHash: "sha256:..."
```

The same logical record can be stored in JSON in a JSON workspace (illustrative):

```json
{
  "kind": "bus.unit",
  "version": 1,
  "schema": "transaction",
  "data": {
    "id": "mtx-2026-01-02-0001",
    "createdAt": "2026-01-02T12:00:00Z",
    "fromUnitId": "unit_client",
    "toUnitId": "unit_service",
    "serviceId": "svc_widgets",
    "amount": "100",
    "asset": "USD",
    "network": "testnet",
    "sourceType": "x402",
    "sourceRef": "tx_abc123",
    "proofHash": "sha256:..."
  }
}
```

### Same Logical Config in TOML

The same unit record can be stored as TOML in a TOML workspace (illustrative):

```toml
kind = "bus.unit"
version = 1
schema = "x402_policy"

[data]
id = "pol_widgets"
serviceId = "svc_widgets"
enabled = true
x402Version = 1
message = "Pay per request"
```

## 6. Central Bus Service Narrative (Future; Spec-Only)

In a future hosted facilitator mode:
- A central Bus server maintains the shared transaction ledger for multiple producers.
- Services can submit transactions to the facilitator over HTTP.
- The CLI remains script-first and can operate:
  - locally (file-backed core), or
  - remotely (`--remote <URL>`, spec-only), calling the same operations over HTTP.

