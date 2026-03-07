# Microservices Architecture

This example shows a typical microservices architecture with an API gateway, multiple services, and a shared data layer.

## Pattern Overview

```
Client → API Gateway → Services → Data Layer
```

## Key D2 Techniques

### Direction for Flow

```d2
direction: right
```

Sets the overall flow direction from left to right, which is natural for reading architectural diagrams.

### Nested Containers

The API Gateway is a container with internal components:

```d2
gateway: API Gateway {
  direction: down
  auth: Auth
  rate_limit: Rate Limiter
  router: Router
  auth -> rate_limit -> router
}
```

### Grid Layout for Services

Using `grid-columns` to arrange services in a 2x2 grid:

```d2
services: Services {
  grid-columns: 2
  user_svc: User Service
  order_svc: Order Service
  product_svc: Product Service
  payment_svc: Payment Service
}
```

### Special Shapes

- `shape: person` for client representation
- `shape: cylinder` for databases (PostgreSQL, Redis)
- `shape: queue` for message queues

## Common Variations

### Add Health Checks

```d2
monitoring: Monitoring {
  prometheus: Prometheus
  grafana: Grafana
}
```

### Add Service Discovery

```d2
gateway {
  service_registry: Service Registry
  router -> service_registry
}
```

### Add Circuit Breaker Pattern

```d2
gateway {
  circuit_breaker: Circuit Breaker
  router -> circuit_breaker -> services
}
```

## Generate

```bash
# Generate D2 code
d2vision template microservices --d2

# Generate and render
d2vision template microservices --d2 | d2 - microservices.svg
```
