# Relations (Typed References)

## What this is
The v1 relations model as schema-driven typed references plus domain properties on the consumer unit.

## Representation (binding)
- Use schema property type: `ref:<SchemaName>`
- Store the referenced unit’s primary id as the value.
- Store contract/relationship terms as additional properties on the consumer unit.

## Provider/consumer convention
- **Provider**: the referenced unit.
- **Consumer**: the unit that holds the reference.

## Scoped uniqueness (binding)
Schema property attribute:
- `uniqueScope: [propA, propB, ...]`

Meaning:
- The property value must be unique per distinct tuple of scope values.

Common multi-tenant pattern:
- `email` unique within `organizationId`:
  - `unique: true`
  - `uniqueScope: [organizationId]`


