# learn

Reverse engineer D2 code from existing SVG diagrams.

## Usage

```bash
d2vision learn <file.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `toon` | Output format: toon, json, yaml |
| `--d2` | `false` | Output D2 code instead of spec |

## What It Does

The `learn` command analyzes a D2-generated SVG and produces either:

1. A **DiagramSpec** (TOON/JSON/YAML) that can be modified and regenerated
2. **D2 code** that recreates the diagram

This helps AI assistants:

- Understand existing diagram patterns
- Learn from examples
- Modify existing diagrams

## Examples

### Get D2 Code

```bash
d2vision learn diagram.svg --d2
```

Output:

```d2
grid-columns: 2

cluster1: Cluster 1 {
  direction: down
  service1: Service 1
  db1: Database { shape: cylinder }
  service1 -> db1
}
```

### Get TOON Spec

```bash
d2vision learn diagram.svg
```

### Round-Trip Editing

```bash
# Learn from existing diagram
d2vision learn original.svg > spec.toon

# Modify the spec (add nodes, change labels, etc.)
# ... edit spec.toon ...

# Regenerate
d2vision generate spec.toon | d2 - modified.svg
```

### Compare Original vs Learned

```bash
# Generate D2 from learned spec
d2vision learn original.svg --d2 > learned.d2

# Render the learned version
d2 learned.d2 learned.svg

# Compare
d2vision diff original.svg learned.svg
```

## What Gets Detected

### Layout

- `grid-columns` from horizontal node alignment
- `direction` from child node positions (variance analysis)

### Structure

- Container hierarchy
- Nested containers
- Node shapes and labels
- Edges (internal and cross-container)

### Shapes

Detected shapes:

- rectangle
- circle
- oval
- cylinder
- diamond
- hexagon

## Limitations

- Styling (colors, fonts) may not be fully preserved
- Complex path-based shapes may default to rectangle
- Layout engine hints (like `near`) are not detected
- Some edge routing details may be lost
