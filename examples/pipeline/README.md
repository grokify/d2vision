# Pipeline Examples

This directory contains PipelineSpec examples demonstrating different rendering modes.

## Examples

| File | Mode | Description |
|------|------|-------------|
| `etl-simple.json` | Detailed | Basic ETL pipeline with inputs/outputs |
| `order-swimlanes.json` | Swimlane | Order processing across Sales/Finance/Warehouse |
| `approval-decision.json` | Decision | Expense approval with conditional branching |
| `combined-swimlane-decision.json` | Combined | Swimlanes with decision nodes |

## Rendering Modes

### Detailed (default)

Full I/O breakdown with inputs, executor, and outputs for each stage.

```bash
d2vision pipeline etl-simple.json --svg -o etl-detailed.svg
```

### Simple

Compact view with only stage boxes and executor type badges.

```bash
d2vision pipeline etl-simple.json --simple --svg -o etl-simple.svg
```

### Swimlane (auto-detected)

Stages grouped by `lane` field into containers. Automatically enabled when any stage has a `lane` field.

```bash
d2vision pipeline order-swimlanes.json --svg -o order-swimlanes.svg
```

### Decision (auto-detected)

Stages with `branches` field rendered as diamonds with labeled edges. Automatically enabled when any stage has branches.

```bash
d2vision pipeline approval-decision.json --simple --svg -o approval-decision.svg
```

### Combined

Swimlanes and decision nodes can be used together. Branch targets are correctly qualified with lane prefixes.

```bash
d2vision pipeline combined-swimlane-decision.json --simple --svg -o combined.svg
```

## Generate All Examples

```bash
# Detailed mode
d2vision pipeline etl-simple.json --svg -o etl-detailed.svg

# Simple mode
d2vision pipeline etl-simple.json --simple --svg -o etl-simple.svg

# Swimlanes
d2vision pipeline order-swimlanes.json --simple --svg -o order-swimlanes.svg

# Decision flow
d2vision pipeline approval-decision.json --simple --svg -o approval-decision.svg

# Combined
d2vision pipeline combined-swimlane-decision.json --simple --svg -o combined.svg
```
