# Cloud Deployment Architecture

This example shows a cloud deployment architecture with multiple layers.

## Pattern Overview

```
Users → Edge → Compute → Data/Storage
              ↓
         Observability
```

## Key D2 Techniques

### Grid Layout for Top Level

```d2
grid-columns: 3
```

Arranges the top-level containers in a 3-column grid for balanced layout.

### Layer Containers

Each architectural layer is a container:

```d2
edge: Edge Layer {
  direction: down
  cdn: CDN { shape: cloud }
  waf: WAF
  lb: Load Balancer
  cdn -> waf -> lb
}
```

### Cloud-Specific Shapes

- `shape: cloud` for CDN and cloud services
- `shape: cylinder` for databases and storage
- `shape: person` for user types

### Mixed Directions

Different containers can have different internal directions:

```d2
# Vertical flow for most layers
edge: { direction: down }
compute: { direction: down }

# Horizontal for data layer (primary → replica)
data: { direction: right }
```

## Architecture Layers

1. **Users**: Web and mobile clients
2. **Edge**: CDN, WAF, load balancing
3. **Compute**: API servers, workers, schedulers
4. **Data**: Primary DB, replicas, cache
5. **Storage**: Object storage, logs
6. **Observability**: Metrics, traces, alerts

## Common Variations

### Add VPC/Network Boundaries

```d2
vpc: VPC {
  public_subnet: Public {
    lb
  }
  private_subnet: Private {
    compute
    data
  }
}
```

### Add Auto-Scaling

```d2
compute {
  asg: Auto Scaling Group {
    api_1: API Server
    api_2: API Server
    api_n: API Server
  }
}
```

### Add CI/CD

```d2
cicd: CI/CD Pipeline {
  build: Build
  test: Test
  deploy: Deploy
  build -> test -> deploy
}
deploy -> compute.api: deploys to
```

### Add Backup/DR

```d2
dr: Disaster Recovery {
  backup_region: Backup Region {
    standby_db: Standby DB { shape: cylinder }
  }
}
data.primary -> dr.standby_db: async replication
```

## Generate

```bash
# Generate D2 code
d2vision template deployment --d2

# Generate and render
d2vision template deployment --d2 | d2 - deployment.svg
```
