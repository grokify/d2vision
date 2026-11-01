# Layout Engines

Choosing a layout engine is really choosing between two different goals:

- **Structured diagram layout** — turn an *authored* graph (an architecture, a
  flow, a sequence) into a readable, stable picture. Nodes and edges carry
  meaning; you want hierarchy, orthogonal routing, and the same output every
  time. This is what D2 does.
- **Force-directed / large-graph layout** — take a large, *unstructured* graph
  (tens of thousands of nodes) and let a physics simulation reveal clusters.
  Position is emergent, not authored, and the result varies run to run. This is
  a different tool class; D2 does not do it.

## What ships with D2

| Engine | Included with D2? | Family | Best for | Deterministic | Practical scale |
|--------|-------------------|--------|----------|---------------|-----------------|
| **Dagre** | ✅ Open, bundled | Hierarchical (layered) | Flows, trees, pipelines, most diagrams | ✅ Yes | ≲ 100s of nodes |
| **ELK** | ✅ Open, bundled | Hierarchical + orthogonal routing | Denser graphs, parallel edges, cleaner label placement | ✅ Yes | ≲ 100s–1,000s |
| **TALA** | ⚠️ Proprietary plugin (separate binary) | Purpose-built for software architecture | Architecture diagrams where aesthetics matter | ✅ Yes | ≲ 100s |
| **ForceAtlas2** | ❌ Not in D2 (Gephi / graph libs) | Force-directed | Cluster discovery in big networks | ❌ No (seeded) | 10,000s+ |
| **SFDP** | ❌ Not in D2 (Graphviz) | Multi-level force-directed | Large sparse graphs | ❌ No | 10,000s+ |
| **CoSE / CoSE-Bilkent** | ❌ Not in D2 (Cytoscape.js) | Compound force-directed | Interactive web graph exploration, nested clusters | ❌ No | 1,000s–10,000s |

!!! note "d2vision uses Dagre and ELK"
    `d2vision` lays out in-process with the two open engines. Dagre is the
    default; ELK is available via `render.CompileWithLayout(ctx, d2, "elk")` and
    is what the [`text-overlap` lint rule](../lint-rules.md#text-overlap) uses.
    TALA is a closed-source plugin the D2 CLI shells out to and is not wired into
    `d2vision`.

## The engines D2 ships

### Dagre — the default

A layered ("Sugiyama") algorithm: nodes are assigned to ranks and edges flow
between them. Fast, deterministic, and the right choice for the large majority
of authored diagrams — flowcharts, decision trees, attack chains, pipelines.

Reach for something else when you have many edges between the same pair of nodes
(their labels crowd) or a dense graph where Dagre produces long edge crossings.

### ELK — denser graphs and cleaner labels

The Eclipse Layout Kernel is also hierarchical but does orthogonal edge routing
and spaces parallel-edge labels far better than Dagre. It is slower but handles
denser graphs and multi-edge node pairs cleanly.

!!! tip "Multi-edge node pairs"
    If several edges connect the same two nodes and their labels overlap under
    Dagre, ELK usually fixes it structurally. This is exactly the case the
    `text-overlap` rule flags — the remediation is often just "lay out with ELK."

### TALA — proprietary, architecture-tuned

Terrastruct's own engine, tuned for software-architecture aesthetics. It is
closed-source and distributed as a separate plugin binary the `d2` CLI invokes
(`d2 --layout tala`). It is not part of the open Go module and is not used by
`d2vision`.

## The engines D2 does *not* ship (and when you'd want them)

These are the force-directed / large-graph engines. They answer a different
question — *"what clusters exist in this network?"* — by running a particle
simulation until the graph settles. They are **stochastic** (seeded from
randomness, so each run differs) and are meant for **exploration of large,
unstructured graphs**, not authored diagrams.

- **ForceAtlas2** (Gephi, and graph libraries): the classic network-analysis
  layout; spreads a large graph so communities become visually obvious.
- **SFDP — Scalable Force-Directed Placement** (Graphviz): a multi-level
  force-directed engine for large sparse graphs; the large-graph companion to
  Graphviz's `dot`/`neato`.
- **CoSE / CoSE-Bilkent** (Cytoscape.js): compound (nested-cluster)
  force-directed layout, common in interactive, often GPU-accelerated, web graph
  viewers.

!!! warning "Two reasons not to force these into a D2 pipeline"
    1. **Non-determinism.** Force-directed engines seed from randomness, so the
       same input yields a different layout each run — the opposite of a
       reproducible diagram build, and it would break layout-dependent checks
       like `text-overlap`.
    2. **Wrong scale and goal.** They optimize cluster separation for huge
       graphs; on a small authored diagram (say an 8-edge attack chain) they make
       it *less* readable, not more.

**If your task is genuinely large-scale cluster discovery** ("here are 30,000
assets — show me the clusters"), that is network analysis, not diagramming. Use
Graphviz `sfdp`, Gephi/ForceAtlas2, Cytoscape.js (CoSE-Bilkent), or
`igraph`/`graph-tool` directly, and treat it as a separate export → layout →
import bridge rather than a D2 layout engine.

## Choosing

```
Authored, structured diagram?
├─ Most diagrams ......................... Dagre  (default)
├─ Dense graph / parallel edges / label
│  crowding .............................. ELK
└─ Architecture, proprietary aesthetics .. TALA   (CLI plugin)

Large, unstructured graph — find clusters?
└─ Not D2. Use Graphviz sfdp / Gephi ForceAtlas2 /
   Cytoscape CoSE-Bilkent as a separate tool.
```

For `d2vision`'s deep-dive diagrams the practical rule is: **Dagre by default,
ELK when a diagram has multiple edges between the same nodes** (which is what
triggers label overlap).
