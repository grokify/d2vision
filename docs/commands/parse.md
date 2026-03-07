# parse

Extract structure from D2-generated SVG files.

## Usage

```bash
d2vision parse <file.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `toon` | Output format: toon, json, json-compact, yaml, text, summary, llm, analysis |
| `--for-generation` | `false` | Output hints for recreating the diagram |

## Output Formats

### TOON (default)

Token-efficient format, ~40% fewer tokens than JSON:

```bash
d2vision parse diagram.svg
```

```
Version: 0.7.1
ViewBox: "(0.00, 0.00, 847.00, 666.00)"
Nodes[3]:
  - ID: a
    Label: a
    Shape: rectangle
    Bounds: "(50.00, 50.00, 100.00, 60.00)"
  ...
```

### JSON

```bash
d2vision parse diagram.svg --format json
```

```json
{
  "version": "0.7.1",
  "viewBox": {"x": 0, "y": 0, "width": 847, "height": 666},
  "nodes": [
    {"id": "a", "label": "a", "shape": "rectangle", "bounds": {"x": 50, "y": 50, "width": 100, "height": 60}}
  ]
}
```

### Text

Human-readable description:

```bash
d2vision parse diagram.svg --format text
```

```
This D2 diagram contains 3 nodes and 2 connections.

Nodes:
- "a" is a rectangle
- "b" is a rectangle
- "c" is a rectangle

Connections:
- "a" connects to "b"
- "b" connects to "c"
```

### Summary

Brief one-line summary:

```bash
d2vision parse diagram.svg --format summary
```

```
D2 diagram with 3 nodes and 2 edges
```

### LLM

Markdown optimized for LLMs:

```bash
d2vision parse diagram.svg --format llm
```

### For Generation

Comprehensive output with layout analysis, insights, and a D2 code skeleton:

```bash
d2vision parse diagram.svg --for-generation
```

Output:

```markdown
# D2 Diagram Analysis for Recreation

## Overview
- 8 nodes, 4 edges
- Layout type: side-by-side
- Direction: right
- Grid: 2 columns x 1 rows

## Layout Analysis
- Containers are arranged horizontally (side-by-side)
- Grid layout detected: 2 columns
- 3 container(s) with max nesting depth of 2

## Generation Hints
- Use `grid-columns: 2` to arrange containers side-by-side
- Set `direction: right` for primary flow
- Define containers first, then add nodes inside them

## Structure

### Containers
- cluster1 (label: "Cluster 1")
  - cluster1.services
    - cluster1.services.service1a
  - cluster1.datastore1 [cylinder]

### Edges
- cluster1.services.service1a -> cluster1.datastore1

## Suggested D2 Code Skeleton

```d2
grid-columns: 2

cluster1: Cluster 1 {
  services {
    service1a: Service 1A
  }
  datastore1: DataStore 1 {
    shape: cylinder
  }
}

# Edges
cluster1.services.service1a -> cluster1.datastore1
```
```

This is useful for understanding how to recreate a diagram or for providing context to an AI assistant.

## Examples

```bash
# Parse and save as JSON
d2vision parse diagram.svg --format json > diagram.json

# Parse multiple files
for f in *.svg; do
  d2vision parse "$f" --format summary
done

# Pipe to jq for processing
d2vision parse diagram.svg --format json | jq '.nodes[].id'
```

## What Gets Extracted

- **Nodes**: ID, label, shape, position/bounds, style, parent container
- **Edges**: Source, target, label, arrow types, path
- **Containers**: Hierarchy of nested elements
- **Metadata**: D2 version, viewBox dimensions
