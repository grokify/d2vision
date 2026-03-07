# API Reference

d2vision can be used as a Go library in your own programs.

## Installation

```bash
go get github.com/grokify/d2vision
```

## Packages

| Package | Description |
|---------|-------------|
| `d2vision` | Core types and SVG parsing |
| `d2vision/format` | TOON/JSON/YAML serialization |
| `d2vision/generate` | D2 code generation |
| `d2vision/render` | D2 rendering (SVG output) |
| `d2vision/convert` | Mermaid/PlantUML conversion |

## Parsing SVGs

```go
package main

import (
    "fmt"
    "log"

    "github.com/grokify/d2vision"
)

func main() {
    // Parse from file
    diagram, err := d2vision.ParseFile("diagram.svg")
    if err != nil {
        log.Fatal(err)
    }

    // Access diagram data
    fmt.Printf("Version: %s\n", diagram.Version)
    fmt.Printf("Nodes: %d\n", len(diagram.Nodes))
    fmt.Printf("Edges: %d\n", len(diagram.Edges))

    // Iterate nodes
    for _, node := range diagram.Nodes {
        fmt.Printf("Node: %s (%s)\n", node.ID, node.Shape)
    }
}
```

## Diagram Type

```go
type Diagram struct {
    Version string   `json:"version,omitempty"`
    Title   string   `json:"title,omitempty"`
    ViewBox Bounds   `json:"viewBox"`
    Nodes   []Node   `json:"nodes"`
    Edges   []Edge   `json:"edges"`
}
```

### Node Type

```go
type Node struct {
    ID       string    `json:"id"`
    Label    string    `json:"label,omitempty"`
    Shape    ShapeType `json:"shape"`
    Bounds   Bounds    `json:"bounds"`
    Parent   string    `json:"parent,omitempty"`
    Children []string  `json:"children,omitempty"`
    Style    NodeStyle `json:"style,omitempty"`
}
```

### Edge Type

```go
type Edge struct {
    ID          string    `json:"id"`
    Source      string    `json:"source"`
    Target      string    `json:"target"`
    Label       string    `json:"label,omitempty"`
    SourceArrow ArrowType `json:"sourceArrow,omitempty"`
    TargetArrow ArrowType `json:"targetArrow"`
    Path        []Point   `json:"path,omitempty"`
}
```

## Serialization

```go
package main

import (
    "fmt"

    "github.com/grokify/d2vision"
    "github.com/grokify/d2vision/format"
)

func main() {
    diagram, _ := d2vision.ParseFile("diagram.svg")

    // TOON format (default, token-efficient)
    toon, _ := format.Marshal(diagram, format.TOON)
    fmt.Println(string(toon))

    // JSON format
    json, _ := format.Marshal(diagram, format.JSON)
    fmt.Println(string(json))

    // YAML format
    yaml, _ := format.Marshal(diagram, format.YAML)
    fmt.Println(string(yaml))
}
```

### Format Constants

```go
const (
    TOON        Format = "toon"
    JSON        Format = "json"
    JSONCompact Format = "json-compact"
    YAML        Format = "yaml"
)
```

## Generating D2 Code

```go
package main

import (
    "fmt"

    "github.com/grokify/d2vision/generate"
)

func main() {
    spec := &generate.DiagramSpec{
        GridColumns: 2,
        Containers: []generate.ContainerSpec{
            {
                ID:        "cluster1",
                Label:     "Cluster 1",
                Direction: "down",
                Nodes: []generate.NodeSpec{
                    {ID: "service1", Label: "Service 1"},
                    {ID: "db1", Label: "Database", Shape: "cylinder"},
                },
                Edges: []generate.EdgeSpec{
                    {From: "service1", To: "db1"},
                },
            },
            {
                ID:        "cluster2",
                Label:     "Cluster 2",
                Direction: "down",
                Nodes: []generate.NodeSpec{
                    {ID: "service2", Label: "Service 2"},
                    {ID: "db2", Label: "Database", Shape: "cylinder"},
                },
                Edges: []generate.EdgeSpec{
                    {From: "service2", To: "db2"},
                },
            },
        },
        Edges: []generate.EdgeSpec{
            {From: "cluster1.db1", To: "cluster2.db2", Label: "sync"},
        },
    }

    gen := generate.NewGenerator()
    d2Code := gen.Generate(spec)
    fmt.Println(d2Code)
}
```

## DiagramSpec Type

```go
type DiagramSpec struct {
    Direction   string          `json:"direction,omitempty"`
    GridColumns int             `json:"gridColumns,omitempty"`
    GridRows    int             `json:"gridRows,omitempty"`
    Nodes       []NodeSpec      `json:"nodes,omitempty"`
    Containers  []ContainerSpec `json:"containers,omitempty"`
    Edges       []EdgeSpec      `json:"edges,omitempty"`
    Sequences   []SequenceSpec  `json:"sequences,omitempty"`
    Tables      []TableSpec     `json:"tables,omitempty"`
}
```

### ContainerSpec

```go
type ContainerSpec struct {
    ID          string          `json:"id"`
    Label       string          `json:"label,omitempty"`
    Direction   string          `json:"direction,omitempty"`
    GridColumns int             `json:"gridColumns,omitempty"`
    GridRows    int             `json:"gridRows,omitempty"`
    Style       StyleSpec       `json:"style,omitempty"`
    Nodes       []NodeSpec      `json:"nodes,omitempty"`
    Containers  []ContainerSpec `json:"containers,omitempty"`
    Edges       []EdgeSpec      `json:"edges,omitempty"`
}
```

### NodeSpec

```go
type NodeSpec struct {
    ID    string    `json:"id"`
    Label string    `json:"label,omitempty"`
    Shape string    `json:"shape,omitempty"`
    Icon  string    `json:"icon,omitempty"`
    Style StyleSpec `json:"style,omitempty"`
}
```

### EdgeSpec

```go
type EdgeSpec struct {
    From        string    `json:"from"`
    To          string    `json:"to"`
    Label       string    `json:"label,omitempty"`
    SourceArrow string    `json:"sourceArrow,omitempty"`
    TargetArrow string    `json:"targetArrow,omitempty"`
    Style       StyleSpec `json:"style,omitempty"`
}
```

### StyleSpec

```go
type StyleSpec struct {
    Fill         string   `json:"fill,omitempty"`
    Stroke       string   `json:"stroke,omitempty"`
    StrokeWidth  *int     `json:"strokeWidth,omitempty"`
    BorderRadius *int     `json:"borderRadius,omitempty"`
    FontSize     *int     `json:"fontSize,omitempty"`
    Opacity      *float64 `json:"opacity,omitempty"`
}

// Helper functions
func IntPtr(i int) *int
func Float64Ptr(f float64) *float64
```

## Rendering D2 Code

The `render` package provides built-in D2 rendering without requiring the D2 CLI:

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/grokify/d2vision/render"
)

func main() {
    d2Code := `
direction: right
client: Client { shape: person }
server: Server {
    api: API
    db: Database { shape: cylinder }
    api -> db
}
client -> server.api
`

    // Create renderer
    r, err := render.New()
    if err != nil {
        log.Fatal(err)
    }

    // Render to SVG
    svg, err := r.RenderSVG(context.Background(), d2Code, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Write to file
    if err := os.WriteFile("diagram.svg", svg, 0644); err != nil {
        log.Fatal(err)
    }
}
```

### Quick Rendering

For simple use cases:

```go
// One-liner to render D2 to SVG
svg, err := render.Quick("a -> b -> c")

// Validate D2 code without rendering
err := render.Validate(ctx, d2Code)
```

### Render Options

```go
opts := &render.Options{
    ThemeID: 1,           // D2 theme (0 = default)
    Pad:     100,         // Padding in pixels
    Sketch:  true,        // Hand-drawn style
    Scale:   2.0,         // Output scale
}

svg, err := r.RenderSVG(ctx, d2Code, opts)
```

## Full Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/grokify/d2vision/generate"
    "github.com/grokify/d2vision/render"
)

func main() {
    // Create a diagram spec
    spec := &generate.DiagramSpec{
        Direction: "right",
        Nodes: []generate.NodeSpec{
            {ID: "client", Label: "Client", Shape: "person"},
        },
        Containers: []generate.ContainerSpec{
            {
                ID:        "server",
                Label:     "Server",
                Direction: "down",
                Nodes: []generate.NodeSpec{
                    {ID: "api", Label: "API"},
                    {ID: "db", Label: "Database", Shape: "cylinder"},
                },
                Edges: []generate.EdgeSpec{
                    {From: "api", To: "db"},
                },
            },
        },
        Edges: []generate.EdgeSpec{
            {From: "client", To: "server.api"},
        },
    }

    // Generate D2 code
    gen := generate.NewGenerator()
    d2Code := gen.Generate(spec)

    // Render to SVG using built-in renderer
    r, err := render.New()
    if err != nil {
        log.Fatal(err)
    }

    svg, err := r.RenderSVG(context.Background(), d2Code, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Write to file
    if err := os.WriteFile("diagram.svg", svg, 0644); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Generated diagram.svg")
}
```
