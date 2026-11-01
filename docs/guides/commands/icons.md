# icons

Browse and search D2's icon library hosted at [icons.terrastruct.com](https://icons.terrastruct.com/).

## Overview

D2 supports icons via URLs. Terrastruct provides a free icon library with 185+ SVG icons covering:

- Cloud providers (AWS, Azure, GCP)
- Development tools and languages
- Infrastructure components
- Common UI elements

All icons are **SVG (vector)** format, so they scale perfectly at any size.

## Usage

```bash
d2vision icons <subcommand> [flags]
```

## Subcommands

### list

List available icons and categories.

```bash
# Show category summary
d2vision icons list

# List icons in a category
d2vision icons list --category aws

# JSON output
d2vision icons list --category dev --format json
```

### search

Search for icons by name, category, or keyword.

```bash
# Search by name
d2vision icons search kubernetes

# Search by keyword
d2vision icons search serverless

# Limit results
d2vision icons search database --limit 5

# JSON output for scripting
d2vision icons search lambda --format json
```

## Categories

| Category | Count | Description |
|----------|-------|-------------|
| `essentials` | 24 | Common UI icons (user, database, cloud, lock, settings) |
| `dev` | 36 | Development tools & languages (docker, kubernetes, go, python) |
| `infra` | 10 | Infrastructure (firewall, router, load-balancer, vpn) |
| `tech` | 10 | Hardware & devices (laptop, server, mobile, cpu) |
| `social` | 8 | Social media & communication (twitter, github, slack) |
| `aws` | 39 | Amazon Web Services (EC2, S3, Lambda, RDS, DynamoDB) |
| `azure` | 28 | Microsoft Azure (VMs, Functions, Cosmos DB, AKS) |
| `gcp` | 30 | Google Cloud Platform (Compute, BigQuery, GKE, Cloud Run) |

## Examples

### List Category Summary

```bash
$ d2vision icons list

D2 Icon Library Categories
==========================

  essentials    24 icons  - Common UI icons (user, database, cloud, lock)
  dev           36 icons  - Development tools & languages (docker, kubernetes, go, python)
  infra         10 icons  - Infrastructure (firewall, router, load-balancer, vpn)
  tech          10 icons  - Hardware & devices (laptop, server, mobile, cpu)
  social         8 icons  - Social media & communication (twitter, github, slack)
  aws           39 icons  - Amazon Web Services (EC2, S3, Lambda, RDS)
  azure         28 icons  - Microsoft Azure (VMs, Functions, Cosmos DB)
  gcp           30 icons  - Google Cloud Platform (Compute, BigQuery, GKE)

Total: 185 icons
```

### Search for Icons

```bash
$ d2vision icons search kubernetes

Search results for 'kubernetes'
===============================

## Compute

  AKS                       https://icons.terrastruct.com/azure/Container%20Service%20Color/Kubernetes%20Services.svg
  EKS                       https://icons.terrastruct.com/aws/Compute/Amazon-EKS.svg
  GKE                       https://icons.terrastruct.com/gcp/Products and services/Compute/Google%20Kubernetes%20Engine.svg

  kubernetes                https://icons.terrastruct.com/dev/kubernetes.svg

Total: 4 icons

Usage in D2:
  node {
    icon: <url>
  }
```

### List AWS Icons

```bash
$ d2vision icons list --category aws

Icons in category 'aws'
=======================

## Analytics

  Kinesis                   https://icons.terrastruct.com/aws/Analytics/Amazon-Kinesis.svg
  Athena                    https://icons.terrastruct.com/aws/Analytics/Amazon-Athena.svg
  ...

## Compute

  EC2                       https://icons.terrastruct.com/aws/Compute/Amazon-EC2.svg
  Lambda                    https://icons.terrastruct.com/aws/Compute/AWS-Lambda.svg
  ...
```

## Using Icons in D2

Once you find an icon URL, use it in your D2 diagram:

```d2
# Basic icon usage
server {
  icon: https://icons.terrastruct.com/essentials/112-server.svg
}

# AWS architecture
web: Web App {
  icon: https://icons.terrastruct.com/aws/Compute/Amazon-EC2.svg
}

db: Database {
  icon: https://icons.terrastruct.com/aws/Database/Amazon-RDS.svg
}

cache: Cache {
  icon: https://icons.terrastruct.com/aws/Database/Amazon-ElastiCache.svg
}

web -> db
web -> cache
```

### Standalone Icons

Use `shape: image` for icons without a box:

```d2
aws_logo: {
  shape: image
  icon: https://icons.terrastruct.com/aws/_Group%20Icons/AWS-Cloud-alt_light-bg.svg
}
```

## Scripting with JSON Output

Use JSON output for scripting and automation:

```bash
# Get all AWS icons as JSON
d2vision icons list --category aws --format json > aws-icons.json

# Search and extract URLs with jq
d2vision icons search database --format json | jq -r '.[].url'
```

## Tips

1. **Use search by keyword**: Icons have keywords like "serverless", "container", "nosql"
2. **Check all cloud providers**: Searching "kubernetes" finds EKS, AKS, and GKE
3. **Bookmark frequently used icons**: Save URLs for icons you use often
4. **Use JSON output**: Integrate with scripts for batch diagram generation
