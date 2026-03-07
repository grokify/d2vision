# Layout Fundamentals

Understanding how D2's layout engine works is key to creating well-structured diagrams.

## Direction

The `direction` keyword controls how children are arranged within a container.

```d2
direction: right  # Children flow left-to-right (default for root)
direction: down   # Children flow top-to-bottom
direction: left   # Children flow right-to-left
direction: up     # Children flow bottom-to-top
```

### Example: Horizontal Flow

```d2
direction: right

a -> b -> c
```

```
[a] → [b] → [c]
```

### Example: Vertical Flow

```d2
direction: down

a -> b -> c
```

```
[a]
 ↓
[b]
 ↓
[c]
```

### Direction in Containers

Each container can have its own direction:

```d2
direction: right

left_box: {
  direction: down
  a -> b -> c
}

right_box: {
  direction: down
  x -> y -> z
}

left_box -> right_box
```

## Grid Layout

Grid layout gives you explicit control over positioning.

### grid-columns

Forces elements into a grid with N columns:

```d2
grid-columns: 3

a
b
c
d
e
f
```

Result:

```
[a] [b] [c]
[d] [e] [f]
```

### grid-rows

Forces elements into a grid with N rows:

```d2
grid-rows: 2

a
b
c
d
e
f
```

Result:

```
[a] [b] [c]
[d] [e] [f]
```

### Combining Grid and Direction

When both are set, the first one determines fill order:

```d2
grid-columns: 2
grid-rows: 3

# Fills columns first (left-to-right, then down)
a; b; c; d; e; f
```

## How the Layout Engine Works

D2's layout engine (ELK or dagre) follows these steps:

1. **Parse** the D2 code into a graph structure
2. **Apply constraints** (direction, grid, dimensions)
3. **Optimize** edge routing to minimize crossings
4. **Position** nodes to satisfy all constraints

### The Edge Routing Problem

The layout engine prioritizes clean edge routing. When you have:

```d2
cluster1: {
  a
}

cluster2: {
  b
}

cluster1.a -> cluster2.b
```

The engine may stack clusters vertically to create a straight edge:

```
[cluster1]
    ↓
[cluster2]
```

Even if you wanted them side-by-side!

### Solution: Grid Override

Use `grid-columns` to force horizontal arrangement:

```d2
grid-columns: 2

cluster1: {
  a
}

cluster2: {
  b
}

cluster1.a -> cluster2.b
```

Now clusters are side-by-side, and the edge routes between them.

## Spacing

### Grid Gap

Control spacing in grid layouts:

```d2
grid-columns: 2
grid-gap: 50

a
b
c
d
```

### Vertical and Horizontal Gap

Fine-tune spacing separately:

```d2
grid-columns: 2
vertical-gap: 20
horizontal-gap: 50

a
b
c
d
```

## Dimensions

### Fixed Dimensions

Set explicit dimensions:

```d2
my_box: {
  width: 200
  height: 100
}
```

### In Grids

Grid cells can have uniform dimensions:

```d2
grid-columns: 3

a: { width: 100; height: 50 }
b: { width: 100; height: 50 }
c: { width: 100; height: 50 }
```

## Best Practices

1. **Start with direction** - Set `direction: right` or `direction: down` at the root level based on your diagram's flow

2. **Use grid for top-level containers** - When you have multiple top-level containers, use `grid-columns` to control their arrangement

3. **Keep nesting shallow** - Deep nesting can confuse the layout engine

4. **Be consistent** - Use the same direction for similar containers

5. **Test early** - Render your diagram frequently to catch layout issues early
