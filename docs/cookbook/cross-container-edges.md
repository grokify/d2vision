# Cross-Container Edges

Edges between containers are the most common source of layout problems in D2.

## The Problem

You have two containers and want an edge between their children:

```d2
cluster1: Cluster 1 {
  service: Service
}

cluster2: Cluster 2 {
  db: Database
}

cluster1.service -> cluster2.db
```

**Expected**: Side-by-side containers with an edge between them.

**Actual**: Containers stack vertically.

## Why This Happens

D2's layout engine (ELK/dagre) optimizes for edge routing:

1. It sees an edge from cluster1 to cluster2
2. It calculates that a vertical arrangement produces a shorter, straighter edge
3. It stacks the containers vertically

## Solution 1: grid-columns (Recommended)

Force horizontal layout at the root level:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  service: Service
}

cluster2: Cluster 2 {
  db: Database
}

cluster1.service -> cluster2.db
```

Now the containers are in a 2-column grid, and the edge routes between them.

## Solution 2: Explicit Dimensions

Force containers to specific positions with width/height:

```d2
cluster1: Cluster 1 {
  width: 200
  service: Service
}

cluster2: Cluster 2 {
  width: 200
  db: Database
}

cluster1.service -> cluster2.db
```

Less reliable than `grid-columns` but sometimes useful.

## Multiple Cross-Container Edges

With multiple edges, `grid-columns` becomes essential:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  direction: down
  service1: Service 1
  service2: Service 2
  db1: Database { shape: cylinder }
}

cluster2: Cluster 2 {
  direction: down
  service3: Service 3
  db2: Database { shape: cylinder }
}

# Multiple cross-container edges
cluster1.service1 -> cluster2.service3
cluster1.db1 -> cluster2.db2: sync
cluster2.service3 -> cluster1.db1: query
```

## Edge Labels

Cross-container edges can have labels:

```d2
grid-columns: 2

source: Source {
  producer: Producer
}

target: Target {
  consumer: Consumer
}

source.producer -> target.consumer: events
```

## Bidirectional Edges

For bidirectional communication:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  db1: Database { shape: cylinder }
}

cluster2: Cluster 2 {
  db2: Database { shape: cylinder }
}

cluster1.db1 <-> cluster2.db2: sync
```

## Deep Nesting

Cross-container edges work with nested containers:

```d2
grid-columns: 2

region1: Region 1 {
  vpc1: VPC 1 {
    subnet1: Subnet 1 {
      instance1: EC2
    }
  }
}

region2: Region 2 {
  vpc2: VPC 2 {
    subnet2: Subnet 2 {
      instance2: EC2
    }
  }
}

region1.vpc1.subnet1.instance1 -> region2.vpc2.subnet2.instance2: VPN
```

## Pattern: Hub and Spoke

Central service with connections to multiple others:

```d2
grid-columns: 3
grid-rows: 3

# Top row
_.1: "" { style.stroke-width: 0 }
service1: Service 1 {
  api: API
}
_.2: "" { style.stroke-width: 0 }

# Middle row
service2: Service 2 {
  api: API
}
hub: Hub {
  router: Router
}
service3: Service 3 {
  api: API
}

# Bottom row
_.3: "" { style.stroke-width: 0 }
service4: Service 4 {
  api: API
}
_.4: "" { style.stroke-width: 0 }

# Connections
service1.api -> hub.router
service2.api -> hub.router
service3.api -> hub.router
service4.api -> hub.router
```

## Pattern: Replication

Database replication across clusters:

```d2
grid-columns: 2

primary: Primary Region {
  direction: down
  app: Application
  db: Primary DB { shape: cylinder }
  app -> db
}

secondary: Secondary Region {
  direction: down
  app: Application
  db: Replica DB { shape: cylinder }
  app -> db
}

primary.db -> secondary.db: async replication
```

## Troubleshooting

### Edges Crossing Unexpectedly

If edges cross through containers, try:

1. Reorder container definitions (defined order affects layout)
2. Add intermediate nodes
3. Use explicit grid positioning

### Edges Too Long

If edges route around containers oddly:

1. Ensure `grid-columns` matches your container count
2. Check container sizes aren't forcing strange routing
3. Consider simplifying the structure

### Still Stacking Vertically

If containers still stack despite `grid-columns`:

1. Verify `grid-columns` is at the root level (not inside a container)
2. Check for syntax errors
3. Ensure there are enough elements to fill the grid
