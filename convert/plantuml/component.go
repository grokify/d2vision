package plantuml

import (
	"github.com/grokify/d2vision/generate"
)

// ComponentConverter converts PlantUML component diagrams to DiagramSpec.
type ComponentConverter struct{}

// Convert transforms a parsed PlantUML component document into a DiagramSpec.
func (c *ComponentConverter) Convert(doc *Document) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	// Track seen elements to avoid duplicates
	seenElements := make(map[string]bool)

	// Convert top-level components to nodes
	for _, comp := range doc.Components {
		if !seenElements[comp.ID] {
			spec.Nodes = append(spec.Nodes, generate.NodeSpec{
				ID:    comp.ID,
				Label: comp.Label,
			})
			seenElements[comp.ID] = true
		}
	}

	// Convert packages to containers
	for _, pkg := range doc.Packages {
		container := c.convertPackage(pkg, seenElements)
		spec.Containers = append(spec.Containers, container)
	}

	// Convert relations to edges
	for _, rel := range doc.Relations {
		// Ensure source and target exist
		if !seenElements[rel.From] {
			spec.Nodes = append(spec.Nodes, generate.NodeSpec{
				ID:    rel.From,
				Label: rel.From,
			})
			seenElements[rel.From] = true
		}
		if !seenElements[rel.To] {
			spec.Nodes = append(spec.Nodes, generate.NodeSpec{
				ID:    rel.To,
				Label: rel.To,
			})
			seenElements[rel.To] = true
		}

		edge := generate.EdgeSpec{
			From:  rel.From,
			To:    rel.To,
			Label: rel.Label,
		}

		// Set arrow styles based on relation type
		source, target := rel.Type.ToD2Arrows()
		if source != "none" {
			edge.SourceArrow = source
		}
		if target != "none" {
			edge.TargetArrow = target
		}

		spec.Edges = append(spec.Edges, edge)
	}

	return spec
}

func (c *ComponentConverter) convertPackage(pkg *Package, seenElements map[string]bool) generate.ContainerSpec {
	container := generate.ContainerSpec{
		ID:    pkg.ID,
		Label: pkg.Label,
	}

	seenElements[pkg.ID] = true

	// Convert components within package
	for _, comp := range pkg.Components {
		if !seenElements[comp.ID] {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    comp.ID,
				Label: comp.Label,
			})
			seenElements[comp.ID] = true
		}
	}

	// Convert classes within package (for mixed diagrams)
	for _, class := range pkg.Classes {
		if !seenElements[class.ID] {
			container.Nodes = append(container.Nodes, generate.NodeSpec{
				ID:    class.ID,
				Label: class.Label,
				Shape: "class",
			})
			seenElements[class.ID] = true
		}
	}

	// Convert nested packages
	for _, nested := range pkg.Packages {
		nestedContainer := c.convertPackage(nested, seenElements)
		container.Containers = append(container.Containers, nestedContainer)
	}

	return container
}
