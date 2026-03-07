# Commands Overview

d2vision provides commands for working with D2 diagrams.

## Command Summary

| Command | Description |
|---------|-------------|
| [parse](parse.md) | Extract structure from D2-generated SVGs |
| [generate](generate.md) | Create D2 code from structured specifications |
| [template](template.md) | Generate common diagram patterns |
| [pipeline](pipeline.md) | Generate workflow diagrams from PipelineSpec |
| [convert](convert.md) | Convert Mermaid/PlantUML diagrams to D2 |
| [learn](learn.md) | Reverse engineer D2 code from SVGs |
| [lint](lint.md) | Check D2 files for layout issues |
| [diff](diff.md) | Compare two diagrams |
| [watch](watch.md) | Auto-render D2 files on changes |
| [analyze](analyze.md) | Analyze layout and provide generation hints |
| [icons](icons.md) | Browse and search D2's icon library |
| [rotate](rotate.md) | Rotate SVG by 90° increments |

## Common Flags

All commands support:

- `-h, --help` - Show help for the command
- `-f, --format` - Output format (varies by command)

## Pipeline Usage

Commands are designed to work well in Unix pipelines:

```bash
# Template → D2 → SVG
d2vision template microservices --d2 | d2 - output.svg

# Learn → Modify → Generate → Render
d2vision learn old.svg > spec.toon
# ... edit spec.toon ...
d2vision generate spec.toon | d2 - new.svg

# Lint before rendering
d2vision lint diagram.d2 && d2 diagram.d2 output.svg
```
