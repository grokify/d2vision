# d2vision Tasks

## Overview

Extend d2vision from SVG parsing to full diagram generation support for AI assistants.

**Default format**: TOON (Token-Oriented Object Notation) for ~40% fewer tokens
**Optional formats**: JSON, JSON-compact, YAML via `--format`

## Tasks

### 1. Add TOON Format Support
**Status**: Complete
**Priority**: High (foundation for other features)

Add TOON as default output format using `github.com/toon-format/toon-go`.

- [x] Add `format` package (similar to structured-changelog)
- [x] Update existing commands to support `--format toon|json|json-compact`
- [x] Make TOON the default output format
- [x] Update tests

### 2. `d2vision generate` - Generate D2 from Structured Input
**Status**: Complete
**Priority**: High

Generate D2 code from TOON/JSON/YAML specifications.

- [x] Define input schema (DiagramSpec)
- [x] Implement TOON/JSON/YAML parsing
- [x] Implement D2 code generator
- [x] Handle layouts (grid, direction)
- [x] Handle shapes, styles, labels
- [x] Handle nested containers
- [x] Handle cross-container edges
- [x] Add tests with real-world examples

### 3. `d2vision template` - Pattern Library
**Status**: Complete
**Priority**: Medium

Provide common diagram templates that AI assistants can use as starting points.

Templates to include:

- [x] `network-boundary` - Side-by-side network zones with services and datastores
- [x] `microservices` - Service mesh with API gateway
- [x] `data-flow` - ETL/data pipeline
- [x] `deployment` - Cloud deployment architecture
- [x] `sequence` - Request/response flow (using D2 sequence diagrams)
- [x] `entity-relationship` - Database schema

### 4. `d2vision learn` - Reverse Engineer Diagrams
**Status**: Complete
**Priority**: High

Analyze an SVG and output D2 code that recreates it.

```bash
d2vision learn diagram.svg > recreated.d2
d2vision learn diagram.svg --format toon > diagram.toon
```

This helps AI assistants:

- Understand existing diagram patterns
- Learn from examples
- Modify existing diagrams

- [x] Extend parser to extract all styling
- [x] Generate D2 code from Diagram struct
- [x] Handle layout hints (direction, grid)
- [x] Preserve labels, shapes, connections
- [x] Round-trip test: D2 → SVG → learn → D2 → SVG (should be similar)

### 5. `d2vision lint` - Validate D2 Files
**Status**: Complete
**Priority**: High

Check D2 files for common layout issues before rendering.

```bash
d2vision lint diagram.d2
```

Checks:

- [x] Cross-container edges that may cause alignment issues
- [x] Missing `grid-columns` for side-by-side layouts
- [x] Inconsistent direction settings
- [x] Deeply nested containers (performance warning)
- [x] Suggest fixes for common problems

### 6. `d2vision diff` - Compare Diagrams
**Status**: Complete
**Priority**: High

Compare two diagrams and show differences.

```bash
d2vision diff diagram1.svg diagram2.svg
```

Features:

- [x] Compare node sets (added, removed, modified)
- [x] Compare edge sets
- [x] Compare positions/bounds
- [x] Compare styles
- [x] Output as TOON/JSON or human-readable diff

### 7. `d2vision describe` Enhancements
**Status**: Complete
**Priority**: Low

Enhance existing describe functionality for AI consumption.

- [x] Basic text/summary/llm formats
- [x] Add `--for-generation` flag that outputs hints for recreating
- [x] Include layout analysis (what makes this diagram work)
- [x] Add `analyze` command for dedicated layout analysis

### 8. CI/CD Setup
**Status**: Complete
**Priority**: High

GitHub Actions and automation.

- [x] go-ci.yaml - Build and test
- [x] go-lint.yaml - Linting
- [x] go-sast-codeql.yaml - Security scanning
- [x] dependabot.yaml - Dependency updates
- [x] goreleaser - Automated releases

### 9. More Examples & Documentation
**Status**: Complete
**Priority**: Medium

Add example diagrams, patterns, and documentation.

- [x] `network_clusters/` - Side-by-side clusters with layout insights
- [x] `microservices/` - Service mesh architecture
- [x] `data_flow/` - ETL pipeline
- [x] `sequence/` - Sequence diagram
- [x] `entity_relationship/` - Database schema
- [x] `deployment/` - Cloud deployment
- [x] MkDocs documentation site
- [x] Cookbook of common patterns (layout, containers, edges, styling)
- [x] Round-trip tests for learn command

### 10. Additional Features (Nice to Have)
**Status**: Partially Complete
**Priority**: Low

Future enhancements:

- [ ] SVG post-processing (adjust positions after render)
- [x] Watch mode (live reload during development)
- [x] Icon library integration
- [x] Mermaid/PlantUML conversion (import from other formats) - `d2vision convert`

### 11. PipelineSpec Multi-Mode Rendering
**Status**: Complete
**Priority**: Medium

Extended PipelineSpec to support multiple visualization modes from a single source of truth:

- [x] Simple mode (`--simple`) - Compact stage boxes without I/O breakdown
- [x] Swimlane mode (auto-detected) - Stages grouped by `lane` field
- [x] Decision mode (auto-detected) - Diamond shapes for stages with `branches`
- [x] Combined swimlane + decision support
- [x] Schema extensions (`Lane`, `Branches`, `BranchSpec`)
- [x] CLI integration (`--simple` flag)
- [x] Tests for all modes
- [x] Documentation updates
- [x] Pipeline examples (`examples/pipeline/`)

## Implementation Order

1. ~~TOON format support~~ ✓
2. ~~generate command~~ ✓
3. ~~CI/CD setup~~ ✓
4. ~~learn command~~ ✓
5. ~~lint command~~ ✓
6. ~~diff command~~ ✓
7. ~~More templates~~ ✓
8. ~~Examples and documentation~~ ✓
9. ~~goreleaser~~ ✓
10. ~~Additional features~~ ✓ (convert, watch, icons)
11. ~~PipelineSpec multi-mode rendering~~ ✓

## Dependencies

- `github.com/toon-format/toon-go` - TOON serialization
- `github.com/spf13/cobra` - CLI
- `gopkg.in/yaml.v3` - YAML parsing

## Notes

- All commands should work well in pipelines: `d2vision generate | d2 - output.svg`
- TOON format saves ~40% tokens vs JSON, important for AI context limits
- Templates should include comments explaining D2 layout tricks (like `grid-columns`)
