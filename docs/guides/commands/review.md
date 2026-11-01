# review

Comprehensive diagram quality review across layout, whitespace, legend, and
hierarchy dimensions. Accepts a D2 source file or a rendered SVG.

## Usage

```bash
d2vision review <file> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `text` | Output format: text, json, toon, markdown |

## What It Reviews

| Dimension | Checks |
|-----------|--------|
| Layout | aspect ratio, direction consistency, grid usage |
| Whitespace | excessive empty space, margin balance |
| Legend | color uniqueness, element mapping clarity |
| Hierarchy | nesting depth, shape semantics |

Each dimension gets a status (pass / warning / fail), and the review reports
metrics (size, aspect ratio, whitespace %, node/edge/container counts) plus
prioritized recommendations.

## Examples

```bash
# Review a rendered SVG (metrics + dimensions)
d2vision review diagram.svg

# Review D2 source (legend + color-conflict analysis)
d2vision review diagram.d2

# Machine-readable output
d2vision review diagram.svg --format json
```

## Related

- [lint](lint.md) — fast, deterministic structural checks (corner-near,
  text-overlap, and more) suitable for CI.
- [Lint Rules](../lint-rules.md) — the rule reference.
