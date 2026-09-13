# compact

Render a D2 diagram with long node labels replaced by short numeric keys, and
emit a legend mapping each key back to its original label. Useful for dense
diagrams where long labels crowd the layout.

## Usage

```bash
d2vision compact <input.d2> -o <output.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | *(required)* | Output SVG file |
| `--legend` | *(stdout)* | Output JSON file for the legend mapping |
| `--font-size` | `0` | Global font-size floor in px (0 = unchanged) |

## How It Works

Leaf-node labels that are long (or contain spaces) are replaced with short keys
(`1`, `2`, …) before rendering, and a `key → original label` legend is written
(to `--legend` or stdout). The rendered SVG lays out more compactly while the
legend preserves the full labels. The optional `--font-size` floor keeps small
text legible.

## Examples

```bash
# Compact and render, legend to stdout
d2vision compact diagram.d2 -o diagram.svg

# Write the legend to a file and enforce a minimum font size
d2vision compact diagram.d2 -o diagram.svg --legend legend.json --font-size 14
```
