# Network Clusters Example

Side-by-side network boundary clusters with services and datastores.

## Layout Challenge

Getting clusters to align horizontally with tops aligned is tricky with D2/ELK. Cross-cluster edges cause vertical stacking.

## Solution

Use `grid-columns: 2` at the root level to force side-by-side layout.

## Key Patterns

```d2
# Force horizontal layout at root
grid-columns: 2

cluster1: Cluster 1 {
  direction: down  # Stack children vertically within cluster

  # Invisible container for horizontal service alignment
  services: "" {
    direction: right
    style.stroke-width: 0  # Hide the container border

    service1a: Service 1A
    service1b: Service 1B
  }

  datastore1: DataStore 1 {
    shape: cylinder
  }

  # Connect services to datastore
  services.service1a -> datastore1
  services.service1b -> datastore1
}
```

## D2 Layout Engine Insights

Understanding how D2's layout engines (ELK/dagre) handle cross-cluster edges:

### Why Cross-Cluster Edges Cause Vertical Stacking

**Cross-cluster edges are the main layout disruptor.** The `DataStore2 -> DataStore1` edge caused ELK and dagre to place clusters vertically because the layout engine tries to route the edge cleanly by stacking source above/below target.

### Why `grid-columns` Works

**`grid-columns: 2` bypasses the layout engine for top-level placement.** It's a CSS-grid-style override that forces two columns regardless of edges. The layout engine only runs *inside* each cluster, not at the grid level.

### Why `direction: right` Alone Doesn't Work

`grid-columns`/`grid-rows` is the right tool for side-by-side clusters — but it only works at the top level when there are no cross-cluster edges fighting the layout engine. Once you add a cross-cluster edge, dagre treats it as a graph dependency and stacks nodes vertically regardless of direction hints.

### The Invisible Container Trick

**The nested invisible container trick** (`services: "" { style.stroke-width: 0 }`) is a clean way to group nodes horizontally inside a `direction: down` cluster without adding a visible box.

### Alternative Approaches (Less Recommended)

- **Invisible spacer nodes** are fragile — trying to equalize heights manually is brittle and doesn't scale. The grid approach is far cleaner.
- **TALA** (D2's paid layout engine) offers better alignment controls but isn't free.
- **ELK** is better than dagre for complex diagrams but still doesn't expose enough alignment controls through D2's wrapper to solve top-alignment with cross-cluster edges.

## Files

- `clusters.d2` - D2 source
- `clusters.svg` - Rendered output

## Regenerate

```bash
d2 clusters.d2 clusters.svg
```

## Generate with d2vision

```bash
# Generate from template
d2vision template network-boundary --d2 > clusters.d2

# Customize: 3 clusters with 4 services each
d2vision template network-boundary --clusters 3 --services 4 --d2 > clusters.d2

# Full pipeline
d2vision template network-boundary --d2 | d2 - clusters.svg
```
