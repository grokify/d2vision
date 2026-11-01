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
| `-f, --format` | `toon` | Input format: toon, json, yaml. The format is not detected from the file name, so name it for JSON and YAML |

## Input Schema

The `DiagramSpec` structure defines the diagram:

```yaml
# Layout settings (optional)
direction: right          # right, down, left, up
gridColumns: 2            # Force grid layout
gridRows: 2
horizontalGap: 90         # Space between grid cells, in pixels
verticalGap: 24

# Shared styles, named once and used by class
classes:
  input:
    fill: "#fff1dc"
    stroke: "#e08a1e"

# Top-level nodes
nodes:
  - id: external
    label: External Service
    shape: cloud
    class: input

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
d2vision generate -f json spec.json > diagram.d2
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
  fontColor: "#111111"
  bold: true
```

### Labels

A label is shown as written, whatever it holds. Text with characters D2 reads as
syntax (`:`, `;`, `{`, `}`, `#`, `$`, quotes) is written in single quotes, so it
compiles and appears unchanged. A line break in a label becomes a space.

### Edge endpoints

A dot separates the containers on the way to a node. `cluster1.db1` is the node
`db1` in the container `cluster1`. Inside a container, a name is relative to that
container. A node ID that itself contains a dot cannot be named by an edge.

### Classes

`classes` names a style once, and a node, container, or edge uses it with
`class`. Classes are written in name order, so the same spec always gives the
same D2.

### Grid spacing

With `gridColumns` or `gridRows`, `horizontalGap` and `verticalGap` set the space
between cells. Leave room between columns when edges run between them.

## Mermaid output

The same spec can be written as a Mermaid flowchart from Go, with
`generate.NewGenerator().Mermaid(spec)`. Containers become subgraphs, a node is
named by its whole path (Mermaid IDs are global), and classes become `classDef`s.
Mermaid has no grid, so `gridColumns` and `gridRows` are ignored, and a spec with
sequences or tables is refused. See the [API reference](../api-reference.md).

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
