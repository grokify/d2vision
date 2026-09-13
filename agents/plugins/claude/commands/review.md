---
name: review
description: Review a D2 diagram for visual quality issues including layout, whitespace, and legend clarity
arguments:
  - name: path
    type: string
    required: true
    description: Path to D2 source file or rendered SVG
  - name: format
    type: string
    required: false
    default: text
    description: Output format (text, json, markdown)
  - name: fix
    type: boolean
    required: false
    default: false
    description: Automatically apply suggested fixes
dependencies:
  - d2vision
  - d2
process:
  - Parse diagram structure from SVG or D2 source
  - Analyze layout dimensions and aspect ratio
  - Calculate whitespace ratio
  - Extract and compare legend colors
  - Check element-color mapping clarity
  - Generate recommendations
---

# Diagram Review Command

Review a D2 diagram for visual quality issues and get actionable recommendations.

## Usage

```bash
# Review SVG
/review attack_chain.svg

# Review D2 source
/review attack_chain.d2

# Get JSON output
/review attack_chain.svg --format json

# Auto-fix issues
/review attack_chain.d2 --fix
```

## Review Dimensions

### 1. Layout Analysis

Checks performed:
- Aspect ratio matches declared direction
- Elements properly utilize available space
- Grid layout used where appropriate
- Container nesting is reasonable (< 4 levels)

### 2. Whitespace Analysis

Metrics calculated:
- Total diagram area
- Content area (sum of node bounds)
- Whitespace ratio
- Margin distribution (top, bottom, left, right)

### 3. Legend Review

Checks performed:
- Color uniqueness across categories
- No semantic conflicts (red used for unrelated concepts)
- Element mapping documented (what each color means)
- Legend positioning doesn't displace main content

### 4. Visual Hierarchy

Checks performed:
- Appropriate shapes for element types
- Consistent styling within categories
- Clear flow direction
- Edge overlap minimized

## Output

### Text Format (default)

```
Diagram Review: attack_chain.d2
================================

Overall: NEEDS_WORK

Layout: WARNING
  - Aspect ratio 0.95:1 for direction: right (expected > 1.5:1)
  - Legend taking 35% of diagram width

Whitespace: WARNING
  - Whitespace ratio: 45% (threshold: 40%)
  - Large left margin (1200px) due to legend placement

Legend: ERROR
  - Color conflict: #ffcdd2 used in STRIDE.S and Assets.crown
  - Missing element mapping: unclear what colors apply to

Recommendations:
  1. Add 'near: bottom-center' to legend (HIGH)
  2. Add 'grid-columns: 3' to legend container (HIGH)
  3. Change Assets colors to gold family (MEDIUM)
  4. Add descriptive labels like 'ATT&CK (arrow color)' (LOW)
```

### JSON Format

```json
{
  "file": "attack_chain.d2",
  "overall": "needs_work",
  "dimensions": {
    "layout": {"status": "warning", "issues": [...]},
    "whitespace": {"status": "warning", "ratio": 0.45},
    "legend": {"status": "error", "conflicts": [...]},
    "hierarchy": {"status": "pass"}
  },
  "metrics": {
    "width": 3380,
    "height": 3555,
    "aspect_ratio": 0.95,
    "whitespace_ratio": 0.45,
    "node_count": 35,
    "edge_count": 12
  },
  "recommendations": [
    {
      "priority": "high",
      "category": "layout",
      "action": "Add 'near: bottom-center' to legend",
      "impact": "Reduce height by ~40%"
    }
  ]
}
```

### Markdown Format

```markdown
# Diagram Review: attack_chain.d2

## Summary

| Dimension | Status |
|-----------|--------|
| Layout | WARNING |
| Whitespace | WARNING |
| Legend | ERROR |
| Hierarchy | PASS |

## Issues

### Layout
- Aspect ratio 0.95:1 for `direction: right` (expected > 1.5:1)

### Legend
- **Color conflict**: `#ffcdd2` used in both STRIDE.S and Assets.crown

## Recommendations

1. **[HIGH]** Add `near: bottom-center` to legend
2. **[HIGH]** Add `grid-columns: 3` to legend container
```

## Integration

The review command integrates with d2vision tools:

```bash
# Full workflow
d2vision lint diagram.d2           # Check for structural issues
d2vision analyze diagram.svg       # Get layout analysis
/review diagram.d2                 # Get comprehensive review
/review diagram.d2 --fix           # Apply fixes automatically
d2 diagram.d2 diagram_fixed.svg    # Re-render
d2vision diff diagram.svg diagram_fixed.svg  # Compare
```
