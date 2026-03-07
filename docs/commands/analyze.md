# analyze

Analyze diagram layout and provide generation hints.

## Usage

```bash
d2vision analyze <file.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `text` | Output format: text, toon, json |

## What Gets Analyzed

The analyze command examines a D2-generated SVG and provides insights about:

### Layout Detection

- **Layout type**: side-by-side, stacked, hierarchical, flow, simple
- **Direction**: right, down, left, up
- **Grid layout**: Detected columns and rows

### Container Analysis

- Container count and hierarchy
- Maximum nesting depth
- Cross-container edge detection

### Generation Hints

Actionable recommendations for recreating or improving the diagram.

## Examples

### Basic Analysis

```bash
d2vision analyze diagram.svg
```

Output:

```
# Layout Analysis

## Overview
- Nodes: 8
- Edges: 4
- Layout type: side-by-side
- Direction: right
- Grid: 2 columns x 1 rows
- Containers: 3 (max depth: 2)
- Cross-container edges: 1

## Insights
- Containers are arranged horizontally (side-by-side)
- Grid layout detected: 2 columns
- 3 container(s) with max nesting depth of 2
- 1 edge(s) cross container boundaries
- Shape types: cylinder (2), rectangle (6)

## Generation Hints
- Use `grid-columns: 2` to arrange containers side-by-side
- Set `direction: right` for primary flow
- Define containers first, then add nodes inside them
- Cross-container edges may affect layout alignment
- Consider using fully-qualified IDs (e.g., container.node) for edges
- Use `shape: cylinder` for database/storage nodes
```

### JSON Output

```bash
d2vision analyze diagram.svg --format json
```

```json
{
  "layoutType": "side-by-side",
  "direction": "right",
  "gridColumns": 2,
  "gridRows": 1,
  "hasContainers": true,
  "containerCount": 3,
  "nestingDepth": 2,
  "crossContainerEdges": 1,
  "insights": [
    "Containers are arranged horizontally (side-by-side)",
    "Grid layout detected: 2 columns",
    "3 container(s) with max nesting depth of 2",
    "1 edge(s) cross container boundaries",
    "Shape types: cylinder (2), rectangle (6)"
  ],
  "generationHints": [
    "Use `grid-columns: 2` to arrange containers side-by-side",
    "Set `direction: right` for primary flow",
    "Define containers first, then add nodes inside them",
    "Cross-container edges may affect layout alignment",
    "Consider using fully-qualified IDs (e.g., container.node) for edges",
    "Use `shape: cylinder` for database/storage nodes"
  ]
}
```

### TOON Output

```bash
d2vision analyze diagram.svg --format toon
```

```
LayoutType: side-by-side
Direction: right
GridColumns: 2
GridRows: 1
HasContainers: true
ContainerCount: 3
NestingDepth: 2
CrossContainerEdges: 1
Insights[5]:
  - Containers are arranged horizontally (side-by-side)
  - Grid layout detected: 2 columns
  ...
GenerationHints[6]:
  - Use `grid-columns: 2` to arrange containers side-by-side
  ...
```

## Use Cases

### Learn from Existing Diagrams

Understand how a well-designed diagram achieves its layout:

```bash
# Analyze a diagram you want to replicate
d2vision analyze reference_diagram.svg

# Use the hints to create a similar structure
```

### Debugging Layout Issues

When your diagram doesn't look right:

```bash
# Generate and render
d2vision template network-boundary --d2 | d2 - output.svg

# Analyze the result
d2vision analyze output.svg
```

### Validate Diagram Structure

Check if a diagram follows expected patterns:

```bash
# Verify grid layout is detected
d2vision analyze diagram.svg --format json | jq '.gridColumns'

# Check for cross-container edges
d2vision analyze diagram.svg --format json | jq '.crossContainerEdges'
```

### AI-Assisted Diagram Creation

Provide context to an AI assistant:

```bash
# Get comprehensive analysis for AI consumption
d2vision analyze existing.svg --format json > analysis.json
```

## Related Commands

- [`parse`](parse.md) - Use `--for-generation` for full recreation hints with D2 skeleton
- [`learn`](learn.md) - Reverse engineer D2 code directly
- [`lint`](lint.md) - Check D2 source files for issues

## Comparison: analyze vs parse --for-generation

| Feature | `analyze` | `parse --for-generation` |
|---------|-----------|--------------------------|
| Layout analysis | ✓ | ✓ |
| Generation hints | ✓ | ✓ |
| D2 code skeleton | ✗ | ✓ |
| Container tree | ✗ | ✓ |
| Node/edge listing | ✗ | ✓ |
| Structured output | ✓ (JSON/TOON) | Text only |

Use `analyze` for quick insights or machine-readable output. Use `parse --for-generation` for comprehensive recreation guidance including a D2 code skeleton.
