# Entity-Relationship Diagram

This example shows a database schema using D2's SQL table support.

## Pattern Overview

Entity-relationship diagrams document database schemas with:

- Tables and their columns
- Data types
- Constraints (PK, FK, UNQ, NOT NULL)
- Relationships between tables

## Key D2 Techniques

### SQL Table Shape

```d2
users {
  shape: sql_table
  id: uuid { constraint: PK }
  email: varchar(255) { constraint: [UNQ, NOT NULL] }
  name: varchar(100)
  created_at: timestamp
}
```

### Constraint Syntax

Single constraint:

```d2
id: uuid { constraint: PK }
```

Multiple constraints:

```d2
email: varchar(255) { constraint: [UNQ, NOT NULL] }
```

### Foreign Key Relationships

Connect the FK column to the referenced PK:

```d2
orders.user_id -> users.id
order_items.order_id -> orders.id
order_items.product_id -> products.id
```

## Common Constraint Abbreviations

| Full Name | Abbreviation |
|-----------|--------------|
| primary_key | PK |
| foreign_key | FK |
| unique | UNQ |
| not_null | NOT NULL |

## Common Variations

### Add Indexes

```d2
users {
  shape: sql_table
  email: varchar(255) { constraint: [UNQ, INDEX] }
}
```

### Add Table Labels

```d2
users: User Accounts {
  shape: sql_table
  ...
}
```

### Group Related Tables

```d2
auth: Authentication {
  users { shape: sql_table; ... }
  sessions { shape: sql_table; ... }
}

commerce: Commerce {
  orders { shape: sql_table; ... }
  products { shape: sql_table; ... }
}
```

### Add Cardinality Labels

```d2
users.id <- orders.user_id: 1:N
```

## Generate

```bash
# Generate D2 code
d2vision template entity-relationship --d2

# Generate and render
d2vision template er --d2 | d2 - er.svg
```
