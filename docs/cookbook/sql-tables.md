# SQL Tables

D2 supports SQL table shapes for entity-relationship diagrams.

## Basic Table

```d2
users {
  shape: sql_table

  id: int
  name: varchar(100)
  email: varchar(255)
  created_at: timestamp
}
```

## Constraints

Add constraints to columns:

```d2
users {
  shape: sql_table

  id: int { constraint: primary_key }
  email: varchar(255) { constraint: unique }
  org_id: int { constraint: foreign_key }
}
```

### Constraint Abbreviations

D2 recognizes and abbreviates common constraints:

| Constraint | Abbreviation |
|------------|--------------|
| `primary_key` | PK |
| `foreign_key` | FK |
| `unique` | UNQ |

### Multiple Constraints

```d2
users {
  shape: sql_table

  id: int { constraint: [primary_key] }
  email: varchar(255) { constraint: [unique, not_null] }
}
```

## Foreign Key Relationships

Connect tables via foreign keys:

```d2
users {
  shape: sql_table
  id: int { constraint: primary_key }
  name: varchar(100)
}

orders {
  shape: sql_table
  id: int { constraint: primary_key }
  user_id: int { constraint: foreign_key }
  total: decimal(10,2)
}

orders.user_id -> users.id
```

The arrow points from the FK column to the PK column it references.

## Complete Example: E-Commerce Schema

```d2
direction: right

users {
  shape: sql_table
  id: uuid { constraint: [primary_key] }
  email: varchar(255) { constraint: [unique, not_null] }
  name: varchar(100)
  password_hash: varchar(255)
  created_at: timestamp
  updated_at: timestamp
}

products {
  shape: sql_table
  id: uuid { constraint: [primary_key] }
  name: varchar(255) { constraint: not_null }
  description: text
  price: decimal(10,2) { constraint: not_null }
  stock: int
  created_at: timestamp
}

orders {
  shape: sql_table
  id: uuid { constraint: [primary_key] }
  user_id: uuid { constraint: foreign_key }
  status: varchar(20)
  total: decimal(10,2)
  created_at: timestamp
}

order_items {
  shape: sql_table
  id: uuid { constraint: [primary_key] }
  order_id: uuid { constraint: foreign_key }
  product_id: uuid { constraint: foreign_key }
  quantity: int
  price: decimal(10,2)
}

# Relationships
orders.user_id -> users.id
order_items.order_id -> orders.id
order_items.product_id -> products.id
```

## Table Labels

Add a descriptive label:

```d2
user_accounts: User Accounts {
  shape: sql_table
  id: int { constraint: primary_key }
  username: varchar(50)
}
```

## Grouping Tables

Group related tables in containers:

```d2
auth: Authentication {
  users {
    shape: sql_table
    id: int { constraint: primary_key }
    email: varchar(255)
  }

  sessions {
    shape: sql_table
    id: int { constraint: primary_key }
    user_id: int { constraint: foreign_key }
    token: varchar(255)
    expires_at: timestamp
  }

  sessions.user_id -> users.id
}

commerce: Commerce {
  products {
    shape: sql_table
    id: int { constraint: primary_key }
    name: varchar(255)
  }

  orders {
    shape: sql_table
    id: int { constraint: primary_key }
    product_id: int { constraint: foreign_key }
  }

  orders.product_id -> products.id
}

# Cross-group relationship
commerce.orders.user_id -> auth.users.id
```

## Common Column Types

```d2
example {
  shape: sql_table

  # Numeric
  id: int
  price: decimal(10,2)
  quantity: smallint
  big_number: bigint

  # String
  name: varchar(100)
  description: text
  code: char(3)

  # Date/Time
  created_at: timestamp
  birth_date: date
  start_time: time

  # Other
  is_active: boolean
  data: jsonb
  uuid: uuid
  tags: varchar[]
}
```

## One-to-Many Relationships

```d2
departments {
  shape: sql_table
  id: int { constraint: primary_key }
  name: varchar(100)
}

employees {
  shape: sql_table
  id: int { constraint: primary_key }
  dept_id: int { constraint: foreign_key }
  name: varchar(100)
}

# One department has many employees
employees.dept_id -> departments.id
```

## Many-to-Many Relationships

Use a junction/join table:

```d2
students {
  shape: sql_table
  id: int { constraint: primary_key }
  name: varchar(100)
}

courses {
  shape: sql_table
  id: int { constraint: primary_key }
  title: varchar(200)
}

# Junction table
enrollments {
  shape: sql_table
  student_id: int { constraint: [primary_key, foreign_key] }
  course_id: int { constraint: [primary_key, foreign_key] }
  enrolled_at: timestamp
}

enrollments.student_id -> students.id
enrollments.course_id -> courses.id
```

## Self-Referencing Relationships

```d2
employees {
  shape: sql_table
  id: int { constraint: primary_key }
  name: varchar(100)
  manager_id: int { constraint: foreign_key }
}

# Employee reports to another employee
employees.manager_id -> employees.id
```

## Best Practices

1. **Use meaningful table names** - Plural nouns (users, orders, products)

2. **Consistent ID naming** - Either `id` everywhere or `table_id` pattern

3. **Show only relevant columns** - Don't include every column for documentation

4. **Group related tables** - Use containers to organize by domain

5. **Label relationships** - Add labels to edges when the relationship isn't obvious

```d2
orders.user_id -> users.id: placed by
```
