# D2Vision Quality Analysis Plugin

This plugin provides agents and skills for analyzing D2 diagram quality, focusing on layout optimization, whitespace reduction, and legend clarity.

## Available Agents

### diagram-quality-analyst
Comprehensive diagram analysis covering layout, whitespace, legend, and visual hierarchy.

### layout-optimizer
Suggests and applies layout optimizations to reduce whitespace and improve flow.

### legend-reviewer
Reviews legend colors for conflicts and ensures clear element mapping.

## Available Skills

### whitespace-analysis
Calculate whitespace ratio and identify regions of excessive empty space.

### color-conflict-detection
Detect when multiple legend categories use the same or similar colors.

## Commands

### /review
Review a diagram for quality issues with actionable recommendations.

```bash
/review diagram.svg              # Text output
/review diagram.d2 --format json # JSON output
/review diagram.d2 --fix         # Auto-fix issues
```

## Integration with d2vision CLI

```bash
# Parse diagram structure
d2vision parse diagram.svg

# Analyze layout
d2vision analyze diagram.svg

# Lint for issues
d2vision lint diagram.d2

# Compare before/after
d2vision diff original.svg optimized.svg
```

## Quality Criteria

### Layout
- Aspect ratio matches direction (horizontal flows should be wider than tall)
- Grid layout used for side-by-side elements
- Container nesting < 4 levels

### Whitespace
- Whitespace ratio < 40%
- Legend positioned with `near:` to avoid displacement
- Margins balanced

### Legend
- Distinct color families per category
- No red used for both threats AND assets
- Labels describe what's colored (e.g., "ATT&CK (arrow color)")

### Color Families
| Category | Family | Use |
|----------|--------|-----|
| STRIDE threats | Semantic | S=red, E=green, I=blue |
| ATT&CK tactics | Blue gradient | Light→dark progression |
| Assets | Gold/Teal | Value classification |
| Malicious actors | Red | Danger indication |
