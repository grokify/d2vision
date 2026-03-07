# Container Patterns

Containers (also called clusters or boundaries) group related elements and provide visual hierarchy.

## Basic Container

```d2
my_container: My Container {
  node1: Node 1
  node2: Node 2
  node1 -> node2
}
```

## Container with Label

The label defaults to the container ID. Override it:

```d2
api: API Gateway {
  # Label is "API Gateway", ID is "api"
}
```

## Nested Containers

Containers can nest to any depth (though deep nesting impacts performance):

```d2
outer: Outer {
  middle: Middle {
    inner: Inner {
      node: Node
    }
  }
}
```

## Invisible Containers

Use invisible containers for layout purposes without visual clutter.

### The Problem

You want horizontal grouping inside a vertical container:

```
[Container]
┌─────────────────┐
│ [A] [B] [C]     │  ← Want these horizontal
│       ↓        │
│     [DB]        │
└─────────────────┘
```

### The Solution

Create an invisible container with `style.stroke-width: 0` and empty label:

```d2
container: Container {
  direction: down

  # Invisible horizontal group
  services: "" {
    direction: right
    style.stroke-width: 0

    a: Service A
    b: Service B
    c: Service C
  }

  db: Database { shape: cylinder }

  services.a -> db
  services.b -> db
  services.c -> db
}
```

### Why It Works

- `""` (empty string) removes the label
- `style.stroke-width: 0` removes the border
- The container still affects layout (grouping children horizontally)

## Container Direction

Each container can have its own direction:

```d2
# Root flows right
direction: right

left: Left Panel {
  # This flows down
  direction: down
  a -> b -> c
}

right: Right Panel {
  # This also flows down
  direction: down
  x -> y -> z
}
```

## Container Grid

Containers can use grid layout:

```d2
dashboard: Dashboard {
  grid-columns: 3

  widget1: CPU
  widget2: Memory
  widget3: Disk
  widget4: Network
  widget5: Requests
  widget6: Errors
}
```

## Container Styling

### Fill Color

```d2
warning_zone: Warning Zone {
  style.fill: "#fff3cd"
  style.stroke: "#ffc107"

  alert1: Alert 1
  alert2: Alert 2
}
```

### Border Radius

```d2
rounded: Rounded Container {
  style.border-radius: 10
  node: Node
}
```

### Multiple Styles

```d2
styled: Styled Container {
  style.fill: "#e3f2fd"
  style.stroke: "#1976d2"
  style.stroke-width: 2
  style.border-radius: 8

  content: Content
}
```

## Referencing Parent

Use `_` to reference the parent container from within:

```d2
outer: Outer {
  inner: Inner {
    node: Node
    # Reference sibling at parent level
    node -> _.sibling
  }
  sibling: Sibling
}
```

## Connecting to Containers

You can connect to a container itself (not its children):

```d2
client: Client
server: Server {
  api: API
  db: Database
}

# Connects to the server container boundary
client -> server
```

Or to specific children:

```d2
client -> server.api
```

## Container Icons

Add icons to containers:

```d2
aws: AWS {
  icon: https://icons.terrastruct.com/aws%2F_Group%20Icons%2FAWS-Cloud-alt_light-bg.svg

  ec2: EC2 Instances
  rds: RDS Database
}
```

## Pattern: Network Zones

```d2
grid-columns: 2

public: Public Zone {
  style.fill: "#e8f5e9"
  direction: down

  lb: Load Balancer
  web: Web Servers
  lb -> web
}

private: Private Zone {
  style.fill: "#fff3e0"
  direction: down

  app: App Servers
  db: Database { shape: cylinder }
  app -> db
}

public.web -> private.app
```

## Pattern: Layered Architecture

```d2
direction: down

presentation: Presentation Layer {
  direction: right
  web: Web UI
  mobile: Mobile App
  api: API Gateway
}

business: Business Layer {
  direction: right
  auth: Auth Service
  orders: Order Service
  inventory: Inventory Service
}

data: Data Layer {
  direction: right
  cache: Redis { shape: cylinder }
  db: PostgreSQL { shape: cylinder }
  queue: RabbitMQ { shape: queue }
}

presentation -> business
business -> data
```
