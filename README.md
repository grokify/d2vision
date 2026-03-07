# d2vision

[![Go Reference](https://pkg.go.dev/badge/github.com/grokify/d2vision.svg)](https://pkg.go.dev/github.com/grokify/d2vision)
[![Go Report Card](https://goreportcard.com/badge/github.com/grokify/d2vision)](https://goreportcard.com/report/github.com/grokify/d2vision)

Tools for [D2](https://d2lang.com) diagram parsing, generation, and AI-assisted creation.

## Overview

d2vision provides a complete toolkit for working with D2 diagrams:

- 🔍 **Parse**: Extract structure from D2-generated SVGs
- ⚙️ **Generate**: Create D2 code from structured specifications
- 📋 **Template**: Quick-start patterns for common diagrams
- 🎓 **Learn**: Reverse engineer D2 code from existing SVGs
- ✅ **Lint**: Check D2 files for common layout issues
- ↔️ **Diff**: Compare two diagrams
- 👁️ **Watch**: Auto-render D2 files on changes
- 📊 **Analyze**: Analyze layout and provide generation hints
- 🎨 **Icons**: Browse and search D2's icon library
- 🔄 **Pipeline**: Generate workflow diagrams from PipelineSpec
- 🔀 **Convert**: Convert Mermaid/PlantUML diagrams to D2
- 🔃 **Rotate**: Rotate SVG by 90° increments (landscape ↔ portrait)

Default output format is **TOON** (Token-Oriented Object Notation), which uses ~40% fewer tokens than JSON - ideal for LLM consumption.

d2vision includes the D2 rendering engine as a library - no separate D2 CLI installation required.

## Installation

```bash
go install github.com/grokify/d2vision/cmd/d2vision@latest
```

## Quick Start

```bash
# Parse an SVG diagram
d2vision parse diagram.svg

# Generate D2 from a template
d2vision template network-boundary --d2 | d2 - output.svg

# List available templates
d2vision template list
```

## Commands

### Parse

Extract structure from D2-generated SVG files.

```bash
# TOON output (default, optimized for LLMs)
d2vision parse diagram.svg

# JSON output
d2vision parse diagram.svg --format json

# Natural language description
d2vision parse diagram.svg --format text

# Brief summary
d2vision parse diagram.svg --format summary

# LLM-optimized markdown
d2vision parse diagram.svg --format llm

# Generation hints for recreating the diagram
d2vision parse diagram.svg --for-generation
```

### Generate

Create D2 code from structured TOON, JSON, or YAML specifications.

```bash
# Generate from TOON spec
d2vision generate spec.toon > diagram.d2

# Generate from JSON
d2vision generate spec.json --format json > diagram.d2

# Pipe from stdin
cat spec.toon | d2vision generate - > diagram.d2

# Full pipeline: spec → D2 → SVG
d2vision generate spec.toon | d2 - output.svg
```

### Template

Generate common diagram patterns.

```bash
# List available templates
d2vision template list

# Get template as TOON spec (for modification)
d2vision template network-boundary

# Get template as D2 code (ready to render)
d2vision template network-boundary --d2

# Customize template
d2vision template network-boundary --clusters 3 --services 4 --d2

# Full pipeline: template → SVG
d2vision template network-boundary --d2 | d2 - output.svg
```

#### Available Templates

| Template | Description |
|----------|-------------|
| `network-boundary` | Side-by-side network zones with services and datastores |
| `microservices` | Service mesh with API gateway |
| `data-flow` | ETL/data pipeline |
| `sequence` | Request/response sequence diagram |
| `entity-relationship` | Database schema with SQL tables (alias: `er`) |
| `deployment` | Cloud deployment architecture |
| `pipeline` | Multi-stage process pipeline (LLM/agent workflow) |

### Pipeline

Generate workflow diagrams from PipelineSpec.

```bash
# Generate D2 code from PipelineSpec JSON
d2vision pipeline workflow.json

# Direct SVG output
d2vision pipeline workflow.json --svg -o workflow.svg

# Simple view (compact, no I/O breakdown)
d2vision pipeline workflow.json --simple --svg -o workflow.svg

# Use pipeline template
d2vision template pipeline --d2 --pipeline-type agent
```

Pipeline types: `etl`, `llm`, `agent`

Rendering modes:

- **Detailed** (default): Full I/O breakdown with inputs/executor/outputs
- **Simple** (`--simple`): Compact stage boxes with type badges
- **Swimlane** (auto): Stages grouped by `lane` field
- **Decision** (auto): Diamond shapes for stages with `branches`

### Convert

Convert Mermaid or PlantUML diagrams to D2.

```bash
# Convert Mermaid flowchart
d2vision convert diagram.mmd

# Convert PlantUML sequence diagram
d2vision convert diagram.puml

# Lint before converting
d2vision convert --lint-only diagram.mmd
```

### Rotate

Rotate SVG diagrams by 90° increments.

```bash
# Rotate landscape to portrait (for left-binding PDF)
d2vision rotate diagram.svg --angle 90 -o portrait.svg

# Chain with pipeline
d2vision pipeline spec.json --svg | d2vision rotate - --angle 90 > portrait.svg
```

### Learn

Reverse engineer D2 code from existing SVG diagrams.

```bash
# Output TOON spec (can be modified and regenerated)
d2vision learn diagram.svg

# Output D2 code directly
d2vision learn diagram.svg --d2

# JSON output for programmatic use
d2vision learn diagram.svg --format json

# Round-trip workflow
d2vision learn diagram.svg > spec.toon
# ... edit spec.toon ...
d2vision generate spec.toon | d2 - new_diagram.svg
```

### Lint

Check D2 files for common layout issues before rendering.

```bash
# Lint a D2 file
d2vision lint diagram.d2

# JSON output for CI integration
d2vision lint diagram.d2 --format json
```

Checks for:

- Cross-container edges that may cause vertical stacking
- Missing `grid-columns` for side-by-side layouts
- Inconsistent direction settings
- Deeply nested containers (performance warning)

### Diff

Compare two diagrams and show differences.

```bash
# Compare two SVG files
d2vision diff old.svg new.svg

# JSON output for programmatic use
d2vision diff old.svg new.svg --format json

# Include position/bounds comparison
d2vision diff old.svg new.svg --bounds

# Adjust position threshold (default: 5 pixels)
d2vision diff old.svg new.svg --bounds --bounds-threshold 10
```

Reports:

- Added, removed, and modified nodes
- Added, removed, and modified edges
- Label and shape changes
- Style changes
- Position and size changes (with `--bounds` flag)

### Watch

Watch D2 files and automatically re-render on changes.

```bash
# Basic watch
d2vision watch diagram.d2

# Watch with explicit output
d2vision watch diagram.d2 output.svg

# Watch with linting before render
d2vision watch diagram.d2 --lint

# Watch with custom d2 arguments
d2vision watch diagram.d2 --d2-args="--theme=200"

# Watch with post-render command
d2vision watch diagram.d2 --on-success="open %s"
```

Features:

- Automatic re-rendering on file save
- Optional linting before render
- Debouncing to prevent multiple renders
- Custom d2 arguments
- Post-render hooks

### Analyze

Analyze diagram layout and provide generation hints.

```bash
# Analyze a diagram's layout
d2vision analyze diagram.svg

# Get analysis as JSON
d2vision analyze diagram.svg --format json

# Get analysis as TOON
d2vision analyze diagram.svg --format toon
```

Provides:

- Layout type detection (side-by-side, stacked, hierarchical, flow)
- Grid layout detection (columns and rows)
- Container hierarchy analysis
- Cross-container edge detection
- Insights about what makes the layout work
- Hints for recreating the diagram in D2

### Icons

Browse and search D2's icon library (185+ SVG icons).

```bash
# List all categories
d2vision icons list

# List icons in a category
d2vision icons list --category aws

# Search for icons
d2vision icons search kubernetes

# Search with JSON output
d2vision icons search database --format json
```

Categories:

| Category | Description |
|----------|-------------|
| essentials | Common UI icons (user, database, cloud, lock) |
| dev | Development tools & languages (docker, kubernetes, go) |
| infra | Infrastructure (firewall, router, load-balancer) |
| tech | Hardware & devices (laptop, server, mobile) |
| social | Social media (twitter, github, slack) |
| aws | Amazon Web Services (EC2, S3, Lambda, RDS) |
| azure | Microsoft Azure (VMs, Functions, Cosmos DB) |
| gcp | Google Cloud Platform (Compute, BigQuery, GKE) |

Using icons in D2:

```d2
server {
  icon: https://icons.terrastruct.com/essentials/112-server.svg
}

aws_lambda {
  icon: https://icons.terrastruct.com/aws/Compute/AWS-Lambda.svg
}
```

## Output Formats

| Format | Flag | Description |
|--------|------|-------------|
| TOON | `--format toon` | Token-Oriented Object Notation (default, ~40% fewer tokens) |
| JSON | `--format json` | Standard JSON with indentation |
| JSON Compact | `--format json-compact` | Minified JSON |
| YAML | `--format yaml` | YAML format |
| Text | `--format text` | Human-readable description |
| Summary | `--format summary` | Brief one-line summary |
| LLM | `--format llm` | Markdown optimized for LLMs |

### TOON Example

```
Version: 0.7.1
ViewBox: "(0.00, 0.00, 847.00, 666.00)"
Nodes[2]:
  - ID: cluster1
    Label: Cluster 1
    Shape: rectangle
    Bounds: "(0.00, 41.00, 418.00, 424.00)"
  - ID: cluster2
    Label: Cluster 2
    Shape: rectangle
    Bounds: "(458.00, 41.00, 187.00, 424.00)"
```

### JSON Example

```json
{
  "version": "0.7.1",
  "viewBox": {"x": 0, "y": 0, "width": 847, "height": 666},
  "nodes": [
    {"id": "cluster1", "label": "Cluster 1", "shape": "rectangle", "bounds": {"x": 0, "y": 41, "width": 418, "height": 424}},
    {"id": "cluster2", "label": "Cluster 2", "shape": "rectangle", "bounds": {"x": 458, "y": 41, "width": 187, "height": 424}}
  ]
}
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/grokify/d2vision"
    "github.com/grokify/d2vision/format"
    "github.com/grokify/d2vision/generate"
    "github.com/grokify/d2vision/render"
)

func main() {
    // Parse SVG
    diagram, err := d2vision.ParseFile("diagram.svg")
    if err != nil {
        log.Fatal(err)
    }

    // Output as TOON
    output, _ := format.Marshal(diagram, format.TOON)
    fmt.Println(string(output))

    // Generate D2 code
    spec := &generate.DiagramSpec{
        GridColumns: 2,
        Containers: []generate.ContainerSpec{
            {
                ID:        "cluster1",
                Label:     "Cluster 1",
                Direction: "down",
                Nodes: []generate.NodeSpec{
                    {ID: "service1", Label: "Service 1"},
                    {ID: "db1", Label: "Database", Shape: "cylinder"},
                },
                Edges: []generate.EdgeSpec{
                    {From: "service1", To: "db1"},
                },
            },
        },
    }

    gen := generate.NewGenerator()
    d2Code := gen.Generate(spec)
    fmt.Println(d2Code)

    // Render D2 to SVG (no external d2 CLI needed)
    r, _ := render.New()
    svg, _ := r.RenderSVG(context.Background(), d2Code, nil)
    os.WriteFile("output.svg", svg, 0644)
}
```

## Diagram Spec Schema

The `DiagramSpec` structure for generating D2 code:

```yaml
direction: right          # Layout direction: right, down, left, up
gridColumns: 2            # Force grid layout with N columns
gridRows: 2               # Force grid layout with N rows

containers:
  - id: cluster1
    label: Cluster 1
    direction: down       # Layout within container
    style:
      fill: "#f0f0f0"
      strokeWidth: 2
    nodes:
      - id: service1
        label: Service 1
        shape: rectangle  # rectangle, cylinder, circle, oval, etc.
      - id: db1
        label: Database
        shape: cylinder
    edges:
      - from: service1
        to: db1
        label: connects

nodes:                    # Top-level nodes
  - id: external
    label: External Service

edges:                    # Cross-container edges
  - from: cluster1.db1
    to: external
    label: syncs to
```

## Examples

See the `examples/` directory for complete examples:

- [`network_clusters/`](examples/network_clusters/) - Side-by-side network boundaries with aligned tops
- [`microservices/`](examples/microservices/) - Microservices architecture with API gateway
- [`data_flow/`](examples/data_flow/) - ETL/data pipeline architecture
- [`sequence/`](examples/sequence/) - Authentication flow sequence diagram
- [`entity_relationship/`](examples/entity_relationship/) - Database schema with SQL tables
- [`deployment/`](examples/deployment/) - Cloud deployment architecture

## How It Works

### Parsing

D2 encodes element IDs as base64 in CSS class names within SVGs:

- `YQ==` → `a`
- `KGEgLT4gYilbMF0=` → `(a -> b)[0]`

d2vision decodes these to reconstruct the diagram structure.

### Generation

The generator converts `DiagramSpec` to D2 code, handling:

- Layout directives (direction, grid-columns, grid-rows)
- Container nesting with proper indentation
- Shape declarations
- Edge connections with labels
- Style properties

## Supported Features

- Node extraction with shape detection (rectangle, circle, oval, cylinder, diamond, hexagon)
- Edge extraction with source/target identification
- Container hierarchy (nested nodes)
- Container-scoped edges
- Labels and styling information
- Positional bounds (x, y, width, height)
- D2 version detection
- TOON/JSON/YAML serialization

## License

MIT
