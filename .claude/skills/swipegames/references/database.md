# Database Style Guide

## Overview
Simplified database architecture for gaming platform MVP and 6-month scale focusing on essential functionality.

**Scale Target**: 10+ million transactions in 6 months with partitioned design from day one.

## General Rules

### Migrations
- Don't use down migrations in goose (leave them empty)
- Use consistent naming for indexes and constraints
- First migration for every table should be `create-triggers.sql` for `updated_at` trigger

### Table Naming
- Use singular form: `user` NOT `users`
- For collections use plural: `free_rounds`
- For many-to-many joining tables use plural: `user_roles`, `game_providers`

### Naming Conventions

**Indexes** (regular):
```
<table-name>_<row_names>_idx
```

**Unique indexes**:
```
<table-name>_<row_names>_uq
```

**Foreign keys**:
```
<table-name>_<row_names>_<ext_table_name>_<ext_row_name>_fk
```

### Required Fields
- Every table: `created_at` and `updated_at` fields
- `updated_at` automatically updated by trigger
- Soft delete: use `deleted_at` field
- Always have `primary key` (prefer `uuid` type)
- Always have `foreign keys` for references
- Create `indexes` for query columns (can add later if needed)

### Performance
- Denormalized data acceptable for better performance

## Tables Partitioning

For heavy tables use `pg_partman` for partitioning.

### Creating Partitioned Table

```sql
select partman.create_parent(
    p_parent_table := '<schema_name>.<table_name>',
    p_control := 'created_at',
    p_interval := '3 months',
    p_premake := 3,
    p_start_partition := date_trunc('month', CURRENT_TIMESTAMP)::text
);
```

**Important**:
- Script in same migration as table creation
- `created_at` field must be in table's primary key
- 3-month partitions for unique TxID guarantee and monthly reports

### Partition Pruning

Always add `created_at` filter when querying partitioned tables:

**Filter ranges by use case**:
- **Ledger transactions**: 1 month (bonus transactions max period)
- **Game sessions**: 1 day (sessions typically 4 hours)
- **Rounds and configs**: 3 months
- **Integration adapters**: 3 months (for reconciliation)

### Test Data and Partition Pruning

For test data (fixtures), bypass partition pruning:
- Maintain list of test data IDs in repository layer
- If requested ID in test data list, skip `created_at` filter

### Uniqueness Constraints

With partitioning by `created_at`:
- Hard to enforce uniqueness across partitions
- Rely only on `UUID uniqueness`
- Single CID can have non-unique transaction IDs across partitions

## Migrations and Fixtures

### Location
- `db/` folder per service contains migrations and fixtures
- Automatically applied on service start

### Management
- When migrations/fixtures applied to all environments, aggregate into single file
- Remove old ones to keep count under control

## Currency Handling

### Fiat Currencies

**2 decimal places** (most currencies):
- 1 cent = 0.01 main units
- Store 1 in DB = 0.01 main unit
- Display: 0.01
- No conversion needed for integration

**3 decimal places** (some currencies):
- 1 cent = 0.001 main units
- Store 1 in DB = 10 subunits
- Display: 0.01 (use 1/100 equation)
- Integration: same 1/100 equation

### Crypto Currencies
- Full decimal places support
- Internally maintained in table

## Database Commands

```bash
# Generate database models after schema changes
make gen-db <service_name>

# Create new migration
make add-migration <service_name>

# Database changes require full reset
make down  # Reset schema before applying changes
```

## Best Practices

### Schema Changes
1. Create migration in `db/migrations/`
2. Include triggers in first migration
3. Apply migration with `make down` then `make up`
4. Generate models with `make gen-db`
5. Create fixtures if needed

### Query Optimization
- Always use partition pruning filters
- Create indexes for frequently queried columns
- Use denormalization when beneficial for reads
- Monitor query performance with logging

### Data Integrity
- Always use foreign keys
- Use unique indexes where needed
- Consider partition boundaries for uniqueness
- UUID primary keys for global uniqueness
