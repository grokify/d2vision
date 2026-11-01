# Getting Started

## Installation

### Go Install

```bash
go install github.com/grokify/d2vision/cmd/d2vision@latest
```

### From Source

```bash
git clone https://github.com/grokify/d2vision.git
cd d2vision
go install ./cmd/d2vision
```

## Prerequisites

- **Go 1.21+** for installation

d2vision includes the D2 rendering engine as a library, so no separate D2 installation is required.

The `card --png`, `card --gif` and `svg2gif` outputs are captured with headless **Chrome or Chromium**, which must be installed. Everything else, including `card --svg`, needs no browser.

## Basic Usage

### Generate a Diagram

```bash
# List available templates
d2vision template list

# Generate D2 code from a template
d2vision template microservices --d2

# Pipe to D2 for rendering
d2vision template microservices --d2 | d2 - output.svg
```

### Parse an Existing SVG

```bash
# Parse to TOON format (default)
d2vision parse diagram.svg

# Parse to JSON
d2vision parse diagram.svg --format json

# Get human-readable description
d2vision parse diagram.svg --format text
```

### Lint D2 Files

```bash
# Check for common issues
d2vision lint diagram.d2

# JSON output for CI
d2vision lint diagram.d2 --format json
```

### Compare Diagrams

```bash
# Show differences between two SVGs
d2vision diff old.svg new.svg
```

### Learn from Existing Diagrams

```bash
# Reverse engineer D2 code from an SVG
d2vision learn diagram.svg --d2

# Get as TOON spec for modification
d2vision learn diagram.svg > spec.toon
```

### Watch and Auto-Render

```bash
# Watch D2 file and re-render on changes
d2vision watch diagram.d2

# Watch with linting
d2vision watch diagram.d2 --lint

# Watch with custom output and post-render hook
d2vision watch diagram.d2 -o output.svg --on-success="open %s"
```

### Analyze Diagram Layout

```bash
# Get layout insights and generation hints
d2vision analyze diagram.svg

# Get comprehensive recreation guide
d2vision parse diagram.svg --for-generation
```

### Render an Infographic Card

Cards are datasheet-style posters described in JSON: stacked panels, each with a short text block beside a simple diagram. Edges can carry animated dots that show what is in flight, and numbered badges show the order.

```bash
# Animated SVG: open it in a browser to watch the dots move
d2vision card render examples/card/request-flow.card.json --svg card.svg

# PNG at 2x, captured 1.5 s into the animation (needs Chrome)
d2vision card render examples/card/request-flow.card.json --png card.png --scale 2 --at 1.5

# Looping GIF of one full loop (needs Chrome)
d2vision card render examples/card/request-flow.card.json --gif card.gif
```

A panel can also be derived from a [PIDL](https://github.com/grokify/pidl) protocol file, so the diagram cannot drift from the protocol. See [`examples/card/`](https://github.com/grokify/d2vision/tree/main/examples/card) and the [card command guide](commands/card.md).

### Convert an Animated SVG to a GIF

```bash
# Any animated SVG: SMIL, CSS animations, or D2's `style.animated` edges
d2vision svg2gif diagram.svg -o diagram.gif
```

## Output Formats

| Format | Flag | Description |
|--------|------|-------------|
| TOON | `--format toon` | Token-efficient (default, ~40% fewer tokens) |
| JSON | `--format json` | Standard JSON |
| JSON Compact | `--format json-compact` | Minified JSON |
| YAML | `--format yaml` | YAML format |
| Text | `--format text` | Human-readable (parse only) |

## Workflow Examples

### Template to SVG

```bash
d2vision template deployment --d2 | d2 - deployment.svg
```

### Modify and Regenerate

```bash
# Get template as spec
d2vision template network-boundary > spec.toon

# Edit spec.toon...

# Generate D2 and render
d2vision generate spec.toon | d2 - output.svg
```

### Round-Trip Editing

```bash
# Learn from existing diagram
d2vision learn existing.svg > spec.toon

# Modify spec.toon...

# Regenerate
d2vision generate spec.toon | d2 - modified.svg
```
