# template

Generate common diagram patterns as starting points.

## Usage

```bash
d2vision template <name> [flags]
d2vision template list
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `toon` | Output format: toon, json, yaml |
| `--d2` | `false` | Output D2 code instead of spec |
| `--clusters` | `2` | Number of clusters (network-boundary) |
| `--services` | `2` | Services per cluster (network-boundary) |

## Available Templates

| Template | Description |
|----------|-------------|
| `network-boundary` | Side-by-side network zones with services and datastores |
| `microservices` | Service mesh with API gateway |
| `data-flow` | ETL/data pipeline |
| `sequence` | Request/response sequence diagram |
| `entity-relationship` | Database schema with SQL tables (alias: `er`) |
| `deployment` | Cloud deployment architecture |

## Examples

### List Templates

```bash
d2vision template list
```

### Get as TOON Spec

```bash
d2vision template microservices > spec.toon
```

### Get as D2 Code

```bash
d2vision template microservices --d2
```

### Render Directly

```bash
d2vision template microservices --d2 | d2 - microservices.svg
```

### Customize Network Boundary

```bash
# 3 clusters with 4 services each
d2vision template network-boundary --clusters 3 --services 4 --d2
```

## Template Details

### network-boundary

Side-by-side network zones demonstrating `grid-columns` for horizontal layout.

```bash
d2vision template network-boundary --d2
```

Key techniques:

- `grid-columns: 2` at root for side-by-side clusters
- `direction: down` inside each cluster
- Invisible container (`""`) for horizontal service grouping
- `shape: cylinder` for datastores

### microservices

Service mesh architecture with API gateway, services, and data layer.

```bash
d2vision template microservices --d2
```

Key techniques:

- `direction: right` for left-to-right flow
- Nested containers for gateway components
- `grid-columns: 2` for service grid
- Multiple shapes: `person`, `cylinder`, `queue`

### data-flow

ETL/data pipeline from sources to consumption.

```bash
d2vision template data-flow --d2
```

Key techniques:

- Five-stage pipeline: Sources → Ingestion → Processing → Storage → Consumption
- Internal edges within each stage
- Cross-stage edges for data flow

### sequence

Authentication flow sequence diagram.

```bash
d2vision template sequence --d2
```

Key techniques:

- `shape: sequence_diagram`
- Actor ordering
- Message sequence
- Groups for error cases

### entity-relationship

Database schema with SQL tables.

```bash
d2vision template er --d2
```

Key techniques:

- `shape: sql_table`
- Column types and constraints
- Foreign key relationships

### deployment

Cloud deployment architecture with multiple layers.

```bash
d2vision template deployment --d2
```

Key techniques:

- `grid-columns: 3` for layer arrangement
- Mixed directions per container
- Cloud-specific shapes
