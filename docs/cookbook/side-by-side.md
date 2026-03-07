# Side-by-Side Layouts

One of the most common layout challenges is getting containers to appear side-by-side instead of stacked vertically.

## The Problem

You want this:

```
[Cluster 1]  [Cluster 2]
```

But D2 gives you this:

```
[Cluster 1]
[Cluster 2]
```

## Why This Happens

D2's layout engine optimizes for edge routing. Without explicit constraints, it positions containers to minimize edge lengths and crossings. Vertical stacking often produces cleaner edges.

## Solution: grid-columns

The most reliable solution is `grid-columns` at the root level:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  service1: Service 1
  db1: Database
}

cluster2: Cluster 2 {
  service2: Service 2
  db2: Database
}
```

This forces the layout engine to place containers in a 2-column grid.

## With Cross-Container Edges

Cross-container edges are the main cause of vertical stacking. Even with `grid-columns`, they work correctly:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  service1: Service 1
  db1: Database { shape: cylinder }
  service1 -> db1
}

cluster2: Cluster 2 {
  service2: Service 2
  db2: Database { shape: cylinder }
  service2 -> db2
}

# Cross-cluster edge - this would cause stacking without grid-columns
cluster2.db2 -> cluster1.db1: replication
```

## Aligning Tops

By default, grid cells align at the top. This is usually what you want for side-by-side containers.

If containers have different heights, they still align at the top:

```d2
grid-columns: 2

short: Short {
  a
}

tall: Tall {
  b
  c
  d
  e
}
```

## Three or More Containers

Just increase `grid-columns`:

```d2
grid-columns: 3

dev: Development { ... }
staging: Staging { ... }
prod: Production { ... }
```

## Mixed Layout: Horizontal + Vertical

For complex layouts, combine grid at the root with direction in containers:

```d2
grid-columns: 2

# Left column: vertical stack
left: Left Side {
  direction: down
  top: Top Section { ... }
  bottom: Bottom Section { ... }
}

# Right column: another vertical stack
right: Right Side {
  direction: down
  top: Top Section { ... }
  bottom: Bottom Section { ... }
}
```

## Alternative: direction: right

For simple cases without cross-container edges, `direction: right` works:

```d2
direction: right

a: Box A
b: Box B
c: Box C
```

But this breaks down with complex edges between containers.

## When NOT to Use grid-columns

- **Sequence diagrams**: They have their own layout rules
- **Simple linear flows**: `direction: right` with `a -> b -> c` works fine
- **When you want the layout engine to decide**: Sometimes automatic positioning is better

## Complete Example

```d2
# Force side-by-side at root
grid-columns: 2

cluster1: Cluster 1 {
  # Vertical flow inside
  direction: down

  # Horizontal service grouping
  services: "" {
    direction: right
    style.stroke-width: 0

    service1a: Service 1A
    service1b: Service 1B
  }

  datastore1: DataStore 1 {
    shape: cylinder
  }

  services.service1a -> datastore1
  services.service1b -> datastore1
}

cluster2: Cluster 2 {
  direction: down

  service2: Service 2

  datastore2: DataStore 2 {
    shape: cylinder
  }

  service2 -> datastore2
}

# Cross-cluster replication
cluster2.datastore2 -> cluster1.datastore1: replication
```

This produces two side-by-side clusters with aligned tops, each containing vertically-arranged services and datastores.
