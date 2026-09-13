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
| `--config` | *(auto)* | Path to a lint config (YAML); defaults to the nearest `.d2vision.yaml` |
| `--list` | `false` | List every lint rule (code, severity, title) and exit |
| `--explain <code>` | | Print the full remediation guidance for a rule code and exit |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No issues found |
| 1 | Issues found or error occurred |

## Configuration

A single YAML config (`.d2vision.yaml`) declares which rules run and how, so one
invocation applies a whole policy — in the spirit of golangci-lint. It is
resolved as `--config <path>`, else the nearest `.d2vision.yaml` walking up from
the target file, else the built-in defaults.

```yaml
rules:
  text-overlap:          # opt-in: runs a full layout pass
    enabled: true
    settings:
      layout: elk        # check the engine you actually render with
  corner-near:
    severity: warning    # override the default severity
  deep-nesting:
    enabled: false       # turn a rule off
```

Discover rules and their remediation from the CLI:

```bash
d2vision lint --list                # every rule: code, severity, title
d2vision lint --explain corner-near # full remediation for one rule
```

See [Lint Rules](../lint-rules.md) for the full rule reference.

## Checks Performed

### corner-near (Error)

An element pinned to a corner via `near` (`top-left`/`top-right`/`bottom-left`/
`bottom-right`) lands diagonally from the diagram body, leaving ~50% whitespace.
A compiler-based (structural) check.

```d2
legend: Legend { near: top-left }   # flagged
```

**Fix**: use an edge-center `near` (`top-center`, `bottom-center`, `center-left`,
`center-right`) so the element stacks above/below/left/right of the diagram.

### text-overlap (Warning, opt-in)

A connection label's rendered rectangle overlaps another label or an unrelated
node. This is a **post-layout** check — it runs a full layout pass, so it is
**engine-dependent** and **opt-in** via config (`rules.text-overlap.enabled:
true`, `settings.layout: elk|dagre`).

**Fix**: switch layout engine (ELK spaces parallel-edge labels better), reduce
parallel edges between the same node pair, or shorten/relocate the label.

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
