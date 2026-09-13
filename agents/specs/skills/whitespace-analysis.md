---
name: whitespace-analysis
description: Analyzes diagram whitespace ratio and identifies areas of excessive empty space
triggers:
  - whitespace
  - empty space
  - layout efficiency
dependencies:
  - d2vision
---

# Whitespace Analysis Skill

Analyze D2 diagrams for excessive whitespace that indicates layout problems.

## Methodology

### 1. Calculate Whitespace Ratio

```
total_area = viewBox.width * viewBox.height
content_area = sum(node.bounds.width * node.bounds.height for each node)
whitespace_ratio = (total_area - content_area) / total_area
```

### 2. Evaluate Thresholds

| Ratio | Rating | Action |
|-------|--------|--------|
| < 30% | Good | No action needed |
| 30-40% | Acceptable | Minor optimizations possible |
| 40-50% | Warning | Consider layout changes |
| > 50% | Problem | Layout optimization required |

### 3. Identify Whitespace Regions

Detect where whitespace accumulates:

- **Top margin**: Space above topmost element
- **Bottom margin**: Space below bottommost element
- **Left margin**: Space left of leftmost element
- **Right margin**: Space right of rightmost element
- **Internal gaps**: Space between elements

### 4. Root Cause Analysis

Common causes of excessive whitespace:

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Large top margin | Legend at top pushing content down | Move legend with `near:` |
| Large left margin | Legend on left side | Use `near: bottom-center` |
| Internal gaps | Missing grid-columns | Add `grid-columns: N` |
| Uneven distribution | Deep container nesting | Flatten hierarchy |
| Aspect ratio mismatch | Wrong direction | Change `direction:` |

## Usage

```bash
# Get diagram dimensions
d2vision analyze diagram.svg

# Extract metrics
d2vision parse diagram.svg --format json | jq '.viewBox'
```

## Output

```json
{
  "whitespace_analysis": {
    "total_area": 12019900,
    "content_area": 5408955,
    "whitespace_area": 6610945,
    "whitespace_ratio": 0.55,
    "rating": "problem",
    "regions": {
      "top_margin": 500,
      "bottom_margin": 200,
      "left_margin": 1200,
      "right_margin": 300
    },
    "likely_causes": [
      "Legend container on left side taking 1200px",
      "Missing grid-columns causing vertical stacking"
    ],
    "recommendations": [
      "Add 'near: bottom-center' to legend",
      "Add 'grid-columns: 3' to legend sub-containers"
    ]
  }
}
```
