# Data Flow / ETL Pipeline

This example shows a typical data pipeline architecture from ingestion to consumption.

## Pattern Overview

```
Sources → Ingestion → Processing → Storage → Consumption
```

## Key D2 Techniques

### Horizontal Flow

```d2
direction: right
```

Data pipelines naturally flow left-to-right, representing the transformation journey.

### Stage Containers

Each stage is a container with internal components:

```d2
ingestion: Ingestion {
  direction: down
  collector: Data Collector
  validator: Validator
  buffer: Buffer { shape: queue }
  collector -> validator -> buffer
}
```

### Appropriate Shapes

- `shape: cylinder` for databases, data lakes, warehouses
- `shape: queue` for buffers and streams
- `shape: page` for files and reports

## Pipeline Stages Explained

1. **Sources**: Where data originates (APIs, databases, files, streams)
2. **Ingestion**: Collecting, validating, and buffering raw data
3. **Processing**: Transform, enrich, and aggregate data
4. **Storage**: Tiered storage (lake → warehouse → mart)
5. **Consumption**: Analytics, reporting, and ML

## Common Variations

### Add Data Quality

```d2
processing {
  quality_check: Data Quality {
    shape: diamond
  }
  transform -> quality_check
  quality_check -> enrich: pass
  quality_check -> dead_letter: fail
}
```

### Add Schema Registry

```d2
ingestion {
  schema_registry: Schema Registry
  validator -> schema_registry
}
```

### Add Monitoring

```d2
monitoring: Pipeline Monitoring {
  airflow: Airflow
  metrics: Metrics
  alerts: Alerts
}
```

## Generate

```bash
# Generate D2 code
d2vision template data-flow --d2

# Generate and render
d2vision template data-flow --d2 | d2 - data_flow.svg
```
