# Examples

Complete example diagrams demonstrating d2vision templates and D2 patterns.

## Available Examples

| Example | Description |
|---------|-------------|
| Network Clusters | Side-by-side network boundaries with services and datastores |
| Microservices | Service mesh with API gateway |
| Data Flow | ETL/data pipeline |
| Sequence | Authentication flow sequence diagram |
| Entity Relationship | Database schema with SQL tables |
| Deployment | Cloud deployment architecture |

## Using Examples

Each example includes:

- **D2 source code** - The `.d2` file
- **Rendered SVG** - The visual output
- **Key techniques** - What D2 features it demonstrates
- **Variations** - How to customize it

## Quick Start with Templates

Generate any example using d2vision:

```bash
# Generate D2 code
d2vision template network-boundary --d2

# Render directly
d2vision template microservices --d2 | d2 - microservices.svg

# Get as spec for modification
d2vision template data-flow > spec.toon
```

## Example Source Files

All examples are available in the repository:

- [`examples/network_clusters/`](https://github.com/grokify/d2vision/tree/main/examples/network_clusters)
- [`examples/microservices/`](https://github.com/grokify/d2vision/tree/main/examples/microservices)
- [`examples/data_flow/`](https://github.com/grokify/d2vision/tree/main/examples/data_flow)
- [`examples/sequence/`](https://github.com/grokify/d2vision/tree/main/examples/sequence)
- [`examples/entity_relationship/`](https://github.com/grokify/d2vision/tree/main/examples/entity_relationship)
- [`examples/deployment/`](https://github.com/grokify/d2vision/tree/main/examples/deployment)
