# rotate

Rotate SVG diagrams by 90° increments.

## Usage

```bash
d2vision rotate <input.svg> [flags]
```

## Description

The `rotate` command rotates an SVG diagram by 90° increments. This is useful for converting landscape diagrams to portrait orientation for PDF embedding with left-side binding.

Rotation uses counter-clockwise direction (mathematical convention).

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--angle` | `-a` | 90 | Rotation angle in degrees |
| `--output` | `-o` | stdout | Output file |

## Rotation Angles

| Angle | Direction | Use Case |
|-------|-----------|----------|
| `90` | 90° CCW | Landscape → Portrait (top goes to left/binding side) |
| `180` | 180° | Flip upside down |
| `270` | 270° CCW | Landscape → Portrait (top goes to right) |
| `-90` | 90° CW | Same as 270° |

## Examples

```bash
# Rotate landscape to portrait for left-binding PDF
d2vision rotate diagram.svg --angle 90 -o portrait.svg

# Rotate and pipe to stdout
d2vision rotate diagram.svg --angle 90

# Read from stdin
cat diagram.svg | d2vision rotate - --angle 90 > portrait.svg

# Chain with pipeline command
d2vision pipeline spec.json --svg | d2vision rotate - --angle 90 > portrait.svg

# Flip upside down
d2vision rotate diagram.svg --angle 180 -o flipped.svg
```

## Use Case: PDF Embedding

When embedding diagrams in PDFs with left-side binding:

1. Generate a landscape diagram (wider than tall)
2. Rotate 90° counter-clockwise
3. The original "top" of the diagram is now on the left (binding side)
4. Reader turns the page 90° clockwise to view naturally

```bash
# Generate landscape workflow
d2vision pipeline workflow.json --svg -o workflow-landscape.svg

# Convert to portrait for PDF
d2vision rotate workflow-landscape.svg --angle 90 -o workflow-portrait.svg
```

## See Also

- [pipeline](pipeline.md) - Generate workflow diagrams
