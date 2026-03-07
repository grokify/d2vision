package mermaid

import (
	"github.com/grokify/d2vision/generate"
)

// FlowchartConverter converts Mermaid flowcharts to DiagramSpec.
type FlowchartConverter struct{}

// Convert transforms a parsed Mermaid flowchart document into a DiagramSpec.
func (c *FlowchartConverter) Convert(doc *Document) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	// Set direction
	if doc.Direction != "" {
		spec.Direction = doc.Direction.ToD2Direction()
	}

	// Track seen nodes to avoid duplicates
	seenNodes := make(map[string]bool)

	// Convert top-level nodes
	for _, node := range doc.Nodes {
		if !seenNodes[node.ID] {
			spec.Nodes = append(spec.Nodes, convertNode(node))
			seenNodes[node.ID] = true
		}
	}

	// Convert subgraphs to containers
	for _, sg := range doc.Subgraphs {
		container := c.convertSubgraph(sg, seenNodes)
		spec.Containers = append(spec.Containers, container)
	}

	// Convert top-level edges
	for _, edge := range doc.Edges {
		// Ensure source and target nodes exist
		if !seenNodes[edge.From] {
			spec.Nodes = append(spec.Nodes, generate.NodeSpec{
				ID:    edge.From,
				Label: edge.From,
			})
			seenNodes[edge.From] = true
		}
		if !seenNodes[edge.To] {
			spec.Nodes = append(spec.Nodes, generate.NodeSpec{
				ID:    edge.To,
				Label: edge.To,
			})
			seenNodes[edge.To] = true
		}

		spec.Edges = append(spec.Edges, convertEdge(edge))
	}

	return spec
}

func (c *FlowchartConverter) convertSubgraph(sg *Subgraph, seenNodes map[string]bool) generate.ContainerSpec {
	container := generate.ContainerSpec{
		ID:    sg.ID,
		Label: sg.Label,
	}

	// Set direction if specified
	if sg.Direction != "" {
		container.Direction = sg.Direction.ToD2Direction()
	}

	// Convert nodes
	for _, node := range sg.Nodes {
		if !seenNodes[node.ID] {
			container.Nodes = append(container.Nodes, convertNode(node))
			seenNodes[node.ID] = true
		}
	}

	// Convert nested subgraphs
	for _, nested := range sg.Subgraphs {
		nestedContainer := c.convertSubgraph(nested, seenNodes)
		container.Containers = append(container.Containers, nestedContainer)
	}

	// Convert edges within subgraph
	for _, edge := range sg.Edges {
		// Ensure source and target nodes exist
		if !seenNodes[edge.From] {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    edge.From,
				Label: edge.From,
			})
			seenNodes[edge.From] = true
		}
		if !seenNodes[edge.To] {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    edge.To,
				Label: edge.To,
			})
			seenNodes[edge.To] = true
		}

		container.Edges = append(container.Edges, convertEdge(edge))
	}

	return container
}

func convertNode(node *Node) generate.NodeSpec {
	spec := generate.NodeSpec{
		ID:    node.ID,
		Label: node.Label,
	}

	// Set shape if not default rectangle
	d2Shape := node.Shape.ToD2Shape()
	if d2Shape != "" && d2Shape != "rectangle" {
		spec.Shape = d2Shape
	}

	// Handle rounded rectangle with border-radius
	if node.Shape == ShapeRoundedRect {
		spec.Style.BorderRadius = generate.IntPtr(8)
	}

	return spec
}

func convertEdge(edge *Edge) generate.EdgeSpec {
	spec := generate.EdgeSpec{
		From:  edge.From,
		To:    edge.To,
		Label: edge.Label,
	}

	// Handle dashed style
	if edge.Style.Dashed {
		spec.Style.StrokeWidth = generate.IntPtr(2)
		// Note: D2 uses stroke-dash in style, which requires string manipulation
		// For now, we'll mark it and let the generator handle it
	}

	// Handle arrow types
	switch edge.Style.ArrowType {
	case ArrowBidir:
		spec.SourceArrow = "triangle"
		spec.TargetArrow = "triangle"
	case ArrowNone:
		spec.TargetArrow = "none"
	case ArrowCircle:
		spec.TargetArrow = "circle"
	case ArrowCross:
		spec.TargetArrow = "cf-one" // D2's closest equivalent
	}

	return spec
}
