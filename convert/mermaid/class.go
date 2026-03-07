package mermaid

import (
	"github.com/grokify/d2vision/generate"
)

// ClassConverter converts Mermaid class diagrams to DiagramSpec.
type ClassConverter struct{}

// Convert transforms a parsed Mermaid class document into a DiagramSpec.
func (c *ClassConverter) Convert(doc *Document) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	// Convert classes to containers (D2 class shapes)
	for _, class := range doc.Classes {
		container := generate.ContainerSpec{
			ID:    class.ID,
			Label: class.Label,
		}

		// Add attributes as nodes
		for i, attr := range class.Attributes {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    class.ID + "_attr_" + string(rune('a'+i)),
				Label: attr,
			})
		}

		// Add methods as nodes
		for i, method := range class.Methods {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    class.ID + "_method_" + string(rune('a'+i)),
				Label: method,
			})
		}

		spec.Containers = append(spec.Containers, container)
	}

	// Convert relationships to edges
	for _, edge := range doc.Edges {
		spec.Edges = append(spec.Edges, generate.EdgeSpec{
			From:  edge.From,
			To:    edge.To,
			Label: edge.Label,
		})
	}

	return spec
}
