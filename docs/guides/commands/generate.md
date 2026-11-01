# generate

Create D2 code from structured TOON, JSON, or YAML specifications.

## Usage

```bash
d2vision generate <file> [flags]
d2vision generate - [flags]  # Read from stdin
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | auto-detect | Input format: toon, json, yaml |

## Input Schema

The `DiagramSpec` structure defines the diagram:

```yaml
# Layout settings (optional)
direction: right          # right, down, left, up
gridColumns: 2            # Force grid layout
gridRows: 2

# Top-level nodes
nodes:
  - id: external
    label: External Service
    shape: cloud

# Containers (clusters/boundaries)
containers:
  - id: cluster1
    label: Cluster 1
    direction: down
    style:
      fill: "#f0f0f0"
    nodes:
      - id: service1
        label: Service 1
      - id: db1
        label: Database
        shape: cylinder
    edges:
      - from: service1
        to: db1

# Cross-container edges
edges:
  - from: cluster1.db1
    to: external
    label: syncs
```

## Examples

### Generate from TOON

```bash
d2vision generate spec.toon > diagram.d2
```

### Generate from JSON

```bash
d2vision generate spec.json > diagram.d2
```

### Pipe from stdin

```bash
cat spec.toon | d2vision generate - > diagram.d2
```

### Full Pipeline

```bash
# Generate and render in one command
d2vision generate spec.toon | d2 - output.svg
```

### From Template

```bash
# Get template, modify, generate
d2vision template network-boundary > spec.toon
# ... edit spec.toon ...
d2vision generate spec.toon | d2 - output.svg
```

## Supported Features

### Shapes

- `rectangle` (default)
- `square`
- `circle`
- `oval`
- `diamond`
- `hexagon`
- `cylinder`
- `queue`
- `page`
- `document`
- `person`
- `cloud`

### Styles

```yaml
style:
  fill: "#ff0000"
  stroke: "#000000"
  strokeWidth: 2
  borderRadius: 8
  fontSize: 14
  opacity: 0.8
```

### Special Diagrams

Sequence diagrams and SQL tables are also supported:

```yaml
sequences:
  - id: auth_flow
    actors:
      - id: user
        shape: person
    steps:
      - from: user
        to: server
        label: login

tables:
  - id: users
    columns:
      - name: id
        type: uuid
        constraints: [PK]
```
