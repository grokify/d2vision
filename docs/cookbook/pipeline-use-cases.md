# Pipeline Use Cases

PipelineSpec is designed for visualizing multi-stage processes with typed executors and explicit data flow. This page discusses which use cases it handles well and which are better served by other diagram types.

## Use Case Matrix

| Use Case | PipelineSpec | Alternative | Why |
|----------|--------------|-------------|-----|
| ETL/Data pipelines | Excellent | - | Data flow is primary concern |
| AI/ML inference chains | Excellent | - | Executor types (LLM, agent) are first-class |
| API orchestration | Excellent | - | Shows service dependencies clearly |
| CI/CD pipelines | Good | - | Sequential stages with artifacts |
| Business processes | Good | - | Supports swimlanes (`lane`) and decisions (`branches`) |
| Approval workflows | Good | - | Decision nodes with conditional branching |
| Event-driven systems | Limited | Flowchart | Async events don't fit stage model |
| State machines | Poor | Flowchart | States aren't stages |
| Org charts | Poor | DiagramSpec | Hierarchy, not flow |

---

## Excellent Fit: Data Pipelines

PipelineSpec excels at ETL and data processing workflows where:

- Data flows through sequential transformations
- Each stage has clear inputs and outputs
- Executor type matters (script vs API vs database)

### Example: ETL Pipeline

```json
{
  "id": "etl-pipeline",
  "label": "Customer Data ETL",
  "direction": "right",
  "stages": [
    {
      "id": "extract",
      "label": "Extract",
      "executor": {"name": "extract.py", "type": "program"},
      "inputs": [
        {"id": "source", "label": "Source DB", "kind": "data", "required": true}
      ],
      "outputs": [
        {"id": "raw", "label": "Raw Records", "kind": "data"}
      ]
    },
    {
      "id": "transform",
      "label": "Transform",
      "executor": {"name": "transform.py", "type": "deterministic"},
      "inputs": [
        {"id": "data", "label": "Raw Data", "kind": "data"},
        {"id": "schema", "label": "Schema", "kind": "config"}
      ],
      "outputs": [
        {"id": "clean", "label": "Clean Data", "kind": "data"}
      ]
    },
    {
      "id": "load",
      "label": "Load",
      "executor": {"name": "Warehouse API", "type": "api", "endpoint": "https://warehouse.example.com"},
      "inputs": [
        {"id": "data", "label": "Clean Data", "kind": "data"}
      ],
      "outputs": [
        {"id": "result", "label": "Load Result", "kind": "artifact"}
      ]
    }
  ]
}
```

**Why it works**: Clear data lineage, typed resources (data, config, artifact), sequential flow.

---

## Excellent Fit: AI/ML Inference Chains

PipelineSpec was designed with AI workflows in mind:

- Distinguishes deterministic code from LLM inference
- Tracks prompts as first-class resources
- Shows model dependencies

### Example: RAG Pipeline

```json
{
  "id": "rag-pipeline",
  "label": "RAG Query Pipeline",
  "direction": "right",
  "stages": [
    {
      "id": "embed",
      "label": "Embed Query",
      "executor": {"name": "text-embedding-3", "type": "llm", "model": "text-embedding-3-small"},
      "inputs": [
        {"id": "query", "label": "User Query", "kind": "data", "required": true}
      ],
      "outputs": [
        {"id": "vector", "label": "Query Vector", "kind": "data"}
      ]
    },
    {
      "id": "retrieve",
      "label": "Retrieve",
      "executor": {"name": "Vector Search", "type": "api", "endpoint": "https://pinecone.io"},
      "inputs": [
        {"id": "vector", "label": "Query Vector", "kind": "data"}
      ],
      "outputs": [
        {"id": "docs", "label": "Retrieved Docs", "kind": "data"}
      ]
    },
    {
      "id": "generate",
      "label": "Generate",
      "executor": {"name": "Claude", "type": "llm", "model": "claude-sonnet-4-20250514"},
      "inputs": [
        {"id": "context", "label": "Context", "kind": "data"},
        {"id": "prompt", "label": "System Prompt", "kind": "prompt"}
      ],
      "outputs": [
        {"id": "response", "label": "Response", "kind": "data"}
      ]
    }
  ]
}
```

**Why it works**: LLM vs API distinction visible, prompt tracking, model metadata.

---

## Excellent Fit: API Orchestration

Microservice choreography with clear service boundaries:

### Example: Order Processing

```json
{
  "id": "order-flow",
  "label": "Order Processing",
  "direction": "right",
  "stages": [
    {
      "id": "validate",
      "label": "Validate Order",
      "executor": {"name": "Order Service", "type": "api", "endpoint": "/orders/validate"},
      "inputs": [
        {"id": "order", "label": "Order Request", "kind": "data", "required": true}
      ],
      "outputs": [
        {"id": "validated", "label": "Validated Order", "kind": "data"}
      ]
    },
    {
      "id": "payment",
      "label": "Process Payment",
      "executor": {"name": "Payment Service", "type": "api", "endpoint": "/payments/charge"},
      "inputs": [
        {"id": "order", "label": "Order", "kind": "data"},
        {"id": "payment", "label": "Payment Info", "kind": "data"}
      ],
      "outputs": [
        {"id": "receipt", "label": "Receipt", "kind": "artifact"}
      ]
    },
    {
      "id": "fulfill",
      "label": "Fulfill",
      "executor": {"name": "Inventory Service", "type": "api", "endpoint": "/inventory/reserve"},
      "outputs": [
        {"id": "confirmation", "label": "Confirmation", "kind": "artifact"}
      ]
    }
  ]
}
```

**Why it works**: Service boundaries clear, API endpoints documented, data contracts visible.

---

## Good Fit: CI/CD Pipelines

Build and deployment workflows with artifacts:

### Example: Build Pipeline

```json
{
  "id": "ci-pipeline",
  "label": "CI/CD Pipeline",
  "direction": "right",
  "stages": [
    {
      "id": "build",
      "label": "Build",
      "executor": {"name": "go build", "type": "program"},
      "inputs": [
        {"id": "source", "label": "Source Code", "kind": "file"}
      ],
      "outputs": [
        {"id": "binary", "label": "Binary", "kind": "artifact"}
      ]
    },
    {
      "id": "test",
      "label": "Test",
      "executor": {"name": "go test", "type": "program"},
      "inputs": [
        {"id": "binary", "label": "Binary", "kind": "artifact"}
      ],
      "outputs": [
        {"id": "report", "label": "Test Report", "kind": "artifact"}
      ]
    },
    {
      "id": "deploy",
      "label": "Deploy",
      "executor": {"name": "kubectl", "type": "program"},
      "inputs": [
        {"id": "binary", "label": "Binary", "kind": "artifact"},
        {"id": "config", "label": "K8s Config", "kind": "config"}
      ]
    }
  ]
}
```

**Why it works**: Artifacts flow between stages, tools are explicit.

---

## Good Fit: Parallel Processing

Fan-out/fan-in patterns for concurrent work:

### Example: Map-Reduce

```json
{
  "id": "map-reduce",
  "label": "Parallel Processing",
  "stages": [
    {
      "id": "split",
      "label": "Split",
      "executor": {"name": "splitter", "type": "deterministic"},
      "inputs": [
        {"id": "dataset", "label": "Full Dataset", "kind": "data"}
      ],
      "parallel": [
        {
          "id": "worker_1",
          "label": "Worker 1",
          "executor": {"name": "mapper", "type": "program"},
          "outputs": [{"id": "partial", "label": "Partial Result", "kind": "data"}]
        },
        {
          "id": "worker_2",
          "label": "Worker 2",
          "executor": {"name": "mapper", "type": "program"},
          "outputs": [{"id": "partial", "label": "Partial Result", "kind": "data"}]
        },
        {
          "id": "worker_3",
          "label": "Worker 3",
          "executor": {"name": "mapper", "type": "program"},
          "outputs": [{"id": "partial", "label": "Partial Result", "kind": "data"}]
        }
      ]
    },
    {
      "id": "reduce",
      "label": "Reduce",
      "executor": {"name": "reducer", "type": "deterministic"},
      "joinFrom": ["worker_1", "worker_2", "worker_3"],
      "outputs": [
        {"id": "result", "label": "Final Result", "kind": "data"}
      ]
    }
  ]
}
```

**Why it works**: Parallelism is explicit, fan-in via `joinFrom`.

---

## Good Fit: Business Processes with Swimlanes

PipelineSpec now supports swimlanes via the `lane` field, making it suitable for business workflows that need to show team/department responsibilities:

### Example: Order Processing with Swimlanes

```json
{
  "id": "order-process",
  "label": "Order Processing",
  "direction": "right",
  "stages": [
    {
      "id": "receive",
      "label": "Receive Order",
      "lane": "Sales",
      "executor": {"name": "order-api", "type": "api"},
      "outputs": [{"id": "order", "label": "Order", "kind": "data"}]
    },
    {
      "id": "validate",
      "label": "Validate Order",
      "lane": "Sales",
      "executor": {"name": "validator", "type": "deterministic"}
    },
    {
      "id": "charge",
      "label": "Process Payment",
      "lane": "Finance",
      "executor": {"name": "stripe-api", "type": "api"}
    },
    {
      "id": "fulfill",
      "label": "Fulfill Order",
      "lane": "Warehouse",
      "executor": {"name": "fulfillment-api", "type": "api"}
    },
    {
      "id": "notify",
      "label": "Send Confirmation",
      "lane": "Sales",
      "executor": {"name": "email-service", "type": "api"}
    }
  ]
}
```

**Why it works**: Swimlanes are auto-detected when `lane` is present. Stages are grouped into lane containers with cross-lane edges properly qualified.

---

## Good Fit: Approval Workflows with Decisions

PipelineSpec supports decision points via the `branches` field:

### Example: Expense Approval

```json
{
  "id": "expense-approval",
  "label": "Expense Approval",
  "stages": [
    {
      "id": "submit",
      "label": "Submit Expense",
      "executor": {"name": "expense-api", "type": "api"}
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
    },
    {
      "id": "reimburse",
      "label": "Process Reimbursement",
      "executor": {"name": "payroll-api", "type": "api"}
    }
  ]
}
```

**Why it works**: Decision nodes are auto-detected and rendered as diamonds with labeled branch edges.

---

## Good Fit: Combined Swimlanes and Decisions

Swimlanes and decisions can be combined for complex business workflows:

### Example: Order with Inventory Check

```json
{
  "id": "order-with-stock-check",
  "stages": [
    {
      "id": "receive",
      "label": "Receive Order",
      "lane": "Sales",
      "executor": {"name": "order-api", "type": "api"}
    },
    {
      "id": "check_stock",
      "label": "In Stock?",
      "lane": "Warehouse",
      "executor": {"name": "inventory-api", "type": "api"},
      "branches": [
        {"label": "Yes", "nextStage": "ship"},
        {"label": "No", "nextStage": "backorder"}
      ]
    },
    {
      "id": "ship",
      "label": "Ship Order",
      "lane": "Warehouse",
      "executor": {"name": "shipping-api", "type": "api"}
    },
    {
      "id": "backorder",
      "label": "Create Backorder",
      "lane": "Warehouse",
      "executor": {"name": "backorder-api", "type": "api"}
    },
    {
      "id": "notify",
      "label": "Notify Customer",
      "lane": "Sales",
      "executor": {"name": "notification-api", "type": "api"}
    }
  ]
}
```

Use `--simple` for a cleaner view without I/O details:

```bash
d2vision pipeline order.json --simple --svg -o order.svg
```

---

## When to Use Alternatives

### Use SequenceSpec for:

- Multiple actors with back-and-forth communication
- Time-ordered message flows
- Protocol documentation

```bash
d2vision generate sequence-spec.json
```

### Use DiagramSpec for:

- State machines (states aren't stages)
- Complex flowcharts with loops
- Org charts (hierarchy, not flow)

---

## Limited Fit: Event-Driven Systems

Event-driven architectures don't fit the sequential stage model:

- Events can arrive in any order
- Multiple consumers per event
- No clear "pipeline" direction

### Better Alternative: Component Diagrams

Use DiagramSpec with containers for services and edges for event flows:

```json
{
  "containers": [
    {
      "id": "events",
      "label": "Event Bus",
      "nodes": [
        {"id": "order_created", "label": "OrderCreated"},
        {"id": "payment_processed", "label": "PaymentProcessed"}
      ]
    }
  ],
  "nodes": [
    {"id": "order_svc", "label": "Order Service"},
    {"id": "payment_svc", "label": "Payment Service"},
    {"id": "notification_svc", "label": "Notification Service"}
  ],
  "edges": [
    {"from": "order_svc", "to": "events.order_created", "label": "publishes"},
    {"from": "events.order_created", "to": "payment_svc", "label": "subscribes"},
    {"from": "events.order_created", "to": "notification_svc", "label": "subscribes"}
  ]
}
```

---

## Poor Fit: State Machines

State machines model states and transitions, not processing stages:

- States persist until an event triggers a transition
- The same event can have different effects depending on current state
- Focus is on valid transitions, not data flow

### Better Alternative: DiagramSpec with State Styling

```json
{
  "nodes": [
    {"id": "draft", "label": "Draft", "style": {"fill": "#e3f2fd"}},
    {"id": "pending", "label": "Pending Review", "style": {"fill": "#fff9c4"}},
    {"id": "approved", "label": "Approved", "style": {"fill": "#c8e6c9"}},
    {"id": "rejected", "label": "Rejected", "style": {"fill": "#ffcdd2"}}
  ],
  "edges": [
    {"from": "draft", "to": "pending", "label": "submit"},
    {"from": "pending", "to": "approved", "label": "approve"},
    {"from": "pending", "to": "rejected", "label": "reject"},
    {"from": "rejected", "to": "draft", "label": "revise"}
  ]
}
```

---

## Choosing the Right Spec

| Question | If Yes | If No |
|----------|--------|-------|
| Is data transformed through stages? | PipelineSpec | DiagramSpec |
| Do executor types matter (LLM vs API vs code)? | PipelineSpec | DiagramSpec |
| Are there decision branches? | PipelineSpec (`branches`) | PipelineSpec |
| Need swimlanes? | PipelineSpec (`lane`) | PipelineSpec |
| Multiple actors communicating back-and-forth? | SequenceSpec | PipelineSpec |
| Is it a state machine with loops? | DiagramSpec | PipelineSpec |
| Need complex flowchart with cycles? | DiagramSpec | PipelineSpec |

---

## Rendering Commands

```bash
# PipelineSpec → D2 → SVG (detailed view)
d2vision pipeline spec.json --svg -o diagram.svg

# PipelineSpec → D2 → SVG (simple view, compact)
d2vision pipeline spec.json --simple --svg -o diagram.svg

# DiagramSpec → D2
d2vision generate spec.json > diagram.d2

# SequenceSpec → D2
d2vision generate sequence.json > diagram.d2

# Any spec format
d2vision generate spec.toon    # TOON format
d2vision generate spec.yaml    # YAML format
```

## See Also

- [pipeline command](../commands/pipeline.md) - PipelineSpec reference
- [generate command](../commands/generate.md) - DiagramSpec and SequenceSpec
- [Container Patterns](containers.md) - Nested structures for complex diagrams
