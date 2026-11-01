# convert

Convert Mermaid or PlantUML diagrams to D2.

## Usage

```bash
d2vision convert <input-file> [flags]
```

## Description

The `convert` command parses Mermaid or PlantUML diagram source files and converts them to D2 format. It automatically detects the source format based on content or file extension.

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--from` | `-f` | auto | Source format: mermaid, plantuml |
| `--format` | | d2 | Output format: d2, spec-toon, spec-json |
| `--lint-only` | | false | Only lint, don't convert |
| `--strict` | | false | Fail on any unsupported features |
| `--output` | `-o` | stdout | Output file |

## Examples

```bash
# Convert Mermaid flowchart to D2
d2vision convert diagram.mmd > diagram.d2

# Convert PlantUML sequence diagram
d2vision convert diagram.puml > diagram.d2

# Lint before converting
d2vision convert --lint-only diagram.mmd

# Strict mode (fail on unsupported features)
d2vision convert --strict diagram.puml

# Full pipeline to SVG
d2vision convert diagram.mmd | d2 - output.svg
```

## Supported Diagram Types

### Mermaid

| Type | Support |
|------|---------|
| Flowchart (`graph`, `flowchart`) | Full |
| Sequence diagram | Full |
| Class diagram | Partial |
| Gantt, Pie, Git | Unsupported |

### PlantUML

| Type | Support |
|------|---------|
| Sequence diagram | Full |
| Component diagram | Full |
| Class diagram | Partial |
| Activity swimlanes | Unsupported |

## Feature Mapping

### Mermaid → D2

| Mermaid | D2 |
|---------|-----|
| `graph LR` | `direction: right` |
| `A[text]` | `A: text` |
| `A((circle))` | `A { shape: circle }` |
| `A{diamond}` | `A { shape: diamond }` |
| `A --> B` | `A -> B` |
| `A -.-> B` | `A -> B { style.stroke-dash: 5 }` |
| `subgraph name` | `name { ... }` |

### PlantUML → D2

| PlantUML | D2 |
|----------|-----|
| `participant A` | Actor in sequence |
| `actor A` | `A { shape: person }` |
| `A -> B: msg` | `A -> B: msg` |
| `A --> B` | `A -> B { style.stroke-dash: 5 }` |
| `[Component]` | Node |
| `package Name` | Container |

## Lint Output

When using `--lint-only`, the command reports:

- Supported features that will convert cleanly
- Unsupported features with suggestions
- Overall convertibility assessment

```bash
$ d2vision convert --lint-only diagram.mmd

Format: mermaid
Type: flowchart
Convertible: yes

Supported:
  - Nodes with shapes
  - Edges with labels
  - Subgraphs

Unsupported:
  - Line 12: click callback (not supported in D2)
```

## See Also

- [generate](generate.md) - Generate D2 from DiagramSpec
- [learn](learn.md) - Reverse engineer D2 from SVG
