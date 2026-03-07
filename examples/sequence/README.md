# Sequence Diagram

This example shows an authentication flow using D2's sequence diagram support.

## Pattern Overview

Sequence diagrams show interactions between actors over time. They're ideal for documenting:

- Authentication/authorization flows
- API request/response patterns
- Multi-service transactions
- Error handling paths

## Key D2 Techniques

### Sequence Diagram Shape

```d2
auth_flow: Authentication Flow {
  shape: sequence_diagram
  ...
}
```

### Actor Definition

Actors can be explicitly defined to control display order:

```d2
user: User { shape: person }
client: Client App
gateway: API Gateway
auth: Auth Service
db: User DB
```

### Message Order

Messages appear in the order they're defined:

```d2
user -> client: Enter credentials
client -> gateway: POST /login
gateway -> auth: Validate token
auth -> db: Query user
db -> auth: User data
auth -> gateway: JWT token
gateway -> client: 200 OK + token
client -> user: Login success
```

### Groups (Fragments)

Groups label subsets of the sequence for conditional flows:

```d2
error_case: Invalid Credentials {
  auth -> gateway: 401 Unauthorized
  gateway -> client: 401 Error
}
```

## Common Variations

### Add Spans (Activation Boxes)

```d2
auth {
  validate: Validating {
    auth -> db: Query
    db -> auth: Result
  }
}
```

### Add Notes

```d2
gateway.note: Rate limited to 100 req/min
```

### Add Self Messages

```d2
auth -> auth: Generate JWT
```

### Add Async Messages

```d2
gateway --> queue: Enqueue audit log
```

## Generate

```bash
# Generate D2 code
d2vision template sequence --d2

# Generate and render
d2vision template sequence --d2 | d2 - sequence.svg
```
