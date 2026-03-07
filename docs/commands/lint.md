# lint

Check D2 files for common layout issues before rendering.

## Usage

```bash
d2vision lint <file.d2> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `text` | Output format: text, toon, json |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No issues found |
| 1 | Issues found or error occurred |

## Checks Performed

### cross-container-edge (Warning)

Cross-container edges can cause the layout engine to stack containers vertically instead of side-by-side.

```d2
# This may cause cluster2 to appear below cluster1
cluster1.service -> cluster2.db
```

**Fix**: Add `grid-columns` at the root level.

### missing-grid (Info)

Multiple root-level containers without `grid-columns` may not lay out as expected.

```d2
# Without grid-columns, layout is determined by edges
cluster1: { ... }
cluster2: { ... }
```

**Fix**: Add `grid-columns: N` to control arrangement.

### mixed-directions (Info)

Inconsistent direction settings across containers may produce confusing layouts.

```d2
cluster1: { direction: down }
cluster2: { direction: right }
cluster3: { direction: up }
```

**Fix**: Use consistent directions unless intentional.

### deep-nesting (Info)

Deeply nested containers (depth > 3) may impact layout performance.

```d2
a: { b: { c: { d: { ... } } } }
```

**Fix**: Consider flattening the structure.

### duplicate-node (Warning)

A node is defined multiple times.

## Examples

### Basic Lint

```bash
d2vision lint diagram.d2
```

Output:

```
Found 2 issue(s) in diagram.d2:

  ⚠ [cross-container-edge] Line 14: Cross-container edge 'cluster1.inner1 -> cluster2.inner3' may cause vertical stacking
    → Add 'grid-columns: N' at root level to control horizontal layout
  ℹ [missing-grid] Line 1: Found 2 root-level containers without grid-columns
    → Add 'grid-columns: N' to control horizontal arrangement
```

### JSON Output for CI

```bash
d2vision lint diagram.d2 --format json
```

```json
{
  "file": "diagram.d2",
  "issues": [
    {
      "line": 14,
      "severity": "warning",
      "code": "cross-container-edge",
      "message": "Cross-container edge 'cluster1.inner1 -> cluster2.inner3' may cause vertical stacking",
      "suggestion": "Add 'grid-columns: N' at root level to control horizontal layout"
    }
  ]
}
```

### CI Integration

```bash
# Lint before rendering
d2vision lint diagram.d2 && d2 diagram.d2 output.svg

# Or in a script
if d2vision lint diagram.d2 --format json | jq -e '.issues | length == 0' > /dev/null; then
  d2 diagram.d2 output.svg
else
  echo "Lint issues found"
  exit 1
fi
```

### Fix Workflow

```bash
# 1. Lint to find issues
d2vision lint diagram.d2

# 2. Fix based on suggestions
# Add grid-columns: 2 at the top of diagram.d2

# 3. Re-lint to verify
d2vision lint diagram.d2
# ✓ diagram.d2: no issues found
```
