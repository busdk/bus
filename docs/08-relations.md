# Relations Model

v1 "relations" are represented as **typed references** (`ref:<schema>`) plus contract properties (start/end/fee/currency) on the consumer unit(s).

## Design Philosophy

Relations are encoded in unit data (typed references + domain properties) rather than introducing a separate relation entity type in v1. This keeps the data model minimal and schema-driven.

## Relation Representation

### Typed References
Relations use the `ref:<SchemaName>` property type:
```yaml
- name: serverId
  type: ref:server
  required: true
```

This creates a typed reference to a unit from the `server` schema.

### Contract Properties
Contract terms are stored as properties on the consumer unit:
* `contractStart` / `contractEnd` - Time range
* `monthlyFeeCents` - Pricing amount
* `currency` - Currency code

Example:
```yaml
data:
  serverId: 10                    # Reference to provider
  contractStart: "2026-01-01"     # Contract start
  contractEnd: null               # Contract end (null = ongoing)
  monthlyFeeCents: 100            # Amount
  currency: "EUR"                 # Currency
```

## Provider/Consumer Pattern

This matches the provider/consumer contract concept:
* **Provider**: The unit being referenced (e.g., `server`)
* **Consumer**: The unit with the reference (e.g., `user`)
* **Contract Terms**: Encoded in consumer unit properties
* **Time Range**: Defined by start/end date properties

## Transactions are Just Units

If you want to record economic effects, you model them with your own `transaction` schema and create transaction units (append-only).

See [Ledger / Transactions (Schema Pattern)](09-transactions.md) and [Billing (Schema Pattern)](06-billing.md).

## Future Considerations

v1 keeps relations simple:
* Relations are implicit in unit data
* No separate relation entity type
* Contract terms stored on consumer units
  * Transactions can be modeled as append-only units (schema pattern)

Future versions may introduce:
* Explicit relation entities
* Bidirectional relations
* Relation metadata
* More complex contract structures

