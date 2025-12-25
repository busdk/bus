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
* `./bus.yml` (empty manifest)
* `./.bus/` directory

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

Edit `transaction.yml` and add:

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

