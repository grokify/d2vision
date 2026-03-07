package d2vision

import (
	"fmt"
	"strings"
)

// OutputFormat specifies the output format for diagram descriptions.
type OutputFormat string

const (
	FormatJSON    OutputFormat = "json"
	FormatText    OutputFormat = "text"
	FormatSummary OutputFormat = "summary"
)

// Describe generates a natural language description of the diagram.
func (d *Diagram) Describe() string {
	return d.DescribeDetailed()
}

// DescribeSummary generates a brief summary of the diagram.
func (d *Diagram) DescribeSummary() string {
	var sb strings.Builder

	nodeCount := len(d.Nodes)
	edgeCount := len(d.Edges)

	// Handle singular/plural
	nodeWord := "nodes"
	if nodeCount == 1 {
		nodeWord = "node"
	}
	edgeWord := "connections"
	if edgeCount == 1 {
		edgeWord = "connection"
	}

	fmt.Fprintf(&sb, "This D2 diagram contains %d %s and %d %s.",
		nodeCount, nodeWord, edgeCount, edgeWord)

	// Add container info if present
	containers := d.ContainerNodes()
	if len(containers) > 0 {
		containerWord := "containers"
		if len(containers) == 1 {
			containerWord = "container"
		}
		fmt.Fprintf(&sb, " It has %d %s.", len(containers), containerWord)
	}

	return sb.String()
}

// DescribeDetailed generates a detailed natural language description.
func (d *Diagram) DescribeDetailed() string {
	var sb strings.Builder

	// Header with counts
	sb.WriteString(d.DescribeSummary())
	sb.WriteString("\n")

	// Describe nodes
	if len(d.Nodes) > 0 {
		sb.WriteString("\nNodes:\n")
		for _, node := range d.Nodes {
			sb.WriteString(describeNode(node))
		}
	}

	// Describe edges
	if len(d.Edges) > 0 {
		sb.WriteString("\nConnections:\n")
		for _, edge := range d.Edges {
			sb.WriteString(describeEdge(edge, d))
		}
	}

	return sb.String()
}

func describeNode(node Node) string {
	label := node.DisplayLabel()

	// Format label with quotes if it contains spaces
	displayLabel := formatLabel(label)

	var desc string
	if node.Shape != ShapeRectangle && node.Shape != ShapeUnknown {
		desc = fmt.Sprintf("- %s is a %s", displayLabel, node.Shape.NaturalName())
	} else {
		desc = fmt.Sprintf("- %s is a rectangle", displayLabel)
	}

	if node.HasChildren() {
		childCount := len(node.Children)
		childWord := "children"
		if childCount == 1 {
			childWord = "child"
		}
		desc += fmt.Sprintf(" (container with %d %s)", childCount, childWord)
	}

	if node.Parent != "" {
		desc += fmt.Sprintf(" (inside %s)", formatLabel(node.Parent))
	}

	return desc + "\n"
}

func describeEdge(edge Edge, diagram *Diagram) string {
	sourceLabel := edge.Source
	targetLabel := edge.Target

	// Try to get display labels from nodes
	if sourceNode := diagram.NodeByID(edge.Source); sourceNode != nil {
		sourceLabel = sourceNode.DisplayLabel()
	}
	if targetNode := diagram.NodeByID(edge.Target); targetNode != nil {
		targetLabel = targetNode.DisplayLabel()
	}

	sourceDisplay := formatLabel(sourceLabel)
	targetDisplay := formatLabel(targetLabel)

	var desc string
	if edge.IsBidirectional() {
		desc = fmt.Sprintf("- %s connects bidirectionally with %s", sourceDisplay, targetDisplay)
	} else {
		desc = fmt.Sprintf("- %s connects to %s", sourceDisplay, targetDisplay)
	}

	if edge.Label != "" {
		desc += fmt.Sprintf(" (labeled %s)", formatLabel(edge.Label))
	}

	return desc + "\n"
}

// formatLabel formats a label for display, adding quotes if needed.
func formatLabel(label string) string {
	if strings.Contains(label, " ") || strings.Contains(label, ".") {
		return fmt.Sprintf("%q", label)
	}
	return label
}

// DescribeForLLM generates a description optimized for LLM consumption.
func (d *Diagram) DescribeForLLM() string {
	var sb strings.Builder

	sb.WriteString("# D2 Diagram Structure\n\n")

	if d.Title != "" {
		fmt.Fprintf(&sb, "Title: %s\n\n", d.Title)
	}

	nodeWord := "nodes"
	if len(d.Nodes) == 1 {
		nodeWord = "node"
	}
	edgeWord := "edges"
	if len(d.Edges) == 1 {
		edgeWord = "edge"
	}
	fmt.Fprintf(&sb, "Contains %d %s and %d %s.\n\n", len(d.Nodes), nodeWord, len(d.Edges), edgeWord)

	// Group nodes by parent
	rootNodes := d.RootNodes()
	containerNodes := d.ContainerNodes()

	if len(containerNodes) > 0 {
		sb.WriteString("## Containers\n")
		for _, c := range containerNodes {
			fmt.Fprintf(&sb, "- %s contains: %s\n", c.ID, strings.Join(c.Children, ", "))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Nodes\n")
	for _, node := range rootNodes {
		if !node.HasChildren() {
			fmt.Fprintf(&sb, "- %s (%s)\n", node.DisplayLabel(), node.Shape)
		}
	}
	sb.WriteString("\n")

	sb.WriteString("## Edges\n")
	for _, edge := range d.Edges {
		arrow := "->"
		if edge.IsBidirectional() {
			arrow = "<->"
		}
		if edge.Label != "" {
			fmt.Fprintf(&sb, "- %s %s %s: %q\n", edge.Source, arrow, edge.Target, edge.Label)
		} else {
			fmt.Fprintf(&sb, "- %s %s %s\n", edge.Source, arrow, edge.Target)
		}
	}

	return sb.String()
}
