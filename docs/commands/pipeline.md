# pipeline

Generate D2 diagrams from PipelineSpec definitions.

## Usage

```bash
d2vision pipeline <spec-file> [flags]
```

## Description

The `pipeline` command reads a PipelineSpec file (JSON, TOON, or YAML) and generates D2 code for multi-stage workflow diagrams.

PipelineSpec is designed for AI/ML pipelines, data workflows, and any multi-stage process with:

- **Stages**: Sequential or parallel execution steps
- **Executors**: Program, API, deterministic code, LLM, or agent
- **Inputs/Outputs**: Data, files, configs, prompts, models
- **Parallelism**: Fan-out/fan-in for concurrent execution
- **Swimlanes**: Group stages by lane (system/team)
- **Decision Nodes**: Conditional branching with labeled edges

## Rendering Modes

| Mode | Trigger | Description |
|------|---------|-------------|
| Detailed | default | Full I/O breakdown with inputs/executor/outputs |
| Simple | `--simple` | Compact stage boxes with executor type badges |
| Swimlane | auto (when `lane` present) | Stages grouped into lane containers |
| Decision | auto (when `branches` present) | Diamond shapes with labeled branch edges |

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | auto | Input format: json, toon, yaml |
| `--output` | `-o` | stdout | Output file |
| `--svg` | | false | Output SVG directly |
| `--simple` | | false | Compact view with only stage boxes (no internal I/O) |

## Examples

```bash
# Generate D2 code from PipelineSpec
d2vision pipeline workflow.json

# Direct SVG output
d2vision pipeline workflow.json --svg -o workflow.svg

# Simple view (compact, no I/O breakdown)
d2vision pipeline workflow.json --simple

# Read from stdin
cat workflow.json | d2vision pipeline -

# Specify input format
d2vision pipeline workflow.toon --format toon
```

## PipelineSpec Structure

```json
{
  "id": "my-pipeline",
  "label": "My Workflow",
  "direction": "right",
  "stages": [
    {
      "id": "step1",
      "label": "Process Data",
      "executor": {
        "name": "processor.py",
        "type": "deterministic"
      },
      "inputs": [
        {"id": "data", "label": "Input Data", "kind": "file", "required": true}
      ],
      "outputs": [
        {"id": "result", "label": "Result", "kind": "data"}
      ]
    }
  ]
}
```

## Executor Types

| Type | Description |
|------|-------------|
| `program` | External program/binary |
| `api` | REST/gRPC API call |
| `deterministic` | Custom code (same input = same output) |
| `llm` | Language model inference |
| `agent` | Autonomous agent execution |

## Resource Kinds

| Kind | Shape | Description |
|------|-------|-------------|
| `data` | cylinder | In-memory data structure |
| `file` | document | File on disk |
| `config` | page | Configuration |
| `prompt` | page | Prompt template |
| `model` | package | ML model weights |
| `program` | hexagon | Executable/script |
| `artifact` | package | Build artifact |

## Parallel Execution

Use `parallel` for fan-out and `joinFrom` for fan-in:

```json
{
  "stages": [
    {
      "id": "split",
      "parallel": [
        {"id": "worker_a", "executor": {"name": "worker", "type": "program"}},
        {"id": "worker_b", "executor": {"name": "worker", "type": "program"}}
      ]
    },
    {
      "id": "merge",
      "joinFrom": ["worker_a", "worker_b"]
    }
  ]
}
```

## Swimlanes

Use `lane` to group stages by system or team. Swimlanes are auto-detected when any stage has a `lane` field:

```json
{
  "id": "order-process",
  "stages": [
    {
      "id": "receive",
      "label": "Receive Order",
      "lane": "Sales",
      "executor": {"name": "order-api", "type": "api"}
    },
    {
      "id": "validate",
      "label": "Validate",
      "lane": "Sales",
      "executor": {"name": "validator", "type": "deterministic"}
    },
    {
      "id": "charge",
      "label": "Charge Card",
      "lane": "Finance",
      "executor": {"name": "payment-api", "type": "api"}
    },
    {
      "id": "ship",
      "label": "Ship",
      "lane": "Warehouse",
      "executor": {"name": "shipping-api", "type": "api"}
    }
  ]
}
```

This generates lane containers with cross-lane edges:

```d2
Sales: Sales {
  receive: "Receive Order" { ... }
  validate: Validate { ... }
}
Finance: Finance {
  charge: "Charge Card" { ... }
}
Warehouse: Warehouse {
  ship: Ship { ... }
}

Sales.receive -> Sales.validate
Sales.validate -> Finance.charge
Finance.charge -> Warehouse.ship
```

## Decision Nodes

Use `branches` to create conditional branching. Decision nodes are auto-detected and rendered as diamonds:

```json
{
  "id": "approval-flow",
  "stages": [
    {
      "id": "receive",
      "label": "Receive Request",
      "executor": {"name": "api", "type": "api"}
    },
    {
      "id": "check_amount",
      "label": "Amount > $1000?",
      "executor": {"name": "policy-check", "type": "deterministic"},
      "branches": [
        {"label": "Yes", "nextStage": "manager_review"},
        {"label": "No", "nextStage": "auto_approve"}
      ]
    },
    {
      "id": "manager_review",
      "label": "Manager Review",
      "executor": {"name": "approval-agent", "type": "agent"}
    },
    {
      "id": "auto_approve",
      "label": "Auto Approve",
      "executor": {"name": "approver", "type": "deterministic"}
    }
  ]
}
```

Decision nodes generate diamond shapes with labeled edges:

```d2
check_amount: "Amount > $1000?" {
  shape: diamond
  style.fill: "#fff9c4"
}

check_amount -> manager_review: Yes
check_amount -> auto_approve: No
```

## Combining Features

Swimlanes and decision nodes can be combined. Branch targets are correctly qualified with lane prefixes:

```json
{
  "stages": [
    {
      "id": "decide",
      "label": "Approve?",
      "lane": "Sales",
      "executor": {"name": "checker", "type": "api"},
      "branches": [
        {"label": "Yes", "nextStage": "process"},
        {"label": "No", "nextStage": "reject"}
      ]
    },
    {
      "id": "process",
      "lane": "Operations",
      "executor": {"name": "processor", "type": "program"}
    },
    {
      "id": "reject",
      "lane": "Sales",
      "executor": {"name": "rejector", "type": "program"}
    }
  ]
}
```

## See Also

- [template](template.md) - Use `d2vision template pipeline` for quick starts
- [generate](generate.md) - Generate D2 from DiagramSpec
