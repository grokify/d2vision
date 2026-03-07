// Package mermaid provides parsing and conversion of Mermaid diagrams to D2.
package mermaid

// DiagramType represents the type of Mermaid diagram.
type DiagramType string

const (
	DiagramFlowchart DiagramType = "flowchart"
	DiagramSequence  DiagramType = "sequence"
	DiagramClass     DiagramType = "class"
	DiagramState     DiagramType = "state"
	DiagramER        DiagramType = "er"
	DiagramGantt     DiagramType = "gantt"
	DiagramPie       DiagramType = "pie"
	DiagramGitGraph  DiagramType = "gitGraph"
	DiagramUnknown   DiagramType = "unknown"
)

// Document represents a parsed Mermaid document.
type Document struct {
	Type      DiagramType
	Direction Direction
	Nodes     []*Node
	Edges     []*Edge
	Subgraphs []*Subgraph

	// Sequence diagram specific
	Actors   []*Actor
	Messages []*Message
	Groups   []*MessageGroup

	// Class diagram specific
	Classes []*Class

	// Raw lines for diagnostics
	Lines []string
}

// Direction represents the layout direction.
type Direction string

const (
	DirectionTB Direction = "TB" // Top to Bottom
	DirectionTD Direction = "TD" // Top to Down (same as TB)
	DirectionBT Direction = "BT" // Bottom to Top
	DirectionRL Direction = "RL" // Right to Left
	DirectionLR Direction = "LR" // Left to Right
)

// ToD2Direction converts Mermaid direction to D2 direction.
func (d Direction) ToD2Direction() string {
	switch d {
	case DirectionLR:
		return "right"
	case DirectionRL:
		return "left"
	case DirectionTB, DirectionTD:
		return "down"
	case DirectionBT:
		return "up"
	default:
		return "down"
	}
}

// Node represents a Mermaid node.
type Node struct {
	ID    string
	Label string
	Shape NodeShape
	Line  int // Source line number
}

// NodeShape represents the shape of a node.
type NodeShape string

const (
	ShapeRectangle    NodeShape = "rectangle"
	ShapeRoundedRect  NodeShape = "rounded"
	ShapeCircle       NodeShape = "circle"
	ShapeDiamond      NodeShape = "diamond"
	ShapeCylinder     NodeShape = "cylinder"
	ShapeHexagon      NodeShape = "hexagon"
	ShapeParallelogram NodeShape = "parallelogram"
	ShapeTrapezoid    NodeShape = "trapezoid"
	ShapeStadium      NodeShape = "stadium"
	ShapeSubroutine   NodeShape = "subroutine"
	ShapeAsymmetric   NodeShape = "asymmetric"
	ShapeDouble       NodeShape = "double"
)

// ToD2Shape converts Mermaid shape to D2 shape.
func (s NodeShape) ToD2Shape() string {
	switch s {
	case ShapeRectangle:
		return "rectangle"
	case ShapeRoundedRect:
		return "rectangle" // D2 uses border-radius for rounding
	case ShapeCircle:
		return "circle"
	case ShapeDiamond:
		return "diamond"
	case ShapeCylinder:
		return "cylinder"
	case ShapeHexagon:
		return "hexagon"
	case ShapeParallelogram:
		return "parallelogram"
	case ShapeStadium:
		return "oval"
	case ShapeSubroutine:
		return "rectangle"
	case ShapeAsymmetric:
		return "rectangle"
	case ShapeDouble:
		return "rectangle"
	default:
		return ""
	}
}

// Edge represents a connection between nodes.
type Edge struct {
	From      string
	To        string
	Label     string
	Style     EdgeStyle
	Line      int // Source line number
}

// EdgeStyle represents the visual style of an edge.
type EdgeStyle struct {
	Dashed    bool
	Thick     bool
	Invisible bool
	ArrowType ArrowType
}

// ArrowType represents the type of arrow on an edge.
type ArrowType string

const (
	ArrowNormal  ArrowType = "normal"
	ArrowOpen    ArrowType = "open"
	ArrowCircle  ArrowType = "circle"
	ArrowCross   ArrowType = "cross"
	ArrowBidir   ArrowType = "bidir"
	ArrowNone    ArrowType = "none"
)

// Subgraph represents a container in Mermaid.
type Subgraph struct {
	ID        string
	Label     string
	Direction Direction
	Nodes     []*Node
	Edges     []*Edge
	Subgraphs []*Subgraph
	Line      int // Source line number
}

// Actor represents a participant in a sequence diagram.
type Actor struct {
	ID    string
	Label string
	Shape string // participant, actor
	Line  int
}

// Message represents a message in a sequence diagram.
type Message struct {
	From      string
	To        string
	Label     string
	Style     MessageStyle
	Line      int
}

// MessageStyle represents the style of a sequence message.
type MessageStyle string

const (
	MessageSolid  MessageStyle = "solid"
	MessageDashed MessageStyle = "dashed"
	MessageAsync  MessageStyle = "async"
)

// MessageGroup represents a group of messages (alt, opt, loop, etc.).
type MessageGroup struct {
	Type     GroupType
	Label    string
	Messages []*Message
	Else     []*Message // For alt/else groups
	Line     int
}

// GroupType represents the type of message group.
type GroupType string

const (
	GroupAlt      GroupType = "alt"
	GroupOpt      GroupType = "opt"
	GroupLoop     GroupType = "loop"
	GroupPar      GroupType = "par"
	GroupCritical GroupType = "critical"
	GroupBreak    GroupType = "break"
)

// Class represents a class in a class diagram.
type Class struct {
	ID         string
	Label      string
	Attributes []string
	Methods    []string
	Line       int
}
