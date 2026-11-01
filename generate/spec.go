// Package generate provides D2 code generation from structured specifications.
package generate

// DiagramSpec defines the structure for generating D2 diagrams.
// This can be serialized as TOON, JSON, or YAML.
type DiagramSpec struct {
	// Layout settings
	Direction   string `json:"direction,omitempty" yaml:"direction,omitempty"`     // right, down, left, up
	GridColumns int    `json:"gridColumns,omitempty" yaml:"gridColumns,omitempty"` // Force grid layout
	GridRows    int    `json:"gridRows,omitempty" yaml:"gridRows,omitempty"`
	// HorizontalGap and VerticalGap are the space between grid cells, in pixels.
	HorizontalGap int `json:"horizontalGap,omitempty" yaml:"horizontalGap,omitempty"`
	VerticalGap   int `json:"verticalGap,omitempty" yaml:"verticalGap,omitempty"`

	// Classes are named styles shared by nodes, containers, and edges that name
	// them with Class. They are written as D2 classes and as Mermaid classDefs.
	Classes map[string]StyleSpec `json:"classes,omitempty" yaml:"classes,omitempty"`

	// Top-level elements
	Nodes      []NodeSpec      `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Containers []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty"`
	Edges      []EdgeSpec      `json:"edges,omitempty" yaml:"edges,omitempty"`

	// Special diagram types
	Sequences []SequenceSpec `json:"sequences,omitempty" yaml:"sequences,omitempty"` // Sequence diagrams
	Tables    []TableSpec    `json:"tables,omitempty" yaml:"tables,omitempty"`       // SQL tables
}

// ContainerSpec defines a container (cluster/boundary) that holds nodes.
type ContainerSpec struct {
	ID    string `json:"id" yaml:"id"`
	Label string `json:"label,omitempty" yaml:"label,omitempty"`

	// Layout within container
	Direction   string `json:"direction,omitempty" yaml:"direction,omitempty"`
	GridColumns int    `json:"gridColumns,omitempty" yaml:"gridColumns,omitempty"`
	GridRows    int    `json:"gridRows,omitempty" yaml:"gridRows,omitempty"`
	// HorizontalGap and VerticalGap are the space between grid cells, in pixels.
	HorizontalGap int `json:"horizontalGap,omitempty" yaml:"horizontalGap,omitempty"`
	VerticalGap   int `json:"verticalGap,omitempty" yaml:"verticalGap,omitempty"`

	// Style
	Style *StyleSpec `json:"style,omitempty" yaml:"style,omitempty"`
	// Class names an entry of DiagramSpec.Classes.
	Class string `json:"class,omitempty" yaml:"class,omitempty"`

	// Children
	Nodes      []NodeSpec      `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Containers []ContainerSpec `json:"containers,omitempty" yaml:"containers,omitempty"` // Nested containers
	Edges      []EdgeSpec      `json:"edges,omitempty" yaml:"edges,omitempty"`           // Internal edges
}

// NodeSpec defines a single node in the diagram.
type NodeSpec struct {
	ID    string     `json:"id" yaml:"id"`
	Label string     `json:"label,omitempty" yaml:"label,omitempty"`
	Shape string     `json:"shape,omitempty" yaml:"shape,omitempty"` // rectangle, cylinder, circle, etc.
	Icon  string     `json:"icon,omitempty" yaml:"icon,omitempty"`   // Icon URL or name
	Style *StyleSpec `json:"style,omitempty" yaml:"style,omitempty"`
	// Class names an entry of DiagramSpec.Classes.
	Class string `json:"class,omitempty" yaml:"class,omitempty"`
}

// EdgeSpec defines a connection between nodes.
type EdgeSpec struct {
	// From and To name nodes. A dot separates the containers on the way to a
	// node, so "stage.item" is the node item in the container stage. Inside a
	// container, a name is relative to that container. A node ID that itself
	// contains a dot cannot be named by an edge.
	From  string `json:"from" yaml:"from"`
	To    string `json:"to" yaml:"to"`
	Label string `json:"label,omitempty" yaml:"label,omitempty"`

	// Arrow style
	SourceArrow string `json:"sourceArrow,omitempty" yaml:"sourceArrow,omitempty"` // none, triangle, diamond, etc.
	TargetArrow string `json:"targetArrow,omitempty" yaml:"targetArrow,omitempty"`

	Style *StyleSpec `json:"style,omitempty" yaml:"style,omitempty"`
	// Class names an entry of DiagramSpec.Classes.
	Class string `json:"class,omitempty" yaml:"class,omitempty"`
}

// StyleSpec defines visual styling properties.
// Pointer types are used to distinguish "not set" from "set to zero".
type StyleSpec struct {
	Fill         string   `json:"fill,omitempty" yaml:"fill,omitempty"`
	Stroke       string   `json:"stroke,omitempty" yaml:"stroke,omitempty"`
	StrokeWidth  *int     `json:"strokeWidth,omitempty" yaml:"strokeWidth,omitempty"`
	BorderRadius *int     `json:"borderRadius,omitempty" yaml:"borderRadius,omitempty"`
	FontSize     *int     `json:"fontSize,omitempty" yaml:"fontSize,omitempty"`
	Opacity      *float64 `json:"opacity,omitempty" yaml:"opacity,omitempty"`
	FontColor    string   `json:"fontColor,omitempty" yaml:"fontColor,omitempty"`
	Bold         *bool    `json:"bold,omitempty" yaml:"bold,omitempty"`
}

// BoolPtr returns a pointer to a bool value.
func BoolPtr(b bool) *bool { return &b }

// IntPtr returns a pointer to an int value. Useful for setting StyleSpec fields.
func IntPtr(i int) *int { return &i }

// Float64Ptr returns a pointer to a float64 value.
func Float64Ptr(f float64) *float64 { return &f }

// SequenceSpec defines a sequence diagram.
type SequenceSpec struct {
	ID     string        `json:"id" yaml:"id"`
	Label  string        `json:"label,omitempty" yaml:"label,omitempty"`
	Actors []ActorSpec   `json:"actors,omitempty" yaml:"actors,omitempty"`
	Steps  []MessageSpec `json:"steps" yaml:"steps"` // Messages in order
	Groups []GroupSpec   `json:"groups,omitempty" yaml:"groups,omitempty"`
}

// ActorSpec defines an actor in a sequence diagram.
type ActorSpec struct {
	ID    string `json:"id" yaml:"id"`
	Label string `json:"label,omitempty" yaml:"label,omitempty"`
	Shape string `json:"shape,omitempty" yaml:"shape,omitempty"` // person, rectangle, etc.
}

// MessageSpec defines a message between actors.
type MessageSpec struct {
	From  string `json:"from" yaml:"from"`
	To    string `json:"to" yaml:"to"`
	Label string `json:"label,omitempty" yaml:"label,omitempty"`
}

// GroupSpec defines a group/fragment in a sequence diagram.
type GroupSpec struct {
	ID       string        `json:"id" yaml:"id"`
	Label    string        `json:"label,omitempty" yaml:"label,omitempty"`
	Messages []MessageSpec `json:"messages" yaml:"messages"`
}

// TableSpec defines an SQL table diagram.
type TableSpec struct {
	ID      string       `json:"id" yaml:"id"`
	Label   string       `json:"label,omitempty" yaml:"label,omitempty"`
	Columns []ColumnSpec `json:"columns" yaml:"columns"`
}

// ColumnSpec defines a column in an SQL table.
type ColumnSpec struct {
	Name        string   `json:"name" yaml:"name"`
	Type        string   `json:"type" yaml:"type"`
	Constraints []string `json:"constraints,omitempty" yaml:"constraints,omitempty"` // PK, FK, UNQ, NOT NULL
}
