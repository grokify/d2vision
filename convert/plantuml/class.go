package plantuml

import (
	"github.com/grokify/d2vision/generate"
)

// ClassConverter converts PlantUML class diagrams to DiagramSpec.
type ClassConverter struct{}

// Convert transforms a parsed PlantUML class document into a DiagramSpec.
func (c *ClassConverter) Convert(doc *Document) *generate.DiagramSpec {
	spec := &generate.DiagramSpec{}

	// Track seen elements
	seenElements := make(map[string]bool)

	// Convert top-level classes
	for _, class := range doc.Classes {
		container := c.convertClass(class)
		spec.Containers = append(spec.Containers, container)
		seenElements[class.ID] = true
	}

	// Convert packages
	for _, pkg := range doc.Packages {
		container := c.convertPackageWithClasses(pkg, seenElements)
		spec.Containers = append(spec.Containers, container)
	}

	// Convert relations to edges
	for _, rel := range doc.Relations {
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

		// Dashed for dependency and realization
		if rel.Type == RelationDependency || rel.Type == RelationRealization {
			edge.Style.StrokeWidth = generate.IntPtr(2)
		}

		spec.Edges = append(spec.Edges, edge)
	}

	return spec
}

func (c *ClassConverter) convertClass(class *Class) generate.ContainerSpec {
	container := generate.ContainerSpec{
		ID:    class.ID,
		Label: class.Label,
	}

	// Add stereotype to label if present
	if class.Stereotype != "" {
		container.Label = "«" + class.Stereotype + "»\n" + class.Label
	} else if class.Interface {
		container.Label = "«interface»\n" + class.Label
	} else if class.Abstract {
		container.Label = "«abstract»\n" + class.Label
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

	return container
}

func (c *ClassConverter) convertPackageWithClasses(pkg *Package, seenElements map[string]bool) generate.ContainerSpec {
	container := generate.ContainerSpec{
		ID:    pkg.ID,
		Label: pkg.Label,
	}

	seenElements[pkg.ID] = true

	// Convert classes within package
	for _, class := range pkg.Classes {
		if !seenElements[class.ID] {
			classContainer := c.convertClass(class)
			container.Containers = append(container.Containers, classContainer)
			seenElements[class.ID] = true
		}
	}

	// Convert nested packages
	for _, nested := range pkg.Packages {
		nestedContainer := c.convertPackageWithClasses(nested, seenElements)
		container.Containers = append(container.Containers, nestedContainer)
	}

	return container
}
