# diff

Compare two diagrams and show differences.

## Usage

```bash
d2vision diff <file1.svg> <file2.svg> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `text` | Output format: text, toon, json |
| `--bounds` | `false` | Include position/bounds comparison |
| `--bounds-threshold` | `5.0` | Minimum position change to report (pixels) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Files are identical |
| 1 | Differences found |

## What Gets Compared

### Nodes

- Added nodes (in file2 but not file1)
- Removed nodes (in file1 but not file2)
- Modified nodes (same ID, different properties)
  - Label changes
  - Shape changes
  - Style changes
  - Parent changes
  - Position changes (with `--bounds` flag)
  - Size changes (with `--bounds` flag)

### Edges

- Added edges
- Removed edges
- Modified edges
  - Label changes
  - Arrow type changes

## Examples

### Basic Comparison

```bash
d2vision diff old.svg new.svg
```

Output:

```
Comparing old.svg → new.svg

Nodes:
  + new_service (added)
  - old_service (removed)
  ~ database:
      label: "DB" → "PostgreSQL"
      shape: rectangle → cylinder

Edges:
  + api -> new_service (added)
  - api -> old_service (removed)

Summary: 3 node change(s), 2 edge change(s)
```

### Identical Files

```bash
d2vision diff diagram.svg diagram.svg
```

Output:

```
✓ diagram.svg and diagram.svg are identical
```

### Position/Bounds Comparison

Compare position and size changes:

```bash
d2vision diff old.svg new.svg --bounds
```

Output:

```
Comparing old.svg → new.svg

Nodes:
  ~ service:
      position: (50.0, 100.0) → (75.0, 100.0)
      size: 120.0x60.0 → 150.0x80.0

Summary: 1 node change(s)
```

Adjust the threshold to ignore small changes:

```bash
# Only report position changes > 10 pixels
d2vision diff old.svg new.svg --bounds --bounds-threshold=10
```

### JSON Output

```bash
d2vision diff old.svg new.svg --format json
```

```json
{
  "file1": "old.svg",
  "file2": "new.svg",
  "nodes": {
    "added": ["new_service"],
    "removed": ["old_service"],
    "modified": [
      {
        "id": "database",
        "changes": [
          "label: \"DB\" → \"PostgreSQL\"",
          "shape: rectangle → cylinder"
        ]
      }
    ]
  },
  "edges": {
    "added": ["api -> new_service"],
    "removed": ["api -> old_service"]
  },
  "same": false
}
```

## Use Cases

### Verify Learn Command

```bash
# Original diagram
d2 original.d2 original.svg

# Learn and regenerate
d2vision learn original.svg --d2 > learned.d2
d2 learned.d2 learned.svg

# Compare
d2vision diff original.svg learned.svg
```

### Track Diagram Changes

```bash
# Before changes
cp diagram.svg diagram_before.svg

# Make changes to diagram.d2
d2 diagram.d2 diagram.svg

# See what changed
d2vision diff diagram_before.svg diagram.svg
```

### CI/CD Validation

```bash
# Ensure diagram hasn't changed unexpectedly
d2vision diff expected.svg actual.svg --format json | jq -e '.same'
```
