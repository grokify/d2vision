# Sequence Diagrams

D2 supports sequence diagrams for showing interactions over time.

## Basic Structure

```d2
my_sequence: {
  shape: sequence_diagram

  alice: Alice
  bob: Bob

  alice -> bob: Hello
  bob -> alice: Hi there
}
```

## Actors

Actors are the participants in the sequence. Define them explicitly to control order:

```d2
sequence: {
  shape: sequence_diagram

  # Actors appear left-to-right in definition order
  user: User { shape: person }
  client: Client App
  server: Server
  database: Database
}
```

### Actor Shapes

```d2
sequence: {
  shape: sequence_diagram

  user: User { shape: person }
  service: Service { shape: rectangle }  # default
  db: Database { shape: cylinder }
}
```

## Messages

Messages flow between actors in the order defined:

```d2
sequence: {
  shape: sequence_diagram

  client: Client
  server: Server

  client -> server: Request
  server -> client: Response
}
```

### Message Labels

```d2
client -> server: POST /api/login
server -> db: SELECT * FROM users
db -> server: User data
server -> client: 200 OK + JWT
```

### Self Messages

An actor can send a message to itself:

```d2
sequence: {
  shape: sequence_diagram

  server: Server

  server -> server: Validate input
  server -> server: Process request
}
```

## Spans (Activation Boxes)

Spans show when an actor is active:

```d2
sequence: {
  shape: sequence_diagram

  client: Client
  server: Server

  client -> server: Request
  server.process: Processing {
    server -> server: Validate
    server -> server: Transform
  }
  server -> client: Response
}
```

## Groups (Fragments)

Groups label sections of the sequence, useful for conditionals, loops, or alternatives:

```d2
sequence: {
  shape: sequence_diagram

  client: Client
  server: Server

  client -> server: Login

  success: Success Case {
    server -> client: 200 OK
  }

  failure: Failure Case {
    server -> client: 401 Unauthorized
  }
}
```

### Common Group Types

```d2
sequence: {
  shape: sequence_diagram

  a: A
  b: B

  # Optional
  opt: opt [condition] {
    a -> b: Maybe
  }

  # Loop
  loop: loop [while condition] {
    a -> b: Repeat
  }

  # Alternative
  alt: alt [if condition] {
    a -> b: Then
  }

  else: else {
    a -> b: Otherwise
  }
}
```

## Notes

Add notes to actors:

```d2
sequence: {
  shape: sequence_diagram

  server: Server

  server.note: Rate limited to 100 req/min

  server -> server: Process
}
```

## Complete Example: Authentication Flow

```d2
auth_flow: Authentication Flow {
  shape: sequence_diagram

  # Actors
  user: User { shape: person }
  browser: Browser
  server: Auth Server
  db: User Database

  # Happy path
  user -> browser: Enter credentials
  browser -> server: POST /login
  server -> db: Query user
  db -> server: User record

  # Validation
  server.validate: Validating {
    server -> server: Check password
    server -> server: Generate JWT
  }

  server -> browser: 200 OK + token
  browser -> user: Login success

  # Error case
  error: Invalid Credentials {
    server -> browser: 401 Unauthorized
    browser -> user: Show error
  }
}
```

## Best Practices

### 1. Define Actors First

Define all actors at the top for consistent ordering:

```d2
sequence: {
  shape: sequence_diagram

  # All actors first
  a: Actor A
  b: Actor B
  c: Actor C

  # Then messages
  a -> b -> c
}
```

### 2. Use Meaningful Labels

```d2
# Good
client -> server: POST /api/users { name: "John" }

# Less good
client -> server: request
```

### 3. Group Related Messages

```d2
sequence: {
  shape: sequence_diagram

  client: Client
  server: Server
  db: Database

  # Authentication phase
  auth: Authentication {
    client -> server: Login
    server -> db: Verify
    db -> server: OK
    server -> client: Token
  }

  # Data fetch phase
  fetch: Fetch Data {
    client -> server: GET /data
    server -> db: Query
    db -> server: Results
    server -> client: Response
  }
}
```

### 4. Keep It Simple

Sequence diagrams can get complex. Consider splitting into multiple diagrams for:

- Different scenarios (happy path, error cases)
- Different phases (auth, processing, cleanup)
- Different detail levels (overview, detailed)

## Limitations

- No parallel/concurrent message support
- Limited styling options within sequence diagrams
- Actor positions are fixed (can't overlap or reorder mid-diagram)
